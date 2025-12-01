-- Drop devices table and related objects
DROP TRIGGER IF EXISTS update_devices_updated_at ON devices;
DROP INDEX IF EXISTS idx_devices_claimed_at;
DROP INDEX IF EXISTS idx_devices_last_seen;
DROP INDEX IF EXISTS idx_devices_device_id;
DROP INDEX IF EXISTS idx_devices_user;
DROP TABLE IF EXISTS devices;