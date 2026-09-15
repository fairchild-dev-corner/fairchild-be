package repository

import (
	"database/sql"
	"errors"
	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"
	"fmt"

	"github.com/gin-gonic/gin"
)

func (r *AuthRepository) FindProviderByCode(ctx *gin.Context, code string) (*models.AuthProvider, error) {
	var p models.AuthProvider

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, code, name, is_enabled FROM auth_providers WHERE code = ?`, code,
	).Scan(&p.ID, &p.Code, &p.Name, &p.IsEnabled)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *AuthRepository) FindSocialAccount(ctx *gin.Context, providerID int8, providerUserID string) (*models.SocialAccount, error) {
	var s models.SocialAccount

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, user_id, provider_id, provider_user_id, provider_email,
		        access_token, refresh_token, token_expires_at, created_at, updated_at
		 FROM user_social_accounts WHERE provider_id = ? AND provider_user_id = ?`,
		providerID, providerUserID,
	).Scan(
		&s.ID, &s.UserID, &s.ProviderID, &s.ProviderUserID, &s.ProviderEmail,
		&s.AccessToken, &s.RefreshToken, &s.TokenExpiresAt, &s.CreatedAt, &s.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// FindLinkedMemberFromSocialProfile implements the "find, never create"
// identity flow for social login, run inside a single transaction so a
// concurrent login with the same provider identity can't race:
//
//  1. Look up user_social_accounts by (provider_id, provider_user_id).
//     - found  -> refresh the stored tokens, return the linked user.
//  2. Not found -> look up an existing user by email (lets a member who
//     registered with a password link a social account to that same
//     account instead of getting a duplicate).
//     - found     -> link this provider identity to that user.
//     - not found -> cc.ErrEmailNotAssociated. Social/SSO login only
//     authenticates existing members - it deliberately never provisions a
//     new one, since that would hand out an active member account to
//     anyone with an email address, skipping OTP/mobile verification and
//     staff review entirely.
func (r *AuthRepository) FindLinkedMemberFromSocialProfile(
	ctx *gin.Context, provider *models.AuthProvider, profile *models.SocialProfile,
) (*models.User, error) {

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op once committed

	var userID int64

	row := tx.QueryRowContext(ctx,
		`SELECT user_id FROM user_social_accounts WHERE provider_id = ? AND provider_user_id = ? FOR UPDATE`,
		provider.ID, profile.ProviderUserID,
	)

	switch err := row.Scan(&userID); {
	case err == nil:
		// existing social identity - refresh its stored tokens.
		if _, err := tx.ExecContext(ctx,
			`UPDATE user_social_accounts
			 SET provider_email = ?, access_token = ?, refresh_token = ?, token_expires_at = ?
			 WHERE provider_id = ? AND provider_user_id = ?`,
			profile.Email, nullIfEmpty(profile.AccessToken), nullIfEmpty(profile.RefreshToken),
			profile.ExpiresAt, provider.ID, profile.ProviderUserID,
		); err != nil {
			return nil, err
		}

	case errors.Is(err, sql.ErrNoRows):
		userID, err = r.linkExistingUserTx(ctx, tx, provider, profile)
		if err != nil {
			return nil, err
		}

	default:
		return nil, err
	}

	user, err := r.findUserByIDTx(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

// linkExistingUserTx links a verified social identity to an existing member
// found by email. It never creates a users row - see
// FindLinkedMemberFromSocialProfile's doc comment for why.
func (r *AuthRepository) linkExistingUserTx(
	ctx *gin.Context, tx *sql.Tx, provider *models.AuthProvider, profile *models.SocialProfile,
) (int64, error) {

	var userID int64

	err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? AND deleted_at IS NULL`, profile.Email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, cc.ErrEmailNotAssociated
	}
	if err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO user_social_accounts
		   (user_id, provider_id, provider_user_id, provider_email, access_token, refresh_token, token_expires_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, provider.ID, profile.ProviderUserID, profile.Email,
		nullIfEmpty(profile.AccessToken), nullIfEmpty(profile.RefreshToken), profile.ExpiresAt,
	); err != nil {
		return 0, fmt.Errorf("failed to link social account: %w", err)
	}

	return userID, nil
}

func (r *AuthRepository) findUserByIDTx(ctx *gin.Context, tx *sql.Tx, id int64) (*models.User, error) {
	var u models.User
	var email sql.NullString

	err := tx.QueryRowContext(ctx,
		`SELECT u.id, u.uuid, u.email, u.username, c.password_hash, u.display_name, u.avatar_url,
		        u.status, u.member_type, u.email_verified_at, c.last_login_at, u.created_at, u.updated_at
		 FROM users u JOIN user_account_cred c ON c.user_id = u.id WHERE u.id = ?`, id,
	).Scan(
		&u.ID, &u.UUID, &email, &u.Username, &u.PasswordHash, &u.DisplayName, &u.AvatarUrl,
		&u.Status, &u.MemberType, &u.EmailVerifiedAt, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	u.Email = email.String
	return &u, nil
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
