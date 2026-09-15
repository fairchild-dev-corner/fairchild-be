package handlers

import (
	loans_service "fairchild_be/internal/services/loans"
	transactions_service "fairchild_be/internal/services/transactions"

	"github.com/gin-gonic/gin"
)

type LoansHandler struct {
	loansService        *loans_service.LoansService
	transactionsService *transactions_service.TransactionsService
}

func NewHandler(loansService *loans_service.LoansService, transactionsService *transactions_service.TransactionsService) *LoansHandler {
	return &LoansHandler{loansService: loansService, transactionsService: transactionsService}
}

// RegisterRoutes mounts /loans, guarded by requireAuth - a shared
// middlewares.RequireMemberAuth instance built once in cmd/api/api.go and
// passed to every per-member feature, so the caller's own client_id/
// branch_id are resolved exactly once per request rather than once per
// feature.
func (h *LoansHandler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	loans := router.Group("/loans")
	loans.GET("", requireAuth, h.handleActiveLoans)
	loans.GET("/transactions/:ref_no", requireAuth, h.handleLoanTransactions)
}
