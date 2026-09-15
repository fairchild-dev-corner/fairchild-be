package handlers

import (
	service "fairchild_be/internal/services/activity"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	activityService *service.ActivityService
}

func NewHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

// RegisterRoutes mounts /activity, guarded by requireAuth - a shared
// middlewares.RequireMemberAuth instance built once in cmd/api/api.go and
// passed to every per-member feature, so the caller's own client_id/
// branch_id are resolved exactly once per request rather than once per
// feature.
func (h *ActivityHandler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	activity := router.Group("/activity")
	activity.GET("", requireAuth, h.handleMonthlyActivity)
}
