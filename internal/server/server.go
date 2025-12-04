// Package server provides HTTP server setup and configuration.
package server

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/sebasr/avt-service/internal/auth"
	"github.com/sebasr/avt-service/internal/config"
	"github.com/sebasr/avt-service/internal/email"
	"github.com/sebasr/avt-service/internal/handlers"
	"github.com/sebasr/avt-service/internal/middleware"
	"github.com/sebasr/avt-service/internal/repository"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new UUID for request ID
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// NewRateLimitMiddleware creates a rate limiting middleware using ulule/limiter.
// It allows 100 requests per minute per IP address.
func NewRateLimitMiddleware() gin.HandlerFunc {
	// Define rate: 100 requests per 1 minute
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100,
	}

	// Create in-memory store
	store := memory.NewStore()

	// Create rate limiter instance
	instance := limiter.New(store, rate)

	// Create and return Gin middleware
	middleware := mgin.NewMiddleware(instance)

	return middleware
}

// Dependencies holds all dependencies needed to create a server
type Dependencies struct {
	Config           *config.Config
	TelemetryRepo    repository.TelemetryRepository
	UserRepo         repository.UserRepository
	RefreshTokenRepo repository.RefreshTokenRepository
	DeviceRepo       repository.DeviceRepository
	EmailService     email.Service // Optional: nil if email not configured
}

// New creates a new Gin router with all routes configured
func New(deps *Dependencies) *gin.Engine {
	// Set Gin to release mode to disable ANSI colors in logs
	gin.SetMode(gin.ReleaseMode)

	// Use gin.New() instead of gin.Default() to have explicit control over middleware
	// gin.Default() includes colored logging which contaminates HTTP responses with ANSI codes
	router := gin.New()

	// Add recovery middleware (without colored output)
	router.Use(gin.Recovery())

	// Add logger middleware without colored output
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(_ gin.LogFormatterParams) string {
			// Custom log format without ANSI color codes
			return ""
		},
		Output:    nil,                        // Disable output to prevent any log contamination
		SkipPaths: []string{"/api/v1/health"}, // Skip health check logging
	}))

	// Add CORS middleware for web client support
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Content-Encoding", "Authorization", "X-Request-ID", "X-Batch-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Add middlewares
	router.Use(RequestIDMiddleware())
	router.Use(NewRateLimitMiddleware())
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		deps.Config.Auth.JWTSecret,
		deps.Config.Auth.JWTAccessTokenTTL,
		deps.Config.Auth.JWTRefreshTokenTTL,
	)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	authRateLimiter := middleware.NewAuthRateLimitMiddleware()

	// Initialize handlers
	telemetryHandler := handlers.NewTelemetryHandler(deps.TelemetryRepo, deps.DeviceRepo)
	authHandler := handlers.NewAuthHandler(deps.UserRepo, deps.RefreshTokenRepo, jwtService)

	// Configure email service if available
	if deps.EmailService != nil {
		authHandler = authHandler.WithEmailService(deps.EmailService)
		if deps.Config.Email.ResetTokenTTL > 0 {
			authHandler = authHandler.WithResetTokenTTL(deps.Config.Email.ResetTokenTTL)
		}
	}

	userHandler := handlers.NewUserHandler(deps.UserRepo)
	deviceHandler := handlers.NewDeviceHandler(deps.DeviceRepo)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint for network quality detection
		v1.GET("/health", func(c *gin.Context) {
			c.PureJSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"version":   "1.0.0",
			})
		})

		// Auth routes (with stricter rate limiting)
		authGroup := v1.Group("/auth")
		authGroup.Use(authRateLimiter)
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/forgot-password", authHandler.ForgotPassword)
			authGroup.POST("/reset-password", authHandler.ResetPassword)
		}

		// Telemetry routes (optional auth for backward compatibility)
		v1.POST("/telemetry", authMiddleware.Optional(), telemetryHandler.HandlePost)
		v1.POST("/telemetry/batch", authMiddleware.Optional(), telemetryHandler.HandleBatchPost)

		// Protected user routes
		users := v1.Group("/users")
		users.Use(authMiddleware.Required())
		{
			users.GET("/me", userHandler.GetProfile)
			users.PATCH("/me", userHandler.UpdateProfile)
			users.POST("/me/change-password", userHandler.ChangePassword)
		}

		// Protected device routes
		devices := v1.Group("/devices")
		devices.Use(authMiddleware.Required())
		{
			devices.GET("", deviceHandler.ListDevices)
			devices.GET("/:id", deviceHandler.GetDevice)
			devices.PATCH("/:id", deviceHandler.UpdateDevice)
			devices.DELETE("/:id", deviceHandler.DeactivateDevice)
		}
	}

	// Legacy routes (for backward compatibility)
	router.POST("/api/telemetry", authMiddleware.Optional(), telemetryHandler.HandlePost)
	router.POST("/api/telemetry/batch", authMiddleware.Optional(), telemetryHandler.HandleBatchPost)

	return router
}
