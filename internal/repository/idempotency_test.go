package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepository_IsBatchProcessed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Clean up test data
	_, _ = db.ExecContext(ctx, "DELETE FROM upload_batches WHERE batch_id LIKE 'test-%'")

	t.Run("returns false for non-existent batch", func(t *testing.T) {
		exists, err := repo.IsBatchProcessed(ctx, "test-non-existent-batch")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns true for existing batch", func(t *testing.T) {
		batchID := "test-existing-batch"

		// Insert a batch
		err := repo.MarkBatchProcessed(ctx, batchID, 10, "device-123", nil)
		require.NoError(t, err)

		// Check if it exists
		exists, err := repo.IsBatchProcessed(ctx, batchID)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Clean up
		_, _ = db.ExecContext(ctx, "DELETE FROM upload_batches WHERE batch_id = $1", batchID)
	})
}

func TestPostgresRepository_MarkBatchProcessed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Clean up test data
	_, _ = db.ExecContext(ctx, "DELETE FROM upload_batches WHERE batch_id LIKE 'test-%'")

	t.Run("successfully marks batch as processed", func(t *testing.T) {
		batchID := "test-batch-1"
		deviceID := "device-456"
		sessionID := "550e8400-e29b-41d4-a716-446655440001" // Valid UUID format

		err := repo.MarkBatchProcessed(ctx, batchID, 25, deviceID, &sessionID)
		assert.NoError(t, err)

		// Verify it was inserted
		var storedDeviceID string
		var storedSessionID *string
		var recordCount int

		err = db.QueryRowContext(ctx,
			"SELECT record_count, device_id, session_id FROM upload_batches WHERE batch_id = $1",
			batchID,
		).Scan(&recordCount, &storedDeviceID, &storedSessionID)

		assert.NoError(t, err)
		assert.Equal(t, 25, recordCount)
		assert.Equal(t, deviceID, storedDeviceID)
		assert.NotNil(t, storedSessionID)
		assert.Equal(t, sessionID, *storedSessionID)

		// Clean up
		_, _ = db.ExecContext(ctx, "DELETE FROM upload_batches WHERE batch_id = $1", batchID)
	})

	t.Run("handles duplicate batch ID gracefully (ON CONFLICT DO NOTHING)", func(t *testing.T) {
		batchID := "test-batch-duplicate"

		// Insert first time
		err := repo.MarkBatchProcessed(ctx, batchID, 10, "device-1", nil)
		assert.NoError(t, err)

		// Insert again with different data - should not error due to ON CONFLICT DO NOTHING
		err = repo.MarkBatchProcessed(ctx, batchID, 20, "device-2", nil)
		assert.NoError(t, err)

		// Verify original data is preserved
		var recordCount int
		var deviceID string
		err = db.QueryRowContext(ctx,
			"SELECT record_count, device_id FROM upload_batches WHERE batch_id = $1",
			batchID,
		).Scan(&recordCount, &deviceID)

		assert.NoError(t, err)
		assert.Equal(t, 10, recordCount, "Original record count should be preserved")
		assert.Equal(t, "device-1", deviceID, "Original device ID should be preserved")

		// Clean up
		_, _ = db.ExecContext(ctx, "DELETE FROM upload_batches WHERE batch_id = $1", batchID)
	})

	t.Run("handles nil session ID", func(t *testing.T) {
		batchID := "test-batch-nil-session"

		err := repo.MarkBatchProcessed(ctx, batchID, 15, "device-999", nil)
		assert.NoError(t, err)

		// Verify session_id is NULL
		var sessionID *string
		err = db.QueryRowContext(ctx,
			"SELECT session_id FROM upload_batches WHERE batch_id = $1",
			batchID,
		).Scan(&sessionID)

		assert.NoError(t, err)
		assert.Nil(t, sessionID)

		// Clean up
		_, _ = db.ExecContext(ctx, "DELETE FROM upload_batches WHERE batch_id = $1", batchID)
	})
}

func TestPostgresRepository_IdempotencyWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	// Clean up test data
	_, _ = db.ExecContext(ctx, "DELETE FROM upload_batches WHERE batch_id LIKE 'test-%'")

	t.Run("complete idempotency workflow", func(t *testing.T) {
		batchID := "test-workflow-batch"
		deviceID := "device-workflow"
		sessionID := "550e8400-e29b-41d4-a716-446655440000" // Valid UUID

		// Step 1: Check if batch exists (should be false)
		exists, err := repo.IsBatchProcessed(ctx, batchID)
		assert.NoError(t, err)
		assert.False(t, exists, "Batch should not exist initially")

		// Step 2: Mark batch as processed
		err = repo.MarkBatchProcessed(ctx, batchID, 100, deviceID, &sessionID)
		assert.NoError(t, err)

		// Step 3: Check if batch exists (should be true now)
		exists, err = repo.IsBatchProcessed(ctx, batchID)
		assert.NoError(t, err)
		assert.True(t, exists, "Batch should exist after marking as processed")

		// Step 4: Try to mark again (should not error)
		err = repo.MarkBatchProcessed(ctx, batchID, 200, "different-device", nil)
		assert.NoError(t, err, "Duplicate marking should not error")

		// Step 5: Verify original data is preserved
		var recordCount int
		err = db.QueryRowContext(ctx,
			"SELECT record_count FROM upload_batches WHERE batch_id = $1",
			batchID,
		).Scan(&recordCount)

		assert.NoError(t, err)
		assert.Equal(t, 100, recordCount, "Original record count should be preserved")

		// Clean up
		_, _ = db.ExecContext(ctx, "DELETE FROM upload_batches WHERE batch_id = $1", batchID)
	})
}
