-- Remove user_id column from existing tables

-- Remove from upload_batches table
DROP INDEX IF EXISTS idx_upload_batches_user;
ALTER TABLE upload_batches DROP CONSTRAINT IF EXISTS fk_upload_batches_user;
ALTER TABLE upload_batches DROP COLUMN IF EXISTS user_id;

-- Remove from sessions table
DROP INDEX IF EXISTS idx_sessions_user;
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS fk_sessions_user;
ALTER TABLE sessions DROP COLUMN IF EXISTS user_id;

-- Remove from telemetry table
DROP INDEX IF EXISTS idx_telemetry_user;
ALTER TABLE telemetry DROP CONSTRAINT IF EXISTS fk_telemetry_user;
ALTER TABLE telemetry DROP COLUMN IF EXISTS user_id;