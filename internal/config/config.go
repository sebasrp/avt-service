// Package config provides configuration management for the AVT service.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL                   string
	Host                  string
	Port                  string
	Name                  string
	User                  string
	Password              string
	SSLMode               string
	MaxConnections        int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			URL:                   os.Getenv("DATABASE_URL"),
			Host:                  getEnv("DB_HOST", "localhost"),
			Port:                  getEnv("DB_PORT", "5432"),
			Name:                  getEnv("DB_NAME", "telemetry_dev"),
			User:                  getEnv("DB_USER", "telemetry_user"),
			Password:              getEnv("DB_PASSWORD", "telemetry_pass"),
			SSLMode:               getEnv("DB_SSLMODE", "disable"),
			MaxConnections:        getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConnections:    getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
			ConnectionMaxLifetime: getEnvAsDuration("DB_CONNECTION_MAX_LIFETIME", "5m"),
		},
	}

	return cfg, nil
}

// ConnectionString returns the database connection string
func (d *DatabaseConfig) ConnectionString() string {
	if d.URL != "" {
		return d.URL
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsDuration gets an environment variable as a duration or returns a default value
func getEnvAsDuration(key, defaultValue string) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		valueStr = defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		defaultDuration, _ := time.ParseDuration(defaultValue)
		return defaultDuration
	}
	return value
}
