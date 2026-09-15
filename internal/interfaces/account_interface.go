package interfaces

import "github.com/gin-gonic/gin"

// AccountRepositoryInterface resolves an authenticated user's own legacy
// ledger account (ClientID + branch) - the single shared lookup every
// per-member feature used to duplicate. See
// middlewares.RequireMemberAuth, which calls this exactly once per
// request and caches the result in the gin.Context, rather than each
// feature service calling it independently.
type AccountRepositoryInterface interface {
	FindClientAccountByUserID(ctx *gin.Context, userID int64) (clientID, branchID string, err error)
}
