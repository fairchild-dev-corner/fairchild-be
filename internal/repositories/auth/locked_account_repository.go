package repository

import (
	"database/sql"
	"errors"

	models "fairchild_be/internal/models/auth"

	"github.com/gin-gonic/gin"
)

// lockedMemberAccountTTLDays is how long a locked_member_accounts row is
// considered current - see FindActiveLockedMemberAccount.
const lockedMemberAccountTTLDays = 10

// FindActiveLockedMemberAccount returns the most recent locked_member_accounts
// row for memberID that is still marked locked and hasn't expired, or
// nil if there isn't one. Used to avoid creating a duplicate row (and
// re-sending the notification email) on every subsequent request while a
// lockout episode is still in effect.
func (r *AuthRepository) FindActiveLockedMemberAccount(ctx *gin.Context, memberID string) (*models.LockedMemberAccount, error) {
	var l models.LockedMemberAccount

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, user_id, member_id, ip_address, attempt_count, reason, is_locked, locked_at, expires_at, notified_at, unlocked_at, created_at
		 FROM locked_member_accounts
		 WHERE member_id = ? AND is_locked = TRUE AND expires_at > NOW()
		 ORDER BY id DESC LIMIT 1`,
		memberID,
	).Scan(
		&l.ID, &l.UserID, &l.MemberID, &l.IPAddress, &l.AttemptCount, &l.Reason, &l.IsLocked,
		&l.LockedAt, &l.ExpiresAt, &l.NotifiedAt, &l.UnlockedAt, &l.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// CreateLockedMemberAccount persists a new lockout record, expiring
// lockedMemberAccountTTLDays from now, and returns its generated id.
func (r *AuthRepository) CreateLockedMemberAccount(ctx *gin.Context, l *models.LockedMemberAccount) (int64, error) {
	res, err := r.DB.ExecContext(ctx,
		`INSERT INTO locked_member_accounts (user_id, member_id, ip_address, attempt_count, reason, expires_at)
		 VALUES (?, ?, ?, ?, ?, NOW() + INTERVAL ? DAY)`,
		l.UserID, l.MemberID, l.IPAddress, l.AttemptCount, l.Reason, lockedMemberAccountTTLDays,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateLockedMemberAccountAttempts bumps the attempt count on an
// already-active lock row instead of inserting a duplicate for a repeat
// trip of the same lockout episode.
func (r *AuthRepository) UpdateLockedMemberAccountAttempts(ctx *gin.Context, id int64, attemptCount int) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE locked_member_accounts SET attempt_count = ? WHERE id = ?`, attemptCount, id)
	return err
}

// MarkLockedMemberAccountNotified records that the lockout email was
// actually sent, so support can tell whether the member was informed.
func (r *AuthRepository) MarkLockedMemberAccountNotified(ctx *gin.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE locked_member_accounts SET notified_at = NOW() WHERE id = ?`, id)
	return err
}

// UnlockMemberAccount clears any active lock row for memberID
// (is_locked = FALSE, unlocked_at stamped) without deleting it, keeping the
// audit trail intact - development-only, see AuthService.UnlockAccountService.
func (r *AuthRepository) UnlockMemberAccount(ctx *gin.Context, memberID string) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE locked_member_accounts SET is_locked = FALSE, unlocked_at = NOW()
		 WHERE member_id = ? AND is_locked = TRUE`,
		memberID,
	)
	return err
}
