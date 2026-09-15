package handlers

import (
	"net/http"
	"strings"

	auth_service "fairchild_be/internal/services/auth"
	service "fairchild_be/internal/services/profile"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
)

const ctxUserIDKey = "profile_user_id"

type ProfileHandler struct {
	profileService *service.ProfileService
	tokens         *auth_service.TokenService
}

func NewHandler(profileService *service.ProfileService, tokens *auth_service.TokenService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService, tokens: tokens}
}

func (h *ProfileHandler) RegisterRoutes(router *gin.RouterGroup) {
	profile := router.Group("/profile")
	profile.GET("", h.requireAuth(), h.handleProfile)
}

// requireAuth validates the "Authorization: Bearer <token>" header the same
// way AuthHandler.RequireAuth does - this endpoint returns another member's
// personal data, so it must never be reachable without a valid access
// token.
func (h *ProfileHandler) requireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		const prefix = "Bearer "

		if header == "" || !strings.HasPrefix(header, prefix) {
			loggers.GetCommonError(ctx, "missing bearer token", http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		claims, err := h.tokens.ParseAccessToken(strings.TrimPrefix(header, prefix))
		if err != nil {
			loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		ctx.Set(ctxUserIDKey, claims.UserID)
		ctx.Next()
	}
}

func userIDFromContext(ctx *gin.Context) (int64, bool) {
	v, exists := ctx.Get(ctxUserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
