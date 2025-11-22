package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sebasr/avt-service/internal/database"
	"github.com/sebasr/avt-service/internal/models"
)

// PostgresRepository implements TelemetryRepository using PostgreSQL/TimescaleDB
type PostgresRepository struct {
	db *database.DB
}

// NewPostgresRepository creates a new PostgreSQL telemetry repository
func NewPostgresRepository(db *database.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Save saves a single telemetry data point
func (r *PostgresRepository) Save(ctx context.Context, data *models.TelemetryData) error {
	query := `
		INSERT INTO telemetry (
			recorded_at, device_id, session_id, itow, time_accuracy, validity_flags,
			latitude, longitude, location,
			wgs_altitude, msl_altitude, speed, heading,
			num_satellites, fix_status, is_fix_valid,
			horizontal_accuracy, vertical_accuracy, speed_accuracy, heading_accuracy, pdop,
			g_force_x, g_force_y, g_force_z,
			rotation_x, rotation_y, rotation_z,
			battery, is_charging
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, ST_SetSRID(ST_MakePoint($8, $7), 4326)::geography,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23,
			$24, $25, $26,
			$27, $28
		)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		data.Timestamp, data.DeviceID, data.SessionID,
		data.ITOW, data.TimeAccuracy, data.ValidityFlags,
		data.GPS.Latitude, data.GPS.Longitude,
		data.GPS.WgsAltitude, data.GPS.MslAltitude, data.GPS.Speed, data.GPS.Heading,
		data.GPS.NumSatellites, data.GPS.FixStatus, data.GPS.IsFixValid,
		data.GPS.HorizontalAccuracy, data.GPS.VerticalAccuracy,
		data.GPS.SpeedAccuracy, data.GPS.HeadingAccuracy, data.GPS.PDOP,
		data.Motion.GForceX, data.Motion.GForceY, data.Motion.GForceZ,
		data.Motion.RotationX, data.Motion.RotationY, data.Motion.RotationZ,
		data.Battery, data.IsCharging,
	).Scan(&data.ID)

	if err != nil {
		return fmt.Errorf("failed to insert telemetry: %w", err)
	}

	return nil
}

// SaveBatch saves multiple telemetry data points in a single transaction
func (r *PostgresRepository) SaveBatch(ctx context.Context, dataPoints []*models.TelemetryData) error {
	if len(dataPoints) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback is safe to call even after Commit
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO telemetry (
			recorded_at, device_id, session_id, itow, time_accuracy, validity_flags,
			latitude, longitude, location,
			wgs_altitude, msl_altitude, speed, heading,
			num_satellites, fix_status, is_fix_valid,
			horizontal_accuracy, vertical_accuracy, speed_accuracy, heading_accuracy, pdop,
			g_force_x, g_force_y, g_force_z,
			rotation_x, rotation_y, rotation_z,
			battery, is_charging
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, ST_SetSRID(ST_MakePoint($8, $7), 4326)::geography,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23,
			$24, $25, $26,
			$27, $28
		)
		RETURNING id
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, data := range dataPoints {
		err := stmt.QueryRowContext(ctx,
			data.Timestamp, data.DeviceID, data.SessionID,
			data.ITOW, data.TimeAccuracy, data.ValidityFlags,
			data.GPS.Latitude, data.GPS.Longitude,
			data.GPS.WgsAltitude, data.GPS.MslAltitude, data.GPS.Speed, data.GPS.Heading,
			data.GPS.NumSatellites, data.GPS.FixStatus, data.GPS.IsFixValid,
			data.GPS.HorizontalAccuracy, data.GPS.VerticalAccuracy,
			data.GPS.SpeedAccuracy, data.GPS.HeadingAccuracy, data.GPS.PDOP,
			data.Motion.GForceX, data.Motion.GForceY, data.Motion.GForceZ,
			data.Motion.RotationX, data.Motion.RotationY, data.Motion.RotationZ,
			data.Battery, data.IsCharging,
		).Scan(&data.ID)
		if err != nil {
			return fmt.Errorf("failed to insert telemetry in batch: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByTimeRange retrieves telemetry data within a time range
func (r *PostgresRepository) GetByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*models.TelemetryData, error) {
	if limit <= 0 {
		limit = 1000
	}

	query := `
		SELECT 
			id, recorded_at, device_id, session_id, itow, time_accuracy, validity_flags,
			latitude, longitude, wgs_altitude, msl_altitude, speed, heading,
			num_satellites, fix_status, is_fix_valid,
			horizontal_accuracy, vertical_accuracy, speed_accuracy, heading_accuracy, pdop,
			g_force_x, g_force_y, g_force_z,
			rotation_x, rotation_y, rotation_z,
			battery, is_charging
		FROM telemetry
		WHERE recorded_at BETWEEN $1 AND $2
		ORDER BY recorded_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry by time range: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryRows(rows)
}

// GetBySession retrieves telemetry data for a specific session
func (r *PostgresRepository) GetBySession(ctx context.Context, sessionID string, limit int) ([]*models.TelemetryData, error) {
	if limit <= 0 {
		limit = 10000
	}

	query := `
		SELECT 
			id, recorded_at, device_id, session_id, itow, time_accuracy, validity_flags,
			latitude, longitude, wgs_altitude, msl_altitude, speed, heading,
			num_satellites, fix_status, is_fix_valid,
			horizontal_accuracy, vertical_accuracy, speed_accuracy, heading_accuracy, pdop,
			g_force_x, g_force_y, g_force_z,
			rotation_x, rotation_y, rotation_z,
			battery, is_charging
		FROM telemetry
		WHERE session_id = $1
		ORDER BY recorded_at ASC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry by session: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryRows(rows)
}

// GetRecent retrieves the most recent telemetry data points
func (r *PostgresRepository) GetRecent(ctx context.Context, limit int) ([]*models.TelemetryData, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT 
			id, recorded_at, device_id, session_id, itow, time_accuracy, validity_flags,
			latitude, longitude, wgs_altitude, msl_altitude, speed, heading,
			num_satellites, fix_status, is_fix_valid,
			horizontal_accuracy, vertical_accuracy, speed_accuracy, heading_accuracy, pdop,
			g_force_x, g_force_y, g_force_z,
			rotation_x, rotation_y, rotation_z,
			battery, is_charging
		FROM telemetry
		ORDER BY recorded_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent telemetry: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryRows(rows)
}

// GetByDevice retrieves telemetry data for a specific device
func (r *PostgresRepository) GetByDevice(ctx context.Context, deviceID string, limit int) ([]*models.TelemetryData, error) {
	if limit <= 0 {
		limit = 1000
	}

	query := `
		SELECT 
			id, recorded_at, device_id, session_id, itow, time_accuracy, validity_flags,
			latitude, longitude, wgs_altitude, msl_altitude, speed, heading,
			num_satellites, fix_status, is_fix_valid,
			horizontal_accuracy, vertical_accuracy, speed_accuracy, heading_accuracy, pdop,
			g_force_x, g_force_y, g_force_z,
			rotation_x, rotation_y, rotation_z,
			battery, is_charging
		FROM telemetry
		WHERE device_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry by device: %w", err)
	}
	defer rows.Close()

	return r.scanTelemetryRows(rows)
}

