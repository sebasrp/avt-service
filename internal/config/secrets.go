// Package config provides configuration management for the AVT service.
package config

import (
	"os"
	"strings"
)

// GetSecret retrieves a secret with multiple fallback sources.
// Priority:
//  1. Direct environment variable (e.g., MAILGUN_API_KEY)
//  2. File path from _FILE environment variable (e.g., MAILGUN_API_KEY_FILE)
//  3. Default value
//
// This allows secrets to be provided via:
//   - Environment variables (e.g., MAILGUN_API_KEY=xxx)
//   - Docker secrets (e.g., MAILGUN_API_KEY_FILE=/run/secrets/mailgun_api_key)
func GetSecret(envVar, defaultValue string) string {
	// 1. Try direct environment variable
	if value := os.Getenv(envVar); value != "" {
		return value
	}

	// 2. Try file-based secret (for Docker secrets)
	if filePath := os.Getenv(envVar + "_FILE"); filePath != "" {
		if data, err := os.ReadFile(filePath); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	// 3. Return default value
	return defaultValue
}
