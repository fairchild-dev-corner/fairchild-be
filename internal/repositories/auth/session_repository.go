package repository

import (
	"database/sql"
	"errors"
	models "fairchild_be/internal/models/auth"

	"github.com/gin-gonic/gin"
)

// CreateSession persists a new refresh-token session. Only the hash of the
// refresh token is stored (see utils.MustStringSHA256) - never the raw value.
func (r *AuthRepository) CreateSession(ctx *gin.Context, s *models.AuthSession) (int64, error) {
	res, err := r.DB.ExecContext(ctx,
		`INSERT INTO auth_sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		s.UserID, s.RefreshTokenHash, s.UserAgent, s.IPAddress, s.ExpiresAt,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (r *AuthRepository) FindActiveSessionByTokenHash(ctx *gin.Context, tokenHash string) (*models.AuthSession, error) {
	var s models.AuthSession

	err := r.DB.QueryRowContext(ctx,
		`SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, created_at
		 FROM auth_sessions
		 WHERE refresh_token_hash = ? AND revoked_at IS NULL AND expires_at > NOW()`,
		tokenHash,
	).Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.RevokedAt, &s.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *AuthRepository) RevokeSession(ctx *gin.Context, tokenHash string) error {
	_, err := r.DB.ExecContext(ctx,
		`UPDATE auth_sessions SET revoked_at = NOW() WHERE refresh_token_hash = ? AND revoked_at IS NULL`,
		tokenHash,
	)
	return err
}

// ListActiveSessionsByUserID returns the user's non-revoked, unexpired
// sessions, most recent first - powers the device/"last login" list on the
// member profile page.
func (r *AuthRepository) ListActiveSessionsByUserID(ctx *gin.Context, userID int64) ([]models.AuthSession, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, created_at
		 FROM auth_sessions
		 WHERE user_id = ? AND revoked_at IS NULL AND expires_at > NOW()
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]models.AuthSession, 0)
	for rows.Next() {
		var s models.AuthSession
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress,
			&s.ExpiresAt, &s.RevokedAt, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}
