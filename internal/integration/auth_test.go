package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/sebasr/avt-service/internal/config"
	"github.com/sebasr/avt-service/internal/database"
	"github.com/sebasr/avt-service/internal/repository"
	"github.com/sebasr/avt-service/internal/server"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestDatabase creates a test database using Testcontainers
func setupTestDatabase(t *testing.T) (*database.DB, func()) {
	ctx := context.Background()

	// Create PostgreSQL container with TimescaleDB
	req := testcontainers.ContainerRequest{
		Image:        "timescale/timescaledb:latest-pg15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60 * time.Second),
	}

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get connection details
	host, err := postgres.Host(ctx)
	require.NoError(t, err)

	port, err := postgres.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create database connection
	cfg := &config.DatabaseConfig{
		Host:     host,
		Port:     port.Port(),
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}

	db, err := database.New(cfg)
	require.NoError(t, err)

	// Run migrations
	migrationPath := "../../internal/database/migrations"
	err = runMigrations(db, migrationPath)
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		_ = db.Close()
		_ = postgres.Terminate(ctx)
	}

	return db, cleanup
}

// runMigrations applies all up migrations
func runMigrations(db *database.DB, path string) error {
	// Note: This is a simplified migration runner for tests.
	// In production, use a proper migration tool like golang-migrate or goose

	migrations := []string{
		"001_create_telemetry_table.up.sql",
		"002_create_sessions_table.up.sql",
		"003_create_upload_batches_table.up.sql",
		"004_create_users_table.up.sql",
		"005_create_user_profiles_table.up.sql",
		"006_create_refresh_tokens_table.up.sql",
		"007_create_devices_table.up.sql",
		"008_add_user_id_to_existing_tables.up.sql",
	}

	// Create tables manually for testing
	// Enable TimescaleDB extension
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;")
	if err != nil {
		return fmt.Errorf("failed to enable timescaledb: %w", err)
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			email_verified BOOLEAN DEFAULT FALSE,
			verification_token VARCHAR(255),
			verification_token_expires_at TIMESTAMPTZ,
			reset_token VARCHAR(255),
			reset_token_expires_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_login_at TIMESTAMPTZ,
			is_active BOOLEAN DEFAULT TRUE
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active) WHERE is_active = TRUE;
		
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		
		DROP TRIGGER IF EXISTS update_users_updated_at ON users;
		CREATE TRIGGER update_users_updated_at
			BEFORE UPDATE ON users
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create user_profiles table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_profiles (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			display_name VARCHAR(255),
			avatar_url TEXT,
			timezone VARCHAR(50) DEFAULT 'UTC',
			units_preference VARCHAR(20) DEFAULT 'metric',
			notifications_enabled BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		
		DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
		CREATE TRIGGER update_user_profiles_updated_at
			BEFORE UPDATE ON user_profiles
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`)
	if err != nil {
		return fmt.Errorf("failed to create user_profiles table: %w", err)
	}

	// Create refresh_tokens table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			revoked_at TIMESTAMPTZ,
			replaced_by UUID REFERENCES refresh_tokens(id) ON DELETE SET NULL,
			user_agent TEXT,
			ip_address VARCHAR(45)
		);
		
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);
		CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
	`)
	if err != nil {
		return fmt.Errorf("failed to create refresh_tokens table: %w", err)
	}

	// Create devices table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS devices (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			device_id VARCHAR(255) UNIQUE NOT NULL,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			device_name VARCHAR(255),
			device_model VARCHAR(255),
			claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_seen_at TIMESTAMPTZ,
			is_active BOOLEAN DEFAULT TRUE,
			metadata JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id, is_active);
		CREATE INDEX IF NOT EXISTS idx_devices_device_id ON devices(device_id);
		
		DROP TRIGGER IF EXISTS update_devices_updated_at ON devices;
		CREATE TRIGGER update_devices_updated_at
			BEFORE UPDATE ON devices
			FOR EACH ROW
			EXECUTE FUNCTION update_updated_at_column();
	`)
	if err != nil {
		return fmt.Errorf("failed to create devices table: %w", err)
	}

	// Create telemetry table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS telemetry (
			id BIGSERIAL PRIMARY KEY,
			recorded_at TIMESTAMPTZ NOT NULL,
			device_id VARCHAR(255),
			session_id VARCHAR(255),
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			itow BIGINT NOT NULL,
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			wgs_altitude DOUBLE PRECISION NOT NULL,
			msl_altitude DOUBLE PRECISION NOT NULL,
			speed DOUBLE PRECISION NOT NULL,
			heading DOUBLE PRECISION NOT NULL,
			num_satellites INTEGER NOT NULL,
			fix_status INTEGER NOT NULL,
			horizontal_accuracy DOUBLE PRECISION NOT NULL,
			vertical_accuracy DOUBLE PRECISION NOT NULL,
			speed_accuracy DOUBLE PRECISION NOT NULL,
			heading_accuracy DOUBLE PRECISION NOT NULL,
			pdop DOUBLE PRECISION NOT NULL,
			is_fix_valid BOOLEAN NOT NULL,
			g_force_x DOUBLE PRECISION NOT NULL,
			g_force_y DOUBLE PRECISION NOT NULL,
			g_force_z DOUBLE PRECISION NOT NULL,
			rotation_x DOUBLE PRECISION NOT NULL,
			rotation_y DOUBLE PRECISION NOT NULL,
			rotation_z DOUBLE PRECISION NOT NULL,
			battery DOUBLE PRECISION NOT NULL,
			is_charging BOOLEAN NOT NULL,
			time_accuracy BIGINT NOT NULL,
			validity_flags INTEGER NOT NULL
		);
		
		SELECT create_hypertable('telemetry', 'recorded_at', if_not_exists => TRUE);
		CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, recorded_at DESC);
		CREATE INDEX IF NOT EXISTS idx_telemetry_session ON telemetry(session_id, recorded_at DESC) WHERE session_id IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_telemetry_user ON telemetry(user_id, recorded_at DESC) WHERE user_id IS NOT NULL;
	`)
	if err != nil {
		return fmt.Errorf("failed to create telemetry table: %w", err)
	}

	_ = migrations // Suppress unused warning
	return nil
}

