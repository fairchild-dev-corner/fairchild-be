package handlers

import (
	"net/http"

	cc "fairchild_be/internal/constants"
	models "fairchild_be/internal/models/auth"
	"fairchild_be/internal/utils"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
)

// RegisterDevRoutes wires debug-only endpoints that must never be reachable
// outside local development. cmd/api/api.go only calls this when
// hostConf.BuildEnv == "dev" - a stage/prod binary never registers these
// routes at all, rather than gating them with a runtime check. Deliberately
// not nested under the rate-limited /auth group: its entire purpose is
// unblocking a developer who just tripped the lockout while testing.
func (h *AuthHandler) RegisterDevRoutes(router *gin.RouterGroup) {
	dev := router.Group("/auth/dev")
	dev.POST("/unlock-account", h.handleUnlockAccount)
}

// handleUnlockAccount clears the failed-login counters and any active
// locked_member_accounts row for a member_id, so a developer doesn't
// have to wait out the real 15-minute lockout window while testing.
func (h *AuthHandler) handleUnlockAccount(ctx *gin.Context) {
	var req models.UnlockAccountRequest
	if !utils.ValidatePayload(ctx, &req) {
		return
	}

	ip := req.IPAddress
	if ip == "" {
		ip = ctx.ClientIP()
	}

	if err := h.authService.UnlockAccountService(ctx, req.MemberID, ip); err != nil {
		loggers.GetCommonError(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	loggers.SuccessResponse(ctx, cc.SUCCESS_ACCOUNT_UNLOCKED)
}
