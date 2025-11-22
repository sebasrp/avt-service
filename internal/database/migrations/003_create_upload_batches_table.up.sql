-- Create upload_batches table for idempotency tracking
CREATE TABLE upload_batches (
    batch_id VARCHAR(36) PRIMARY KEY,
    record_count INTEGER NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    server_response TEXT,
    device_id VARCHAR(50),
    session_id UUID
);

-- Create index for cleanup queries (delete old batches)
CREATE INDEX idx_upload_batches_uploaded_at ON upload_batches (uploaded_at DESC);

-- Create index for device queries
CREATE INDEX idx_upload_batches_device ON upload_batches (device_id, uploaded_at DESC) WHERE device_id IS NOT NULL;

-- Create index for session queries
CREATE INDEX idx_upload_batches_session ON upload_batches (session_id, uploaded_at DESC) WHERE session_id IS NOT NULL;