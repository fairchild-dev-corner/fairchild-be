package middlewares

import (
	"net/http"
	"strings"

	interfaces "fairchild_be/internal/interfaces"
	auth_service "fairchild_be/internal/services/auth"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
)

const (
	memberUserIDKey   = "member_user_id"
	memberClientIDKey = "member_client_id"
	memberBranchIDKey = "member_branch_id"
)

// RequireMemberAuth validates the "Authorization: Bearer <token>" header and
// resolves the caller's own legacy ledger account (ClientID + branch)
// exactly once, stashing both in the request context for every downstream
// per-member feature (loans, ar, amortization, soa, transactions, health,
// activity, balance) to read via UserIDFromContext/ClientAccountFromContext.
// Each of those features previously ran its own identical
// FindClientAccountByUserID query independently - up to 8 redundant round
// trips on a single dashboard page load - all now collapsed into this one
// lookup per request.
//
// A user with no linked ledger account is NOT rejected here - clientID/
// branchID are simply set to "" and downstream services already treat that
// as "nothing to show" rather than an error, exactly as before this change.
func RequireMemberAuth(tokens *auth_service.TokenService, accountRepo interfaces.AccountRepositoryInterface) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		const prefix = "Bearer "

		if header == "" || !strings.HasPrefix(header, prefix) {
			loggers.GetCommonError(ctx, "missing bearer token", http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		claims, err := tokens.ParseAccessToken(strings.TrimPrefix(header, prefix))
		if err != nil {
			loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		clientID, branchID, err := accountRepo.FindClientAccountByUserID(ctx, claims.UserID)
		if err != nil {
			loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
			ctx.Abort()
			return
		}

		ctx.Set(memberUserIDKey, claims.UserID)
		ctx.Set(memberClientIDKey, clientID)
		ctx.Set(memberBranchIDKey, branchID)
		ctx.Next()
	}
}

// UserIDFromContext reads the authenticated caller's user id, set by
// RequireMemberAuth.
func UserIDFromContext(ctx *gin.Context) (int64, bool) {
	v, exists := ctx.Get(memberUserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// ClientAccountFromContext reads the authenticated caller's own resolved
// ledger account, set by RequireMemberAuth. ok is false only if the
// middleware never ran; clientID == "" with ok == true is a valid state (an
// authenticated user with no linked legacy account) - callers should treat
// that as "nothing to show", not as a missing-context error.
func ClientAccountFromContext(ctx *gin.Context) (clientID, branchID string, ok bool) {
	c, exists := ctx.Get(memberClientIDKey)
	if !exists {
		return "", "", false
	}
	b, _ := ctx.Get(memberBranchIDKey)
	clientID, _ = c.(string)
	branchID, _ = b.(string)
	return clientID, branchID, true
}
