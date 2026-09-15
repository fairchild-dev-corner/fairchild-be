package handlers

import (
	service "fairchild_be/internal/services/amortization"

	"github.com/gin-gonic/gin"
)

type AmortizationHandler struct {
	amortizationService *service.AmortizationService
}

func NewHandler(amortizationService *service.AmortizationService) *AmortizationHandler {
	return &AmortizationHandler{amortizationService: amortizationService}
}

// RegisterRoutes mounts /amortization, guarded by requireAuth - a shared
// middlewares.RequireMemberAuth instance built once in cmd/api/api.go and
// passed to every per-member feature, so the caller's own client_id/
// branch_id are resolved exactly once per request rather than once per
// feature.
func (h *AmortizationHandler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	amortization := router.Group("/amortization")
	amortization.GET("", requireAuth, h.handleAmortizationSchedule)
	amortization.GET("/:ref_no", requireAuth, h.handleLoanSchedule)
}
