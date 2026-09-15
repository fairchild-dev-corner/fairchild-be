package router

import (
	"strings"

	logger "fairchild_be/internal/utils/loggers"

	"github.com/gin-gonic/gin"
)

/**
 * App Router Configuration
 * -------------------------
 * This section sets up the application's routing structure using Gin.
 * It defines the base API path (/api/v1), applies global middleware for CORS,
 * and ensures proper handling of CORS preflight (OPTIONS) requests for all API endpoints.
 *
 * Middleware Order:
 * - CORS middleware must run before JWT so unauthenticated preflight requests pass.
 *   Per-route JWT guarding is applied via AuthHandler.RequireAuth() (see auth_routes.go).
 *
 * Note:
 * - Subrouter is used to scope API versioning.
 * - OPTIONS handler is registered to prevent 404 errors on CORS preflight.
 */

type RouterConfig struct {
	router           *gin.Engine
	originsPermitted string
}

// Initialize Routes
func NewRouterConfig(m *gin.Engine, origins string) *RouterConfig {
	return &RouterConfig{
		router:           m,
		originsPermitted: origins,
	}
}

func (r *RouterConfig) InitConfig() {
	r.router.Use(logger.ErrorHandler())
	r.router.Use(r.CorsMiddleware())
}

func (r RouterConfig) RoutesGroup() *gin.RouterGroup {
	return r.router.Group("/api/v1")
}

// CORS middleware function definition
func (r *RouterConfig) CorsMiddleware() gin.HandlerFunc {
	// Define allowed origins as a comma-separated string

	var allowedOrigins []string
	if r.originsPermitted != "" {
		// Split the originsString into individual origins and store them in allowedOrigins slice
		allowedOrigins = strings.Split(r.originsPermitted, ",")
	}

	// Return the actual middleware handler function
	return func(c *gin.Context) {
		// Function to check if a given origin is allowed
		isOriginAllowed := func(origin string, allowedOrigins []string) bool {
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			return false
		}

		// Get the Origin header from the request
		origin := c.Request.Header.Get("Origin")

		// Check if the origin is allowed
		if isOriginAllowed(origin, allowedOrigins) {
			// If the origin is allowed, set CORS headers in the response
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight OPTIONS requests by aborting with status 204
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Call the next handler
		c.Next()
	}
}
