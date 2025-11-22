# Idempotency Implementation

## Overview

The idempotency feature prevents duplicate processing of telemetry data batches when clients retry uploads due to network issues or timeouts. This is critical for the Flutter RaceBox exporter app which operates in environments with spotty cellular connectivity.

## Architecture

### Database Schema

The `upload_batches` table tracks processed batches:

```sql
CREATE TABLE upload_batches (
    batch_id VARCHAR(36) PRIMARY KEY,      -- Client-generated UUID
    record_count INTEGER NOT NULL,          -- Number of records in batch
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    server_response TEXT,                   -- Optional response metadata
    device_id VARCHAR(50),                  -- Device identifier
    session_id UUID                         -- Optional session identifier
);
```

**Indexes:**

- Primary key on `batch_id` for fast duplicate detection
- `idx_upload_batches_uploaded_at` for time-based queries
- `idx_upload_batches_device` for device-specific queries
- `idx_upload_batches_session` for session-specific queries

### Repository Interface

Two new methods were added to [`TelemetryRepository`](../internal/repository/telemetry_repository.go):

```go
// IsBatchProcessed checks if a batch with the given ID has already been processed
IsBatchProcessed(ctx context.Context, batchID string) (bool, error)

// MarkBatchProcessed marks a batch as processed for idempotency
MarkBatchProcessed(ctx context.Context, batchID string, recordCount int, 
                   deviceID string, sessionID *string) error
```

### Implementation Details

#### Duplicate Detection

The [`IsBatchProcessed`](../internal/repository/postgres_repository.go:292) method uses an efficient `EXISTS` query:

```go
query := `SELECT EXISTS(SELECT 1 FROM upload_batches WHERE batch_id = $1)`
```

This returns immediately without scanning the entire table.

#### Batch Marking

The [`MarkBatchProcessed`](../internal/repository/postgres_repository.go:305) method uses PostgreSQL's `ON CONFLICT DO NOTHING` to handle race conditions:

```go
INSERT INTO upload_batches (batch_id, record_count, device_id, session_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (batch_id) DO NOTHING
```

This ensures:

- **Atomicity**: Insert is atomic, preventing race conditions
- **Idempotency**: Duplicate inserts are silently ignored
- **Data Preservation**: Original batch data is never overwritten

## Usage Workflow

### Client-Side (Flutter App)

1. Generate a UUID for each batch before upload
2. Include the `batch_id` in the upload request
3. On network failure or timeout, retry with the same `batch_id`
4. Server will process the batch only once

### Server-Side Handler

```go
// 1. Check if batch was already processed
processed, err := repo.IsBatchProcessed(ctx, request.BatchID)
if err != nil {
    return err
}

if processed {
    // Return success without processing
    return c.JSON(http.StatusOK, gin.H{
        "message": "Batch already processed",
        "batch_id": request.BatchID,
    })
}

// 2. Process the batch
err = repo.SaveBatch(ctx, request.Data)
if err != nil {
    return err
}

// 3. Mark batch as processed
err = repo.MarkBatchProcessed(ctx, request.BatchID, len(request.Data), 
                              deviceID, sessionID)
if err != nil {
    // Log error but don't fail the request
    log.Printf("Failed to mark batch as processed: %v", err)
}
```

## Testing

Comprehensive integration tests verify:

1. **Non-existent batch detection**: Returns `false` for new batches
2. **Existing batch detection**: Returns `true` for processed batches
3. **Successful batch marking**: Correctly stores batch metadata
4. **Duplicate handling**: `ON CONFLICT DO NOTHING` prevents errors
5. **Null session handling**: Supports batches without session IDs
6. **Complete workflow**: End-to-end idempotency verification

Run tests:

```bash
go test -v ./internal/repository -run "Idempotency|IsBatchProcessed|MarkBatchProcessed"
```

## Performance Considerations

### Database Performance

- **Primary key lookup**: O(log n) using B-tree index
- **EXISTS query**: Short-circuits on first match
- **ON CONFLICT**: Single round-trip to database
- **No table scans**: All queries use indexes

### Memory Efficiency

- Batch tracking uses minimal storage (~100 bytes per batch)
- No in-memory caching required
- Database handles all state management

### Scalability

- Supports millions of batches without performance degradation
- Indexes ensure consistent query performance
- Can be partitioned by `uploaded_at` if needed

## Migration

The migration is located at:

- Up: [`003_create_upload_batches_table.up.sql`](../internal/database/migrations/003_create_upload_batches_table.up.sql)
- Down: [`003_create_upload_batches_table.down.sql`](../internal/database/migrations/003_create_upload_batches_table.down.sql)

Apply migration:

```bash
make migrate-up
```

Rollback migration:

```bash
make migrate-down
```

## Security Considerations

1. **UUID Validation**: Client-generated UUIDs should be validated
2. **Rate Limiting**: Prevent abuse by limiting batch submissions
3. **Batch Size Limits**: Enforce maximum records per batch
4. **Retention Policy**: Consider purging old batch records

## Future Enhancements

1. **Batch Expiration**: Auto-delete batches older than N days
2. **Batch Status**: Track processing status (pending/success/failed)
3. **Retry Metadata**: Store retry count and timestamps
4. **Batch Analytics**: Query batch statistics by device/session

## Related Documentation

- [Flutter Integration Architecture](./flutter-integration-architecture.md)
- [Database Schema](./database.md)
- [Testing Guide](./testing.md)