// TestFullRegistrationFlow tests the complete user registration flow
func TestFullRegistrationFlow(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	// Setup dependencies
	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:          "test-secret-integration",
			JWTAccessTokenTTL:  time.Hour,
			JWTRefreshTokenTTL: 24 * time.Hour,
		},
	}

	deps := &server.Dependencies{
		Config:           cfg,
		TelemetryRepo:    repository.NewPostgresRepository(db),
		UserRepo:         repository.NewPostgresUserRepository(db),
		RefreshTokenRepo: repository.NewPostgresRefreshTokenRepository(db.DB),
		DeviceRepo:       repository.NewPostgresDeviceRepository(db.DB),
	}

	router := server.New(deps)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "newuser@example.com",
			"password": "securePassword123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "accessToken")
		assert.Contains(t, response, "refreshToken")

		// Verify user was created in database
		user, err := deps.UserRepo.GetByEmail(context.Background(), "newuser@example.com")
		require.NoError(t, err)
		assert.Equal(t, "newuser@example.com", user.Email)
		assert.NotEmpty(t, user.PasswordHash)
	})

	t.Run("duplicate email registration", func(t *testing.T) {
		reqBody := map[string]string{
			"email":    "newuser@example.com", // Same email as above
			"password": "anotherPassword123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
	})
}

