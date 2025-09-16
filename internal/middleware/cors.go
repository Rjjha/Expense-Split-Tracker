package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a CORS middleware with appropriate settings
func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()

	// Allow all origins in development
	config.AllowAllOrigins = true

	// Allow common headers
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"Idempotency-Key",
		"X-Requested-With",
	}

	// Allow all common methods
	config.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
		"HEAD",
		"OPTIONS",
	}

	// Expose custom headers
	config.ExposeHeaders = []string{
		"X-Idempotent-Replayed",
		"X-Request-ID",
	}

	return cors.New(config)
}
