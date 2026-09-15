package handlers

import (
	service "fairchild_be/internal/services/soa"

	"github.com/gin-gonic/gin"
)

type SOAHandler struct {
	soaService *service.SOAService
}

func NewHandler(soaService *service.SOAService) *SOAHandler {
	return &SOAHandler{soaService: soaService}
}

// RegisterRoutes mounts /soa, guarded by requireAuth - a shared
// middlewares.RequireMemberAuth instance built once in cmd/api/api.go and
// passed to every per-member feature, so the caller's own client_id/
// branch_id are resolved exactly once per request rather than once per
// feature.
func (h *SOAHandler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	soa := router.Group("/soa")
	soa.GET("", requireAuth, h.handleSOAHistory)
	soa.GET("/:ctrl_no", requireAuth, h.handleSOAStatement)
}
