package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
)

// DeviceRepository defines the interface for device data access
type DeviceRepository interface {
	// Create stores a new device
	Create(ctx context.Context, device *models.Device) error

	// GetByID retrieves a device by its UUID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error)

	// GetByDeviceID retrieves a device by its hardware device ID
	GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error)

	// ListByUserID retrieves all devices owned by a user
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Device, error)

	// Update updates a device's information
	Update(ctx context.Context, device *models.Device) error

	// UpdateLastSeen updates the last_seen_at timestamp for a device
	UpdateLastSeen(ctx context.Context, deviceID string) error
}
