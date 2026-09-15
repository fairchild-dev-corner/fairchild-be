package handlers

import (
	service "fairchild_be/internal/services/balance"

	"github.com/gin-gonic/gin"
)

type BalanceHandler struct {
	balanceService *service.BalanceService
}

func NewHandler(balanceService *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}

// RegisterRoutes mounts /sd/balance, guarded by requireAuth - a shared
// middlewares.RequireMemberAuth instance built once in cmd/api/api.go and
// passed to every per-member feature, so the caller's own client_id/
// branch_id are resolved exactly once per request rather than once per
// feature.
func (h *BalanceHandler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	sd := router.Group("/sd")
	sd.GET("/balance", requireAuth, h.handleDepositBalances)
}
