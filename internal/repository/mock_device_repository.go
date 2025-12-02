package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sebasr/avt-service/internal/models"
)

// MockDeviceRepository is a mock implementation of DeviceRepository for testing
type MockDeviceRepository struct {
	CreateFunc         func(ctx context.Context, device *models.Device) error
	GetByIDFunc        func(ctx context.Context, id uuid.UUID) (*models.Device, error)
	GetByDeviceIDFunc  func(ctx context.Context, deviceID string) (*models.Device, error)
	ListByUserIDFunc   func(ctx context.Context, userID uuid.UUID) ([]*models.Device, error)
	UpdateFunc         func(ctx context.Context, device *models.Device) error
	UpdateLastSeenFunc func(ctx context.Context, deviceID string) error
}

// NewMockDeviceRepository creates a new mock device repository
func NewMockDeviceRepository() *MockDeviceRepository {
	return &MockDeviceRepository{
		CreateFunc: func(_ context.Context, _ *models.Device) error {
			return nil
		},
		GetByIDFunc: func(_ context.Context, _ uuid.UUID) (*models.Device, error) {
			return nil, ErrDeviceNotFound
		},
		GetByDeviceIDFunc: func(_ context.Context, _ string) (*models.Device, error) {
			return nil, ErrDeviceNotFound
		},
		ListByUserIDFunc: func(_ context.Context, _ uuid.UUID) ([]*models.Device, error) {
			return []*models.Device{}, nil
		},
		UpdateFunc: func(_ context.Context, _ *models.Device) error {
			return nil
		},
		UpdateLastSeenFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}
}

// Create implements DeviceRepository.Create
func (m *MockDeviceRepository) Create(ctx context.Context, device *models.Device) error {
	return m.CreateFunc(ctx, device)
}

// GetByID implements DeviceRepository.GetByID
func (m *MockDeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	return m.GetByIDFunc(ctx, id)
}

// GetByDeviceID implements DeviceRepository.GetByDeviceID
func (m *MockDeviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	return m.GetByDeviceIDFunc(ctx, deviceID)
}

// ListByUserID implements DeviceRepository.ListByUserID
func (m *MockDeviceRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Device, error) {
	return m.ListByUserIDFunc(ctx, userID)
}

// Update implements DeviceRepository.Update
func (m *MockDeviceRepository) Update(ctx context.Context, device *models.Device) error {
	return m.UpdateFunc(ctx, device)
}

// UpdateLastSeen implements DeviceRepository.UpdateLastSeen
func (m *MockDeviceRepository) UpdateLastSeen(ctx context.Context, deviceID string) error {
	return m.UpdateLastSeenFunc(ctx, deviceID)
}
