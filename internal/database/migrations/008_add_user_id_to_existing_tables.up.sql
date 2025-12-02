-- Add user_id column to existing tables for user ownership tracking

-- Add user_id to telemetry table (hypertable - no foreign key constraint due to TimescaleDB limitations)
-- Note: Foreign keys on hypertables require the partitioning column in the referenced table's primary key
-- Since users table is not partitioned by time, we cannot add a foreign key constraint
-- Application logic must ensure referential integrity
ALTER TABLE telemetry ADD COLUMN user_id UUID;
CREATE INDEX idx_telemetry_user ON telemetry(user_id, recorded_at DESC) WHERE user_id IS NOT NULL;

-- Add user_id to sessions table
ALTER TABLE sessions ADD COLUMN user_id UUID;
ALTER TABLE sessions ADD CONSTRAINT fk_sessions_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_sessions_user ON sessions(user_id, started_at DESC) WHERE user_id IS NOT NULL;

-- Add user_id to upload_batches table
ALTER TABLE upload_batches ADD COLUMN user_id UUID;
ALTER TABLE upload_batches ADD CONSTRAINT fk_upload_batches_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_upload_batches_user ON upload_batches(user_id, uploaded_at DESC) WHERE user_id IS NOT NULL;