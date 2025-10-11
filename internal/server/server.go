// Package server provides HTTP server setup and configuration.
package server

import (
	"github.com/gin-gonic/gin"

	"github.com/sebasr/avt-service/internal/handlers"
)

// New creates a new Gin router with all routes configured
func New() *gin.Engine {
	router := gin.Default()

	// Register routes
	router.GET("/", handlers.HelloHandler)
	router.GET("/health", handlers.HealthHandler)
	router.GET("/api/greeting/:name", handlers.GreetingHandler)
	router.POST("/api/telemetry", handlers.TelemetryHandler)

	return router
}
