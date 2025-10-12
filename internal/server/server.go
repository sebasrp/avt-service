// Package server provides HTTP server setup and configuration.
package server

import (
	"github.com/gin-gonic/gin"

	"github.com/sebasr/avt-service/internal/handlers"
	"github.com/sebasr/avt-service/internal/repository"
)

// New creates a new Gin router with all routes configured
func New(repo repository.TelemetryRepository) *gin.Engine {
	router := gin.Default()

	// Initialize handlers
	telemetryHandler := handlers.NewTelemetryHandler(repo)

	// Register routes
	router.POST("/api/telemetry", telemetryHandler.HandlePost)

	return router
}
