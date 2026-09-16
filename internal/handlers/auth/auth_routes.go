package handlers

import (
	"net/http"
	"strings"
	"time"

	middlewares "fairchild_be/internal/middlewares"
	service "fairchild_be/internal/services/auth"
	loggers "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const ctxUserIDKey = "auth_user_id"

const (
	// authRateLimitRefill: sustained refill - 1 token every 2s.
	authRateLimitRefill = 2 * time.Second
	// authRateLimitBurst: initial burst allowance before throttling kicks in.
	// Combined with authRateLimitRefill this yields ~5 requests per 10s per IP.
	authRateLimitBurst = 5
)

// refreshTokenCookieName is shared by AuthHandler and SSOHandler - both issue
// and clear the same cookie.
const refreshTokenCookieName = "refresh_token"

type AuthHandler struct {
	authService *service.AuthService
	buildEnv    string
}

func NewHandler(authService *service.AuthService, buildEnv string) *AuthHandler {
	return &AuthHandler{authService: authService, buildEnv: buildEnv}
}

func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {

	auth := router.Group("/auth")
	limiter := middlewares.NewIPRateLimiter(rate.Every(authRateLimitRefill), authRateLimitBurst)
	auth.Use(limiter.Middleware())

	auth.POST("/register", h.handleRegister)
	auth.POST("/register/send-otp", h.handleSendRegisterOTP)
	auth.POST("/register/verify-otp", h.handleVerifyRegisterOTP)
	auth.POST("/register/young-saver", h.handleRegisterYoungSaver)
	auth.POST("/login", h.handleLogin)
	auth.POST("/login/verify-otp", h.handleVerifyLoginOTP)
	auth.POST("/forgot-password", h.handleForgotPassword)
	auth.POST("/forgot-password/verify-otp", h.handleVerifyForgotPasswordOTP)
	auth.POST("/forgot-password/reset", h.handleResetForgotPassword)
	auth.POST("/account-claim", h.handleClaimAccount)
	auth.POST("/social/login", h.handleSocialLogin)
	auth.POST("/refresh", h.handleRefreshToken)
	auth.POST("/logout", h.handleLogout)

	// Guarded auth
	auth.GET("/profile", h.RequireAuth(), h.handleProfile)
	auth.PATCH("/profile", h.RequireAuth(), h.handleUpdateProfile)
	auth.PATCH("/password", h.RequireAuth(), h.handleChangePassword)
	auth.GET("/settings", h.RequireAuth(), h.handleGetSettings)
	auth.PATCH("/settings", h.RequireAuth(), h.handleUpdateSettings)
	auth.GET("/sessions", h.RequireAuth(), h.handleSessions)
}

// RequireAuth validates the "Authorization: Bearer <token>" header and, on
// success, stores the authenticated user id in the request context for
// downstream handlers to read via userIDFromContext. Apply this middleware
// to any route group that must be guarded by a valid JWT access token.
func (h *AuthHandler) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		const prefix = "Bearer "

		if header == "" || !strings.HasPrefix(header, prefix) {
			loggers.GetCommonError(ctx, "missing bearer token", http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		claims, err := h.authService.ParseAccessToken(strings.TrimPrefix(header, prefix))
		if err != nil {
			loggers.GetCommonError(ctx, err.Error(), http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		ctx.Set(ctxUserIDKey, claims.UserID)
		ctx.Next()
	}
}

// setRefreshTokenCookie stores the refresh token as an HttpOnly cookie -
// it is scoped to /api/v1/auth (the only paths that ever need it: refresh,
// logout, and the SSO callback that issues it) and never exposed to JS, so
// an XSS bug can't exfiltrate the long-lived credential. In dev the frontend
// and backend are both "localhost" on different ports (same-site, plain
// HTTP), so SameSite=Lax/Secure=false works; stage/prod may be on different
// registrable domains, so SameSite=None/Secure=true is required there (the
// CORS middleware already reflects a specific origin with
// Access-Control-Allow-Credentials: true - see router_config.go).
func setRefreshTokenCookie(ctx *gin.Context, buildEnv, token string, maxAge time.Duration) {
	setRefreshTokenCookieRaw(ctx, buildEnv, token, int(maxAge.Seconds()))
}

// clearRefreshTokenCookie deletes the refresh_token cookie set by
// setRefreshTokenCookie - the SameSite/Secure attributes must match exactly,
// otherwise the browser treats it as a different cookie and won't delete it.
// maxAge=-1 (not 0) is what tells the browser to expire it immediately.
func clearRefreshTokenCookie(ctx *gin.Context, buildEnv string) {
	setRefreshTokenCookieRaw(ctx, buildEnv, "", -1)
}

func setRefreshTokenCookieRaw(ctx *gin.Context, buildEnv, token string, maxAgeSeconds int) {
	secure := buildEnv != "dev"
	if secure {
		ctx.SetSameSite(http.SameSiteNoneMode)
	} else {
		ctx.SetSameSite(http.SameSiteLaxMode)
	}
	ctx.SetCookie(refreshTokenCookieName, token, maxAgeSeconds, "/api/v1/auth", "", secure, true)
}

func userIDFromContext(ctx *gin.Context) (int64, bool) {
	v, exists := ctx.Get(ctxUserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
