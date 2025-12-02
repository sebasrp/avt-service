// Package main is the entry point for the AVT service HTTP server.
package main

import (
	"log"

	"github.com/sebasr/avt-service/internal/config"
	"github.com/sebasr/avt-service/internal/database"
	"github.com/sebasr/avt-service/internal/repository"
	"github.com/sebasr/avt-service/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	log.Println("Successfully connected to database")

	// Create repositories
	telemetryRepo := repository.NewPostgresRepository(db)
	userRepo := repository.NewPostgresUserRepository(db)
	refreshTokenRepo := repository.NewPostgresRefreshTokenRepository(db.DB)
	deviceRepo := repository.NewPostgresDeviceRepository(db.DB)

	// Create server dependencies
	deps := &server.Dependencies{
		Config:           cfg,
		TelemetryRepo:    telemetryRepo,
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
		DeviceRepo:       deviceRepo,
	}

	// Create and start the server
	srv := server.New(deps)

	log.Printf("Starting server on port %s", cfg.Server.Port)
	if err := srv.Run(":" + cfg.Server.Port); err != nil {
		log.Printf("Failed to start server: %v", err)
		panic(err) // Use panic instead of log.Fatalf to ensure defer runs
	}
}
