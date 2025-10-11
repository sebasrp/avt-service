// Package main is the entry point for the AVT service HTTP server.
package main

import (
	"log"
	"os"

	"github.com/sebasr/avt-service/internal/server"
)

func main() {
	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create and start the server
	srv := server.New()

	log.Printf("Starting server on port %s", port)
	if err := srv.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
