package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/sebasr/avt-service/internal/database"
	"github.com/sebasr/avt-service/internal/models"
)

// setupTestDB sets up a TimescaleDB test container and returns a database connection
func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	ctx := context.Background()

	// Set Docker socket for Colima if not already set
	if os.Getenv("DOCKER_HOST") == "" {
		// Try common Colima socket location
		colimaSocket := os.ExpandEnv("$HOME/.colima/default/docker.sock")
		if _, err := os.Stat(colimaSocket); err == nil {
			os.Setenv("DOCKER_HOST", "unix://"+colimaSocket)
			// Disable Ryuk container for Colima (socket can't be mounted)
			os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
			t.Logf("Using Colima Docker socket: %s (Ryuk disabled)", colimaSocket)
		}
	}

	// Create TimescaleDB container with PostGIS support
	pgContainer, err := postgres.Run(ctx,
		"timescale/timescaledb-ha:pg16",
		postgres.WithDatabase("test_telemetry"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to database
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	db := &database.DB{DB: sqlDB}

	// Run migrations
	if err := runTestMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

// runTestMigrations runs the database migrations for testing
func runTestMigrations(db *database.DB) error {
	migrations := []string{
		// Enable extensions
		`CREATE EXTENSION IF NOT EXISTS timescaledb;`,
		`CREATE EXTENSION IF NOT EXISTS postgis;`,

		// Create users table (needed for foreign keys)
		`CREATE TABLE users (
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
		);`,

		// Create telemetry table
		`CREATE TABLE telemetry (
			id BIGSERIAL,
			recorded_at TIMESTAMPTZ NOT NULL,
			device_id VARCHAR(50),
			session_id UUID,
			user_id UUID,
			itow BIGINT,
			time_accuracy BIGINT,
			validity_flags INTEGER,
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			location GEOGRAPHY(POINT, 4326),
			wgs_altitude DOUBLE PRECISION,
			msl_altitude DOUBLE PRECISION,
			speed DOUBLE PRECISION,
			heading DOUBLE PRECISION,
			num_satellites SMALLINT,
			fix_status SMALLINT,
			is_fix_valid BOOLEAN,
			horizontal_accuracy DOUBLE PRECISION,
			vertical_accuracy DOUBLE PRECISION,
			speed_accuracy DOUBLE PRECISION,
			heading_accuracy DOUBLE PRECISION,
			pdop DOUBLE PRECISION,
			g_force_x DOUBLE PRECISION,
			g_force_y DOUBLE PRECISION,
			g_force_z DOUBLE PRECISION,
			rotation_x DOUBLE PRECISION,
			rotation_y DOUBLE PRECISION,
			rotation_z DOUBLE PRECISION,
			battery DOUBLE PRECISION,
			is_charging BOOLEAN,
			PRIMARY KEY (recorded_at, id)
		);`,

		// Convert to hypertable
		`SELECT create_hypertable('telemetry', 'recorded_at');`,

		// Add foreign key constraint for user_id (after hypertable creation)
		`ALTER TABLE telemetry ADD CONSTRAINT fk_telemetry_user
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;`,

		// Create indexes
		`CREATE INDEX idx_telemetry_device_time ON telemetry (device_id, recorded_at DESC);`,
		`CREATE INDEX idx_telemetry_session ON telemetry (session_id, recorded_at DESC) WHERE session_id IS NOT NULL;`,
		`CREATE INDEX idx_telemetry_user ON telemetry(user_id, recorded_at DESC) WHERE user_id IS NOT NULL;`,

		// Create upload_batches table for idempotency
		`CREATE TABLE upload_batches (
			batch_id VARCHAR(36) PRIMARY KEY,
			record_count INTEGER NOT NULL,
			uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			server_response TEXT,
			device_id VARCHAR(50),
			session_id UUID,
			user_id UUID
		);`,

		// Add foreign key for upload_batches
		`ALTER TABLE upload_batches ADD CONSTRAINT fk_upload_batches_user
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;`,

		// Create indexes for upload_batches
		`CREATE INDEX idx_upload_batches_uploaded_at ON upload_batches (uploaded_at DESC);`,
		`CREATE INDEX idx_upload_batches_device ON upload_batches (device_id, uploaded_at DESC) WHERE device_id IS NOT NULL;`,
		`CREATE INDEX idx_upload_batches_session ON upload_batches (session_id, uploaded_at DESC) WHERE session_id IS NOT NULL;`,
		`CREATE INDEX idx_upload_batches_user ON upload_batches(user_id, uploaded_at DESC) WHERE user_id IS NOT NULL;`,
	}

	ctx := context.Background()
	for _, migration := range migrations {
		if _, err := db.ExecContext(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// createSampleTelemetry creates a sample telemetry data for testing
func createSampleTelemetry(timestamp time.Time, deviceID string) *models.TelemetryData {
	return &models.TelemetryData{
		Timestamp:     timestamp,
		DeviceID:      deviceID,
		ITOW:          118286240,
		TimeAccuracy:  25,
		ValidityFlags: 7,
		GPS: models.GpsData{
			Latitude:           42.6719035,
			Longitude:          23.2887238,
			WgsAltitude:        625.761,
			MslAltitude:        590.095,
			Speed:              125.5,
			Heading:            270.5,
			NumSatellites:      11,
			FixStatus:          3,
			HorizontalAccuracy: 0.924,
			VerticalAccuracy:   1.836,
			SpeedAccuracy:      0.704,
			HeadingAccuracy:    145.26856,
			PDOP:               3.0,
			IsFixValid:         true,
		},
		Motion: models.MotionData{
			GForceX:   -0.003,
			GForceY:   0.113,
			GForceZ:   0.974,
			RotationX: 2.09,
			RotationY: 0.86,
			RotationZ: 0.04,
		},
		Battery:    89.0,
		IsCharging: false,
	}
}

func TestPostgresRepository_Save(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	telemetry := createSampleTelemetry(time.Now().UTC(), "device-001")

	// Save telemetry
	err := repo.Save(ctx, telemetry)
	if err != nil {
		t.Fatalf("Failed to save telemetry: %v", err)
	}

	// Verify ID was assigned
	if telemetry.ID == 0 {
		t.Error("Expected ID to be assigned after save")
	}

	t.Logf("Successfully saved telemetry with ID: %d", telemetry.ID)
}

func TestPostgresRepository_SaveBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	baseTime := time.Now().UTC()
	telemetryBatch := []*models.TelemetryData{
		createSampleTelemetry(baseTime, "device-001"),
		createSampleTelemetry(baseTime.Add(1*time.Second), "device-001"),
		createSampleTelemetry(baseTime.Add(2*time.Second), "device-001"),
	}

	// Save batch
	err := repo.SaveBatch(ctx, telemetryBatch)
	if err != nil {
		t.Fatalf("Failed to save batch: %v", err)
	}

	// Verify data was saved
	recent, err := repo.GetRecent(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get recent telemetry: %v", err)
	}

	if len(recent) != 3 {
		t.Errorf("Expected 3 telemetry records, got %d", len(recent))
	}

	t.Logf("Successfully saved batch of %d records", len(telemetryBatch))
}

func TestPostgresRepository_GetByTimeRange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert test data
	baseTime := time.Now().UTC().Truncate(time.Second)
	for i := 0; i < 5; i++ {
		telemetry := createSampleTelemetry(baseTime.Add(time.Duration(i)*time.Minute), "device-001")
		if err := repo.Save(ctx, telemetry); err != nil {
			t.Fatalf("Failed to save telemetry: %v", err)
		}
	}

	// Query time range (BETWEEN is inclusive on both ends)
	start := baseTime.Add(-1 * time.Minute)
	end := baseTime.Add(3 * time.Minute)
	results, err := repo.GetByTimeRange(ctx, start, end, 100)
	if err != nil {
		t.Fatalf("Failed to query by time range: %v", err)
	}

	// Should get records at 0, 1, 2, 3 minutes (4 records, since BETWEEN is inclusive)
	expectedCount := 4
	if len(results) != expectedCount {
		t.Errorf("Expected %d records, got %d", expectedCount, len(results))
	}

	t.Logf("Retrieved %d records in time range", len(results))
}

func TestPostgresRepository_GetBySession(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert test data with session ID
	sessionID := "550e8400-e29b-41d4-a716-446655440000"
	baseTime := time.Now().UTC()
	for i := 0; i < 3; i++ {
		telemetry := createSampleTelemetry(baseTime.Add(time.Duration(i)*time.Second), "device-001")
		telemetry.SessionID = &sessionID
		if err := repo.Save(ctx, telemetry); err != nil {
			t.Fatalf("Failed to save telemetry: %v", err)
		}
	}

	// Insert data without session ID
	telemetry := createSampleTelemetry(baseTime.Add(10*time.Second), "device-001")
	if err := repo.Save(ctx, telemetry); err != nil {
		t.Fatalf("Failed to save telemetry: %v", err)
	}

	// Query by session
	results, err := repo.GetBySession(ctx, sessionID, 100)
	if err != nil {
		t.Fatalf("Failed to query by session: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 records for session, got %d", len(results))
	}

	// Verify all results have the same session ID
	for _, r := range results {
		if r.SessionID == nil || *r.SessionID != sessionID {
			t.Errorf("Expected session ID %s, got %v", sessionID, r.SessionID)
		}
	}

	t.Logf("Retrieved %d records for session", len(results))
}

func TestPostgresRepository_GetRecent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert test data
	baseTime := time.Now().UTC()
	for i := 0; i < 10; i++ {
		telemetry := createSampleTelemetry(baseTime.Add(time.Duration(i)*time.Second), "device-001")
		if err := repo.Save(ctx, telemetry); err != nil {
			t.Fatalf("Failed to save telemetry: %v", err)
		}
	}

	// Get recent with limit
	results, err := repo.GetRecent(ctx, 5)
	if err != nil {
		t.Fatalf("Failed to get recent telemetry: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 recent records, got %d", len(results))
	}

	// Verify results are ordered by time (most recent first)
	for i := 0; i < len(results)-1; i++ {
		if results[i].Timestamp.Before(results[i+1].Timestamp) {
			t.Error("Expected results to be ordered by timestamp descending")
		}
	}

	t.Logf("Retrieved %d recent records", len(results))
}

func TestPostgresRepository_GetByDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Insert test data for multiple devices
	baseTime := time.Now().UTC()
	for i := 0; i < 3; i++ {
		telemetry1 := createSampleTelemetry(baseTime.Add(time.Duration(i)*time.Second), "device-001")
		telemetry2 := createSampleTelemetry(baseTime.Add(time.Duration(i)*time.Second), "device-002")

		if err := repo.Save(ctx, telemetry1); err != nil {
			t.Fatalf("Failed to save telemetry for device-001: %v", err)
		}
		if err := repo.Save(ctx, telemetry2); err != nil {
			t.Fatalf("Failed to save telemetry for device-002: %v", err)
		}
	}

	// Query by device
	results, err := repo.GetByDevice(ctx, "device-001", 100)
	if err != nil {
		t.Fatalf("Failed to query by device: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 records for device-001, got %d", len(results))
	}

	// Verify all results are for the correct device
	for _, r := range results {
		if r.DeviceID != "device-001" {
			t.Errorf("Expected device ID 'device-001', got '%s'", r.DeviceID)
		}
	}

	t.Logf("Retrieved %d records for device-001", len(results))
}
