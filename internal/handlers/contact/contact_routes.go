package handlers

import (
	"time"

	middlewares "fairchild_be/internal/middlewares"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	// contactRateLimitRefill: sustained refill - 1 token every 10s.
	contactRateLimitRefill = 10 * time.Second
	// contactRateLimitBurst: initial burst allowance before throttling
	// kicks in. Combined with contactRateLimitRefill this yields ~3
	// submissions per 30s per IP - generous for a genuine visitor, tight
	// enough to blunt naive form-spam bots.
	contactRateLimitBurst = 3
)

// RegisterRoutes mounts the public /contact and /newsletter endpoints - no
// auth required, since both are reachable from marketing pages a visitor
// hits before ever creating an account (see ContactForm.tsx and
// AppPromo.tsx on the frontend).
func (h *ContactHandler) RegisterRoutes(router *gin.RouterGroup) {
	limiter := middlewares.NewIPRateLimiter(rate.Every(contactRateLimitRefill), contactRateLimitBurst)

	router.POST("/contact", limiter.Middleware(), h.handleSubmitContactMessage)
	router.POST("/newsletter", limiter.Middleware(), h.handleSubscribeNewsletter)
}
