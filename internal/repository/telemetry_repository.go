// Package repository provides data access interfaces and implementations.
package repository

import (
	"context"
	"time"

	"github.com/sebasr/avt-service/internal/models"
)

// TelemetryRepository defines the interface for telemetry data access
type TelemetryRepository interface {
	// Save saves a single telemetry data point
	Save(ctx context.Context, data *models.TelemetryData) error

	// SaveBatch saves multiple telemetry data points in a single transaction
	SaveBatch(ctx context.Context, data []*models.TelemetryData) error

	// GetByTimeRange retrieves telemetry data within a time range
	GetByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*models.TelemetryData, error)

	// GetBySession retrieves telemetry data for a specific session
	GetBySession(ctx context.Context, sessionID string, limit int) ([]*models.TelemetryData, error)

	// GetRecent retrieves the most recent telemetry data points
	GetRecent(ctx context.Context, limit int) ([]*models.TelemetryData, error)

	// GetByDevice retrieves telemetry data for a specific device
	GetByDevice(ctx context.Context, deviceID string, limit int) ([]*models.TelemetryData, error)

	// IsBatchProcessed checks if a batch with the given ID has already been processed
	IsBatchProcessed(ctx context.Context, batchID string) (bool, error)

	// MarkBatchProcessed marks a batch as processed for idempotency
	MarkBatchProcessed(ctx context.Context, batchID string, recordCount int, deviceID string, sessionID *string) error
}
