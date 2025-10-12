package repository

import (
	"context"
	"time"

	"github.com/sebasr/avt-service/internal/models"
)

// MockRepository is a mock implementation of TelemetryRepository for testing
type MockRepository struct {
	SaveFunc           func(ctx context.Context, data *models.TelemetryData) error
	SaveBatchFunc      func(ctx context.Context, data []*models.TelemetryData) error
	GetByTimeRangeFunc func(ctx context.Context, start, end time.Time, limit int) ([]*models.TelemetryData, error)
	GetBySessionFunc   func(ctx context.Context, sessionID string, limit int) ([]*models.TelemetryData, error)
	GetRecentFunc      func(ctx context.Context, limit int) ([]*models.TelemetryData, error)
	GetByDeviceFunc    func(ctx context.Context, deviceID string, limit int) ([]*models.TelemetryData, error)
}

// NewMockRepository creates a new mock repository with default implementations
func NewMockRepository() *MockRepository {
	return &MockRepository{
		SaveFunc: func(_ context.Context, _ *models.TelemetryData) error {
			return nil
		},
		SaveBatchFunc: func(_ context.Context, _ []*models.TelemetryData) error {
			return nil
		},
		GetByTimeRangeFunc: func(_ context.Context, _ time.Time, _ time.Time, _ int) ([]*models.TelemetryData, error) {
			return []*models.TelemetryData{}, nil
		},
		GetBySessionFunc: func(_ context.Context, _ string, _ int) ([]*models.TelemetryData, error) {
			return []*models.TelemetryData{}, nil
		},
		GetRecentFunc: func(_ context.Context, _ int) ([]*models.TelemetryData, error) {
			return []*models.TelemetryData{}, nil
		},
		GetByDeviceFunc: func(_ context.Context, _ string, _ int) ([]*models.TelemetryData, error) {
			return []*models.TelemetryData{}, nil
		},
	}
}

// Save implements TelemetryRepository.Save
func (m *MockRepository) Save(ctx context.Context, data *models.TelemetryData) error {
	return m.SaveFunc(ctx, data)
}

// SaveBatch implements TelemetryRepository.SaveBatch
func (m *MockRepository) SaveBatch(ctx context.Context, data []*models.TelemetryData) error {
	return m.SaveBatchFunc(ctx, data)
}

// GetByTimeRange implements TelemetryRepository.GetByTimeRange
func (m *MockRepository) GetByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*models.TelemetryData, error) {
	return m.GetByTimeRangeFunc(ctx, start, end, limit)
}

// GetBySession implements TelemetryRepository.GetBySession
func (m *MockRepository) GetBySession(ctx context.Context, sessionID string, limit int) ([]*models.TelemetryData, error) {
	return m.GetBySessionFunc(ctx, sessionID, limit)
}

// GetRecent implements TelemetryRepository.GetRecent
func (m *MockRepository) GetRecent(ctx context.Context, limit int) ([]*models.TelemetryData, error) {
	return m.GetRecentFunc(ctx, limit)
}

// GetByDevice implements TelemetryRepository.GetByDevice
func (m *MockRepository) GetByDevice(ctx context.Context, deviceID string, limit int) ([]*models.TelemetryData, error) {
	return m.GetByDeviceFunc(ctx, deviceID, limit)
}
