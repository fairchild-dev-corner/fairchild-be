package repository

import (
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"
)

/**
* AccountRepository
* @Description: Wraps the pooled MySQL connection for the one lookup every
* per-member feature (loans, ar, amortization, soa, transactions, health,
* activity, balance) used to duplicate - resolving an authenticated user's
* own legacy ledger account (ClientID + branch). Called exactly once per
* request by middlewares.RequireMemberAuth instead of once per feature.
**/
type AccountRepository struct {
	DB *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{DB: db}
}

// FindClientAccountByUserID resolves the caller's own ledger account
// (ClientID + branch) from users.user_client_id - callers are never trusted
// to supply their client_id/branch_id directly, since that would let any
// authenticated user read another member's data. A user with no linked
// legacy account returns "", "", nil - not an error - so callers can treat
// it as "no account" rather than a failure.
func (r *AccountRepository) FindClientAccountByUserID(ctx *gin.Context, userID int64) (clientID, branchID string, err error) {
	err = r.DB.QueryRowContext(ctx, `
		SELECT c.ClientID, c.ClientIDBrCode
		FROM users u
		INNER JOIN client c ON c.ClientID = u.user_client_id
		WHERE u.id = ? AND u.user_client_id IS NOT NULL`,
		userID,
	).Scan(&clientID, &branchID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}

	return clientID, branchID, nil
}
