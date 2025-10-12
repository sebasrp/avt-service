package repository

import (
	"context"
	"database/sql"
	"fmt"
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

		// Create telemetry table
		`CREATE TABLE telemetry (
			id BIGSERIAL,
			recorded_at TIMESTAMPTZ NOT NULL,
			device_id VARCHAR(50),
			session_id UUID,
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

		// Create indexes
		`CREATE INDEX idx_telemetry_device_time ON telemetry (device_id, recorded_at DESC);`,
		`CREATE INDEX idx_telemetry_session ON telemetry (session_id, recorded_at DESC) WHERE session_id IS NOT NULL;`,
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

	// Query time range
	start := baseTime.Add(-1 * time.Minute)
	end := baseTime.Add(3 * time.Minute)
	results, err := repo.GetByTimeRange(ctx, start, end, 100)
	if err != nil {
		t.Fatalf("Failed to query by time range: %v", err)
	}

	// Should get records at 0, 1, 2 minutes (3 records)
	expectedCount := 3
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
