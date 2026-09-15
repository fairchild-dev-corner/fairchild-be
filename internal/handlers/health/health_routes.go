package handlers

import (
	service "fairchild_be/internal/services/health"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	healthService *service.HealthService
}

func NewHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

// RegisterRoutes mounts /health, guarded by requireAuth - a shared
// middlewares.RequireMemberAuth instance built once in cmd/api/api.go and
// passed to every per-member feature, so the caller's own client_id/
// branch_id are resolved exactly once per request rather than once per
// feature.
func (h *HealthHandler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	health := router.Group("/health")
	health.GET("", requireAuth, h.handleCoopHealth)
}
