package repository

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// CreateUser inserts a new local (email/password) or provider-only user row
// and returns the generated auto-increment id.
//
// RegisterService already pre-checks email/member_id uniqueness, but
// that check-then-insert is only a friendly fast path - it can't prevent two
// concurrent registrations for the same email/badge from both passing the
// pre-check before either commits. The UNIQUE constraints on the users
// table are the real enforcement; a raced request lands here as a MySQL
// duplicate-entry error, which is translated to the same sentinel the
// pre-check uses so both paths converge on one clean, consistent response.
func (r *AuthRepository) CreateUser(ctx *gin.Context, u *models.User) (int64, error) {
	if u.UUID == "" {
		u.UUID = uuid.NewString()
	}
	if u.Status == "" {
		u.Status = models.UserStatusActive
	}
	if u.MemberType == "" {
		u.MemberType = models.MemberTypeRegular
	}

	// email is nullable (Young Saver accounts may not have one) - an empty
	// Go string must become SQL NULL rather than "", otherwise a second
	// email-less registration would collide with the first on
	// uq_users_email (NULL is unique-safe against itself, "" is not).
	email := sql.NullString{String: u.Email, Valid: u.Email != ""}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	res, err := tx.ExecContext(ctx,
		`INSERT INTO users (
			uuid, email, username,
			first_name, middle_name, last_name, suffix, member_id, date_of_birth, gender, mobile_number, location,
			display_name, avatar_url, status, member_type, guardian_full_name, guardian_mobile_number, guardian_consent_at
		 ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.UUID, email, u.Username,
		u.FirstName, u.MiddleName, u.LastName, u.Suffix, u.MemberID, u.DateOfBirth, u.Gender, u.MobileNumber, u.Location,
		u.DisplayName, u.AvatarUrl, u.Status, u.MemberType, u.GuardianFullName, u.GuardianMobileNumber, u.GuardianConsentAt,
	)
	if err != nil {
		return 0, translateDuplicateUserError(err)
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Every user row gets a paired credential row, even social/SSO-only
	// accounts (password_hash NULL) - this lets every read join
	// user_account_cred with a plain INNER JOIN instead of a LEFT JOIN.
	var lastPasswordChangeAt sql.NullTime
	if u.PasswordHash != nil {
		lastPasswordChangeAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO user_account_cred (user_id, member_id, password_hash, last_password_change_at) VALUES (?, ?, ?, ?)`,
		userID, u.MemberID, u.PasswordHash, lastPasswordChangeAt,
	); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}

// translateDuplicateUserError maps a MySQL ER_DUP_ENTRY (1062) on one of the
// users table's unique keys to the matching sentinel error, so a losing
// request in a registration race gets the same clean response as the
// pre-check path instead of a raw driver error string.
func translateDuplicateUserError(err error) error {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return err
	}

	switch {
	case strings.Contains(mysqlErr.Message, "uq_users_email"):
		return cc.ErrEmailAlreadyRegistered
	case strings.Contains(mysqlErr.Message, "uq_users_member_id"):
		return cc.ErrMemberIDAlreadyRegistered
	default:
		return err
	}
}

const userSelectColumns = `u.id, u.uuid, u.email, u.username, c.password_hash,
	u.first_name, u.middle_name, u.last_name, u.suffix, u.member_id, u.date_of_birth, u.gender, u.mobile_number, u.location,
	u.display_name, u.avatar_url, u.status, u.member_type, u.guardian_full_name, u.guardian_mobile_number, u.guardian_consent_at,
	u.email_verified_at, c.last_login_at, c.last_password_change_at, u.created_at, u.updated_at`

const userFromClause = `users u JOIN user_account_cred c ON c.user_id = u.id`

func (r *AuthRepository) FindUserByEmail(ctx *gin.Context, email string) (*models.User, error) {
	row := r.DB.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+`
		 FROM `+userFromClause+` WHERE u.email = ? AND u.deleted_at IS NULL`,
		email,
	)

	return scanUser(row)
}

func (r *AuthRepository) FindUserByMemberID(ctx *gin.Context, memberID string) (*models.User, error) {
	row := r.DB.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+`
		 FROM `+userFromClause+` WHERE u.member_id = ? AND u.deleted_at IS NULL`,
		memberID,
	)

	return scanUser(row)
}

func (r *AuthRepository) FindUserByID(ctx *gin.Context, id int64) (*models.User, error) {
	row := r.DB.QueryRowContext(ctx,
		`SELECT `+userSelectColumns+`
		 FROM `+userFromClause+` WHERE u.id = ? AND u.deleted_at IS NULL`,
		id,
	)

	return scanUser(row)
}

// UpdateUserProfile overwrites the caller's own name/email/mobile fields on
// the users table. middleName is nullable (same as at registration); the
// other fields are required by UpdateProfileRequest's validation, so they're
// plain strings here. A raced update to an email another account already
// holds surfaces as ER_DUP_ENTRY on uq_users_email, translated the same way
// CreateUser's duplicate check is.
func (r *AuthRepository) UpdateUserProfile(ctx *gin.Context, userID int64, firstName, lastName string, middleName *string, email, mobileNumber string) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE users SET first_name = ?, middle_name = ?, last_name = ?, email = ?, mobile_number = ?
		 WHERE id = ? AND deleted_at IS NULL`,
		firstName, middleName, lastName, email, mobileNumber, userID,
	)
	if err != nil {
		return translateDuplicateUserError(err)
	}
	return nil
}

func (r *AuthRepository) TouchLastLogin(ctx *gin.Context, userID int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE user_account_cred SET last_login_at = NOW() WHERE user_id = ?`, userID)
	return err
}

// UpdateUserPassword overwrites the account's password hash - used to issue
// a temporary password once a forgot-password OTP challenge is verified.
func (r *AuthRepository) UpdateUserPassword(ctx *gin.Context, userID int64, passwordHash string) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE user_account_cred SET password_hash = ?, last_password_change_at = NOW() WHERE user_id = ?`,
		passwordHash, userID)
	return err
}

func scanUser(row *sql.Row) (*models.User, error) {
	var u models.User
	var email sql.NullString

	err := row.Scan(
		&u.ID, &u.UUID, &email, &u.Username, &u.PasswordHash,
		&u.FirstName, &u.MiddleName, &u.LastName, &u.Suffix, &u.MemberID, &u.DateOfBirth, &u.Gender, &u.MobileNumber, &u.Location,
		&u.DisplayName, &u.AvatarUrl,
		&u.Status, &u.MemberType, &u.GuardianFullName, &u.GuardianMobileNumber, &u.GuardianConsentAt,
		&u.EmailVerifiedAt, &u.LastLoginAt, &u.LastPasswordChangeAt, &u.CreatedAt, &u.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	u.Email = email.String
	return &u, nil
}
