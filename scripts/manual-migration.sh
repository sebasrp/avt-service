#!/bin/bash
# Quick script to manually run migrations on production server
# Run this on your Hetzner server if migrations haven't been applied yet

set -e

echo "=== Manual Database Migration Script ==="
echo ""
echo "This script will create the telemetry, sessions, and upload_batches tables"
echo "in your production database."
echo ""

# Check if we're in the right directory
if [ ! -f "docker-compose.prod.yml" ]; then
    echo "Error: docker-compose.prod.yml not found!"
    echo "Please run this script from /opt/avt-service directory"
    exit 1
fi

echo "Step 1: Creating telemetry table..."
docker compose -f docker-compose.prod.yml exec -T timescaledb psql -U telemetry_user -d telemetry_prod << 'EOF'
-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Enable PostGIS for spatial queries
CREATE EXTENSION IF NOT EXISTS postgis;

-- Main telemetry table
CREATE TABLE IF NOT EXISTS telemetry (
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
SELECT create_hypertable('telemetry', 'recorded_at', if_not_exists => TRUE);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry (device_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_telemetry_session ON telemetry (session_id, recorded_at DESC) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_telemetry_location ON telemetry USING GIST(location);
CREATE INDEX IF NOT EXISTS idx_telemetry_speed ON telemetry (speed) WHERE speed > 0;

-- Enable compression (compress data older than 7 days)
ALTER TABLE telemetry SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id'
);

-- Add compression policy (only if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM timescaledb_information.jobs 
        WHERE proc_name = 'policy_compression' 
        AND hypertable_name = 'telemetry'
    ) THEN
        PERFORM add_compression_policy('telemetry', INTERVAL '7 days');
    END IF;
END $$;
EOF

echo "✓ Telemetry table created"
echo ""

echo "Step 2: Creating sessions table..."
docker compose -f docker-compose.prod.yml exec -T timescaledb psql -U telemetry_user -d telemetry_prod << 'EOF'
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    telemetry_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_device_started ON sessions (device_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_started ON sessions (started_at DESC);
EOF

echo "✓ Sessions table created"
echo ""

echo "Step 3: Creating upload_batches table..."
docker compose -f docker-compose.prod.yml exec -T timescaledb psql -U telemetry_user -d telemetry_prod << 'EOF'
CREATE TABLE IF NOT EXISTS upload_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(50) NOT NULL,
    batch_size INTEGER NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_upload_batches_device ON upload_batches (device_id, uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_upload_batches_uploaded ON upload_batches (uploaded_at DESC);
EOF

echo "✓ Upload batches table created"
echo ""

echo "Step 4: Verifying tables..."
docker compose -f docker-compose.prod.yml exec -T timescaledb psql -U telemetry_user -d telemetry_prod -c "\dt"

echo ""
echo "=== Migration Complete! ==="
echo ""
echo "All tables have been created successfully."
echo "You can now test your API endpoints."
echo ""
echo "Test with:"
echo "  curl https://YOUR_DOMAIN/api/v1/health"
echo ""