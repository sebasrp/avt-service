-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Enable PostGIS for spatial queries
CREATE EXTENSION IF NOT EXISTS postgis;

-- Main telemetry table
CREATE TABLE telemetry (
    id BIGSERIAL,
    recorded_at TIMESTAMPTZ NOT NULL,
    device_id VARCHAR(50),
    session_id UUID,
    
    -- Timestamp data
    itow BIGINT,
    time_accuracy BIGINT,
    validity_flags INTEGER,
    
    -- GPS position
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    location GEOGRAPHY(POINT, 4326),
    
    -- GPS altitude
    wgs_altitude DOUBLE PRECISION,
    msl_altitude DOUBLE PRECISION,
    
    -- GPS velocity
    speed DOUBLE PRECISION,
    heading DOUBLE PRECISION,
    
    -- GPS quality
    num_satellites SMALLINT,
    fix_status SMALLINT,
    is_fix_valid BOOLEAN,
    horizontal_accuracy DOUBLE PRECISION,
    vertical_accuracy DOUBLE PRECISION,
    speed_accuracy DOUBLE PRECISION,
    heading_accuracy DOUBLE PRECISION,
    pdop DOUBLE PRECISION,
    
    -- Motion data (accelerometer)
    g_force_x DOUBLE PRECISION,
    g_force_y DOUBLE PRECISION,
    g_force_z DOUBLE PRECISION,
    
    -- Motion data (gyroscope)
    rotation_x DOUBLE PRECISION,
    rotation_y DOUBLE PRECISION,
    rotation_z DOUBLE PRECISION,
    
    -- Device metadata
    battery DOUBLE PRECISION,
    is_charging BOOLEAN,
    
    PRIMARY KEY (recorded_at, id)
);

-- Convert to TimescaleDB hypertable (partitioned by time)
SELECT create_hypertable('telemetry', 'recorded_at');

-- Create indexes for common query patterns
-- Note: For hypertables, unique indexes must include the partitioning column (recorded_at)
CREATE INDEX idx_telemetry_device_time ON telemetry (device_id, recorded_at DESC);
CREATE INDEX idx_telemetry_session ON telemetry (session_id, recorded_at DESC) WHERE session_id IS NOT NULL;
CREATE INDEX idx_telemetry_location ON telemetry USING GIST(location);
CREATE INDEX idx_telemetry_speed ON telemetry (speed, recorded_at DESC) WHERE speed > 0;

-- Enable compression (compress data older than 7 days)
ALTER TABLE telemetry SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id'
);

SELECT add_compression_policy('telemetry', INTERVAL '7 days');

-- Optional: Add retention policy (uncomment to auto-delete old data)
-- SELECT add_retention_policy('telemetry', INTERVAL '1 year');

