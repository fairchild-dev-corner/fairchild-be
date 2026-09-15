package handlers

import (
	service "fairchild_be/internal/services/ar"

	"github.com/gin-gonic/gin"
)

type ARHandler struct {
	arService *service.ARService
}

func NewHandler(arService *service.ARService) *ARHandler {
	return &ARHandler{arService: arService}
}

// RegisterRoutes mounts /ar, guarded by requireAuth - a shared
// middlewares.RequireMemberAuth instance built once in cmd/api/api.go and
// passed to every per-member feature, so the caller's own client_id/
// branch_id are resolved exactly once per request rather than once per
// feature.
func (h *ARHandler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	ar := router.Group("/ar")
	ar.GET("", requireAuth, h.handleActiveAR)
}
