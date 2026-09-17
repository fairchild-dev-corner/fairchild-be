package handlers

import (
	"errors"
	"fmt"
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
// the verified identity belongs to, and redirects back to the frontend with
// the issued access token in the URL fragment - fragments are never sent to
// the server or logged, unlike a query string, which matters since this app
// has no cookie-based auth to fall back on for it.
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

	setRefreshTokenCookie(ctx, h.buildEnv, res.RefreshToken, h.authService.RefreshTokenTTL())

	// h.frontendURL (SSO_FRONTEND_REDIRECT_URL) already points at the
	// /sso/complete landing page itself, not the site root - it must not
	// have another path appended here. Only the access token travels in the
	// fragment; the refresh token was just set as an HttpOnly cookie above
	// and must never be JS-readable. sso/complete.tsx is what fetches the
	// profile with that access token, calls setSession to populate the auth
	// store, and only then navigates on to /member/dashboard - that store
	// write is the only thing RequireAuthGuard checks, so skipping this page
	// leaves the user looking logged in on the backend but bounced to
	// /login on the frontend.
	target := fmt.Sprintf("%s#access_token=%s&token_type=%s&expires_in=%d",
		h.frontendURL,
		url.QueryEscape(res.AccessToken),
		url.QueryEscape(res.TokenType),
		res.ExpiresIn,
	)
	ctx.Redirect(http.StatusTemporaryRedirect, target)
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
