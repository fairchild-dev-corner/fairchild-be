package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	cc "fairchild_be/internal/constants"
	middlewares "fairchild_be/internal/middlewares"
	models "fairchild_be/internal/models/auth"
	service "fairchild_be/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"golang.org/x/time/rate"
)

// SSOHandler implements the server-initiated OAuth2 redirect flow via goth,
// as opposed to AuthHandler.handleSocialLogin which re-verifies a token the
// client already obtained from a provider's native SDK. It is a separate
// handler because it speaks in redirects rather than JSON, but it mounts
// under the same /api/v1/auth group as AuthHandler (see RegisterRoutes) -
// the final routes are /api/v1/auth/sso/:provider and .../callback.
type SSOHandler struct {
	authService *service.AuthService
	frontendURL string
	buildEnv    string
}

func NewSSOHandler(authService *service.AuthService, frontendURL, buildEnv string) *SSOHandler {
	return &SSOHandler{authService: authService, frontendURL: frontendURL, buildEnv: buildEnv}
}

func (h *SSOHandler) RegisterRoutes(router *gin.RouterGroup) {
	sso := router.Group("/auth/sso")

	limiter := middlewares.NewIPRateLimiter(rate.Every(authRateLimitRefill), authRateLimitBurst)
	sso.Use(limiter.Middleware())

	sso.GET("/:provider", h.handleSSOBegin)
	sso.GET("/:provider/callback", h.handleSSOCallback)
}

// handleSSOBegin redirects the browser to the requested provider's consent
// screen. gothic reads the provider name off the request context rather than
// gin's own path params, so it must be injected via GetContextWithProvider.
func (h *SSOHandler) handleSSOBegin(ctx *gin.Context) {
	ctx.Request = gothic.GetContextWithProvider(ctx.Request, ctx.Param("provider"))
	gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
}

// handleSSOCallback completes the OAuth2 exchange, looks up the local member
// the verified identity belongs to, sets the refresh token as an HttpOnly
// cookie, and redirects straight to /member/dashboard - no token is passed
// via the URL, the dashboard establishes its session by calling
// POST /auth/refresh on mount.
func (h *SSOHandler) handleSSOCallback(ctx *gin.Context) {
	ctx.Request = gothic.GetContextWithProvider(ctx.Request, ctx.Param("provider"))

	gothUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		slog.Error("SSO callback failed: CompleteUserAuth", "provider", ctx.Param("provider"), "error", err)
		h.redirectWithError(ctx, "sso_failed")
		return
	}

	res, err := h.authService.SSOLoginService(ctx, socialProfileFromGothUser(gothUser))
	if err != nil {
		// SSOLoginService never creates an account - an OAuth identity whose
		// email doesn't match an existing member, or that belongs to a
		// suspended one, is reported distinctly so the frontend can show a
		// real message instead of a generic "something went wrong".
		switch {
		case errors.Is(err, cc.ErrEmailNotAssociated):
			h.redirectWithError(ctx, "not_a_member")
		case errors.Is(err, cc.ErrAccountSuspended):
			h.redirectWithError(ctx, "account_suspended")
		default:
			slog.Error("SSO callback failed: SSOLoginService", "provider", ctx.Param("provider"), "error", err)
			h.redirectWithError(ctx, "sso_failed")
		}
		return
	}

	// The refresh token was just set as an HttpOnly cookie above, so the
	// dashboard doesn't need anything passed via the URL - it establishes
	// the session by calling POST /auth/refresh on mount, which reads that
	// cookie and issues an access token.
	setRefreshTokenCookie(ctx, h.buildEnv, res.RefreshToken, h.authService.RefreshTokenTTL())

	ctx.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/member/dashboard")
}

func (h *SSOHandler) redirectWithError(ctx *gin.Context, reason string) {
	ctx.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"#error="+url.QueryEscape(reason))
}

func socialProfileFromGothUser(u goth.User) *models.SocialProfile {
	var expiresAt *time.Time
	if !u.ExpiresAt.IsZero() {
		t := u.ExpiresAt
		expiresAt = &t
	}

	return &models.SocialProfile{
		Provider:       u.Provider,
		ProviderUserID: u.UserID,
		Email:          u.Email,
		Name:           u.Name,
		AvatarUrl:      u.AvatarURL,
		AccessToken:    u.AccessToken,
		RefreshToken:   u.RefreshToken,
		ExpiresAt:      expiresAt,
	}
}
