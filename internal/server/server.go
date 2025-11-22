// Package server provides HTTP server setup and configuration.
package server

import (
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sebasr/avt-service/internal/handlers"
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

// New creates a new Gin router with all routes configured
func New(repo repository.TelemetryRepository) *gin.Engine {
	router := gin.Default()

	// Add middlewares
	router.Use(RequestIDMiddleware())
	router.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithDecompressFn(gzip.DefaultDecompressHandle)))

	// Initialize handlers
	telemetryHandler := handlers.NewTelemetryHandler(repo)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint for network quality detection
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"version":   "1.0.0",
			})
		})

		v1.POST("/telemetry", telemetryHandler.HandlePost)
		v1.POST("/telemetry/batch", telemetryHandler.HandleBatchPost)
	}

	// Legacy routes (for backward compatibility) - redirect to v1
	router.POST("/api/telemetry", telemetryHandler.HandlePost)
	router.POST("/api/telemetry/batch", telemetryHandler.HandleBatchPost)

	return router
}
