package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// NewAuthRateLimitMiddleware creates a stricter rate limiting middleware for auth endpoints.
// It allows 10 requests per minute per IP address (vs 100/min for general endpoints).
func NewAuthRateLimitMiddleware() gin.HandlerFunc {
	// Define rate: 10 requests per 1 minute (stricter for auth endpoints)
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  10,
	}

	// Create in-memory store
	store := memory.NewStore()

	// Create rate limiter instance
	instance := limiter.New(store, rate)

	// Create and return Gin middleware
	middleware := mgin.NewMiddleware(instance)

	return middleware
}

// NewAuthRateLimitMiddlewareWithConfig creates a rate limiting middleware with custom configuration
func NewAuthRateLimitMiddlewareWithConfig(limit int64, period time.Duration) gin.HandlerFunc {
	rate := limiter.Rate{
		Period: period,
		Limit:  limit,
	}

	store := memory.NewStore()
	instance := limiter.New(store, rate)
	middleware := mgin.NewMiddleware(instance)

	return middleware
}