// scanTelemetryRows scans database rows into TelemetryData structs
func (r *PostgresRepository) scanTelemetryRows(rows *sql.Rows) ([]*models.TelemetryData, error) {
	var results []*models.TelemetryData

	for rows.Next() {
		data := &models.TelemetryData{}
		var sessionID sql.NullString

		err := rows.Scan(
			&data.ID, &data.Timestamp, &data.DeviceID, &sessionID,
			&data.ITOW, &data.TimeAccuracy, &data.ValidityFlags,
			&data.GPS.Latitude, &data.GPS.Longitude,
			&data.GPS.WgsAltitude, &data.GPS.MslAltitude, &data.GPS.Speed, &data.GPS.Heading,
			&data.GPS.NumSatellites, &data.GPS.FixStatus, &data.GPS.IsFixValid,
			&data.GPS.HorizontalAccuracy, &data.GPS.VerticalAccuracy,
			&data.GPS.SpeedAccuracy, &data.GPS.HeadingAccuracy, &data.GPS.PDOP,
			&data.Motion.GForceX, &data.Motion.GForceY, &data.Motion.GForceZ,
			&data.Motion.RotationX, &data.Motion.RotationY, &data.Motion.RotationZ,
			&data.Battery, &data.IsCharging,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry row: %w", err)
		}

		if sessionID.Valid {
			data.SessionID = &sessionID.String
		}

		results = append(results, data)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating telemetry rows: %w", err)
	}

	return results, nil
}
