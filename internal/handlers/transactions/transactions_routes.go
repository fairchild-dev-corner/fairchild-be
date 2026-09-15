package handlers

import (
	service "fairchild_be/internal/services/transactions"

	"github.com/gin-gonic/gin"
)

type TransactionsHandler struct {
	transactionsService *service.TransactionsService
}

func NewHandler(transactionsService *service.TransactionsService) *TransactionsHandler {
	return &TransactionsHandler{transactionsService: transactionsService}
}

// RegisterRoutes mounts /transactions, guarded by requireAuth - a shared
// middlewares.RequireMemberAuth instance built once in cmd/api/api.go and
// passed to every per-member feature, so the caller's own client_id/
// branch_id are resolved exactly once per request rather than once per
// feature.
func (h *TransactionsHandler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	transactions := router.Group("/transactions")
	transactions.GET("", requireAuth, h.handleLastTransactions)
}