// TestFullLoginFlow tests the complete login flow
func TestFullLoginFlow(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:          "test-secret-integration",
			JWTAccessTokenTTL:  time.Hour,
			JWTRefreshTokenTTL: 24 * time.Hour,
		},
	}

	deps := &server.Dependencies{
		Config:           cfg,
		TelemetryRepo:    repository.NewPostgresRepository(db),
		UserRepo:         repository.NewPostgresUserRepository(db),
		RefreshTokenRepo: repository.NewPostgresRefreshTokenRepository(db.DB),
		DeviceRepo:       repository.NewPostgresDeviceRepository(db.DB),
	}

	router := server.New(deps)

	// First, register a user
	email := "logintest@example.com"
	password := "testPassword123"

	registerBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	t.Run("successful login", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    email,
			"password": password,
		}
		body, _ := json.Marshal(loginBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "user")
		assert.Contains(t, response, "accessToken")
		assert.Contains(t, response, "refreshToken")

		// Verify tokens are valid
		accessToken := response["accessToken"].(string)
		assert.NotEmpty(t, accessToken)

		// Verify last_login_at was updated
		user, err := deps.UserRepo.GetByEmail(context.Background(), email)
		require.NoError(t, err)
		assert.NotNil(t, user.LastLoginAt)
	})

	t.Run("login with wrong password", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    email,
			"password": "wrongPassword",
		}
		body, _ := json.Marshal(loginBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("login with non-existent email", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": password,
		}
		body, _ := json.Marshal(loginBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestTokenRefreshFlow tests the token refresh flow
func TestTokenRefreshFlow(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:          "test-secret-integration",
			JWTAccessTokenTTL:  time.Hour,
			JWTRefreshTokenTTL: 24 * time.Hour,
		},
	}

	deps := &server.Dependencies{
		Config:           cfg,
		TelemetryRepo:    repository.NewPostgresRepository(db),
		UserRepo:         repository.NewPostgresUserRepository(db),
		RefreshTokenRepo: repository.NewPostgresRefreshTokenRepository(db.DB),
		DeviceRepo:       repository.NewPostgresDeviceRepository(db.DB),
	}

	router := server.New(deps)

	// Register and login to get initial tokens
	email := "refreshtest@example.com"
	password := "testPassword123"

	registerBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	refreshToken := registerResponse["refreshToken"].(string)

	t.Run("successful token refresh", func(t *testing.T) {
		refreshBody := map[string]string{
			"refreshToken": refreshToken,
		}
		body, _ := json.Marshal(refreshBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "accessToken")
		assert.Contains(t, response, "refreshToken")

		newAccessToken := response["accessToken"].(string)
		newRefreshToken := response["refreshToken"].(string)

		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
		assert.NotEqual(t, refreshToken, newRefreshToken) // Token rotation
	})

	t.Run("refresh with invalid token", func(t *testing.T) {
		refreshBody := map[string]string{
			"refreshToken": "invalid-token",
		}
		body, _ := json.Marshal(refreshBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestProtectedEndpointAccess tests accessing protected endpoints
func TestProtectedEndpointAccess(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:          "test-secret-integration",
			JWTAccessTokenTTL:  time.Hour,
			JWTRefreshTokenTTL: 24 * time.Hour,
		},
	}

	deps := &server.Dependencies{
		Config:           cfg,
		TelemetryRepo:    repository.NewPostgresRepository(db),
		UserRepo:         repository.NewPostgresUserRepository(db),
		RefreshTokenRepo: repository.NewPostgresRefreshTokenRepository(db.DB),
		DeviceRepo:       repository.NewPostgresDeviceRepository(db.DB),
	}

	router := server.New(deps)

	// Register a user and get access token
	email := "protectedtest@example.com"
	password := "testPassword123"

	registerBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	accessToken := registerResponse["accessToken"].(string)

	t.Run("access protected endpoint with valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "email")
		assert.Equal(t, email, response["email"])
	})

	t.Run("access protected endpoint without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("access protected endpoint with invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestDeviceClaimingFlow tests the device claiming flow with telemetry
func TestDeviceClaimingFlow(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:          "test-secret-integration",
			JWTAccessTokenTTL:  time.Hour,
			JWTRefreshTokenTTL: 24 * time.Hour,
		},
	}

	deps := &server.Dependencies{
		Config:           cfg,
		TelemetryRepo:    repository.NewPostgresRepository(db),
		UserRepo:         repository.NewPostgresUserRepository(db),
		RefreshTokenRepo: repository.NewPostgresRefreshTokenRepository(db.DB),
		DeviceRepo:       repository.NewPostgresDeviceRepository(db.DB),
	}

	router := server.New(deps)

	// Register a user
	email := "devicetest@example.com"
	password := "testPassword123"

	registerBody := map[string]string{
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	accessToken := registerResponse["accessToken"].(string)
	userMap := registerResponse["user"].(map[string]interface{})
	userID, _ := uuid.Parse(userMap["id"].(string))

	t.Run("device auto-claimed on first telemetry upload", func(t *testing.T) {
		deviceID := "test-device-integration-001"

		telemetryBody := map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"deviceId":  deviceID,
			"iTOW":      118286240,
			"gps": map[string]interface{}{
				"latitude":  42.0,
				"longitude": 23.0,
			},
			"motion": map[string]interface{}{
				"gForceX": 0.0,
				"gForceY": 0.0,
				"gForceZ": 1.0,
			},
			"battery":    85.0,
			"isCharging": false,
		}
		body, _ := json.Marshal(telemetryBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify device was claimed
		device, err := deps.DeviceRepo.GetByDeviceID(context.Background(), deviceID)
		require.NoError(t, err)
		assert.Equal(t, userID, device.UserID)
		assert.Equal(t, deviceID, device.DeviceID)
		assert.True(t, device.IsActive)
	})

	t.Run("user can list their devices", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		devices := response["devices"].([]interface{})
		assert.GreaterOrEqual(t, len(devices), 1)
	})
}
