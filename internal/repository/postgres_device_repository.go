package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
)

var (
	// ErrDeviceNotFound is returned when a device is not found
	ErrDeviceNotFound = errors.New("device not found")

	// ErrDeviceExists is returned when trying to create a device with an existing device_id
	ErrDeviceExists = errors.New("device already exists")
)

// PostgresDeviceRepository implements DeviceRepository using PostgreSQL
type PostgresDeviceRepository struct {
	db *sql.DB
}

// NewPostgresDeviceRepository creates a new PostgreSQL device repository
func NewPostgresDeviceRepository(db *sql.DB) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{db: db}
}

// Create stores a new device
func (r *PostgresDeviceRepository) Create(ctx context.Context, device *models.Device) error {
	query := `
		INSERT INTO devices (
			id, device_id, user_id, device_name, device_model,
			claimed_at, last_seen_at, is_active, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	var metadataJSON []byte
	var err error
	if device.Metadata != nil {
		metadataJSON, err = json.Marshal(device.Metadata)
		if err != nil {
			return err
		}
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		device.ID,
		device.DeviceID,
		device.UserID,
		device.DeviceName,
		device.DeviceModel,
		device.ClaimedAt,
		device.LastSeenAt,
		device.IsActive,
		metadataJSON,
		device.CreatedAt,
		device.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return ErrDeviceExists
		}
		return err
	}

	return nil
}

// GetByID retrieves a device by its UUID
func (r *PostgresDeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	query := `
		SELECT 
			id, device_id, user_id, device_name, device_model,
			claimed_at, last_seen_at, is_active, metadata,
			created_at, updated_at
		FROM devices
		WHERE id = $1
	`

	var device models.Device
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&device.ID,
		&device.DeviceID,
		&device.UserID,
		&device.DeviceName,
		&device.DeviceModel,
		&device.ClaimedAt,
		&device.LastSeenAt,
		&device.IsActive,
		&metadataJSON,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, err
		}
	}

	return &device, nil
}

// GetByDeviceID retrieves a device by its hardware device ID
func (r *PostgresDeviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	query := `
		SELECT 
			id, device_id, user_id, device_name, device_model,
			claimed_at, last_seen_at, is_active, metadata,
			created_at, updated_at
		FROM devices
		WHERE device_id = $1
	`

	var device models.Device
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&device.ID,
		&device.DeviceID,
		&device.UserID,
		&device.DeviceName,
		&device.DeviceModel,
		&device.ClaimedAt,
		&device.LastSeenAt,
		&device.IsActive,
		&metadataJSON,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, err
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
			return nil, err
		}
	}

	return &device, nil
}

// ListByUserID retrieves all devices owned by a user
func (r *PostgresDeviceRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Device, error) {
	query := `
		SELECT 
			id, device_id, user_id, device_name, device_model,
			claimed_at, last_seen_at, is_active, metadata,
			created_at, updated_at
		FROM devices
		WHERE user_id = $1
		ORDER BY claimed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		var device models.Device
		var metadataJSON []byte

		err := rows.Scan(
			&device.ID,
			&device.DeviceID,
			&device.UserID,
			&device.DeviceName,
			&device.DeviceModel,
			&device.ClaimedAt,
			&device.LastSeenAt,
			&device.IsActive,
			&metadataJSON,
			&device.CreatedAt,
			&device.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &device.Metadata); err != nil {
				return nil, err
			}
		}

		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return devices, nil
}

// Update updates a device's information
func (r *PostgresDeviceRepository) Update(ctx context.Context, device *models.Device) error {
	query := `
		UPDATE devices
		SET 
			device_name = $1,
			device_model = $2,
			last_seen_at = $3,
			is_active = $4,
			metadata = $5,
			updated_at = $6
		WHERE id = $7
	`

	var metadataJSON []byte
	var err error
	if device.Metadata != nil {
		metadataJSON, err = json.Marshal(device.Metadata)
		if err != nil {
			return err
		}
	}

	device.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		device.DeviceName,
		device.DeviceModel,
		device.LastSeenAt,
		device.IsActive,
		metadataJSON,
		device.UpdatedAt,
		device.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeviceNotFound
	}

	return nil
}

// UpdateLastSeen updates the last_seen_at timestamp for a device
func (r *PostgresDeviceRepository) UpdateLastSeen(ctx context.Context, deviceID string) error {
	query := `
		UPDATE devices
		SET last_seen_at = NOW(), updated_at = NOW()
		WHERE device_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeviceNotFound
	}

	return nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// Check for PostgreSQL unique violation error code (23505)
	return err.Error() == "pq: duplicate key value violates unique constraint \"devices_device_id_key\"" ||
		err.Error() == "ERROR: duplicate key value violates unique constraint \"devices_device_id_key\" (SQLSTATE 23505)"
}
