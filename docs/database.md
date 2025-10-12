# Database Documentation

## Overview

The AVT Service uses **TimescaleDB**, a PostgreSQL extension optimized for time-series data, to store telemetry information from vehicles.

## Why TimescaleDB?

- **Time-Series Optimized**: Automatic partitioning by time
- **PostgreSQL Compatible**: Standard SQL with all PostgreSQL features
- **High Performance**: 100k+ inserts/second
- **Efficient Compression**: 10-20x data reduction
- **PostGIS Support**: Native geospatial queries
- **Retention Policies**: Automatic data lifecycle management

## Schema

### Telemetry Table

The main `telemetry` table is a **hypertable** automatically partitioned by the `recorded_at` timestamp.

#### Structure

```sql
CREATE TABLE telemetry (
    -- Primary key
    id BIGSERIAL,
    recorded_at TIMESTAMPTZ NOT NULL,

    -- Device and session identifiers
    device_id VARCHAR(50),
    session_id UUID,

    -- Timestamp data
    itow BIGINT,                      -- GPS time of week (milliseconds)
    time_accuracy BIGINT,             -- Time accuracy (nanoseconds)
    validity_flags INTEGER,

    -- GPS position
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    location GEOGRAPHY(POINT, 4326),  -- PostGIS point for spatial queries

    -- GPS altitude
    wgs_altitude DOUBLE PRECISION,    -- WGS84 altitude (meters)
    msl_altitude DOUBLE PRECISION,    -- Mean sea level altitude (meters)

    -- GPS velocity
    speed DOUBLE PRECISION,           -- Speed (km/h)
    heading DOUBLE PRECISION,         -- Heading (degrees, 0-360)

    -- GPS quality
    num_satellites SMALLINT,
    fix_status SMALLINT,              -- 0=no fix, 2=2D, 3=3D
    is_fix_valid BOOLEAN,
    horizontal_accuracy DOUBLE PRECISION,  -- Meters
    vertical_accuracy DOUBLE PRECISION,    -- Meters
    speed_accuracy DOUBLE PRECISION,       -- km/h
    heading_accuracy DOUBLE PRECISION,     -- Degrees
    pdop DOUBLE PRECISION,            -- Position dilution of precision

    -- Motion data (accelerometer)
    g_force_x DOUBLE PRECISION,       -- Front/back
    g_force_y DOUBLE PRECISION,       -- Right/left
    g_force_z DOUBLE PRECISION,       -- Up/down

    -- Motion data (gyroscope)
    rotation_x DOUBLE PRECISION,      -- Roll (degrees/second)
    rotation_y DOUBLE PRECISION,      -- Pitch (degrees/second)
    rotation_z DOUBLE PRECISION,      -- Yaw (degrees/second)

    -- Device metadata
    battery DOUBLE PRECISION,         -- 0-100% or voltage for Micro
    is_charging BOOLEAN,

    PRIMARY KEY (recorded_at, id)
);

-- Convert to hypertable
SELECT create_hypertable('telemetry', 'recorded_at');
```

#### Indexes

```sql
-- Device + time (for device-specific queries)
CREATE INDEX idx_telemetry_device_time
    ON telemetry (device_id, recorded_at DESC);

-- Session + time (for session-specific queries)
CREATE INDEX idx_telemetry_session 
    ON telemetry (session_id, recorded_at DESC)
    WHERE session_id IS NOT NULL;

-- Location (for geospatial queries)
CREATE INDEX idx_telemetry_location
    ON telemetry USING GIST(location);

-- Speed (for filtering by speed)
CREATE INDEX idx_telemetry_speed
    ON telemetry (speed)
    WHERE speed > 0;
```

### Sessions Table (Optional)

The `sessions` table groups telemetry data into logical sessions (e.g., track days, races).

```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,

    -- Metadata
    name VARCHAR(255),
    location VARCHAR(255),
    notes TEXT,

    -- Cached aggregates
    total_distance DOUBLE PRECISION,
    max_speed DOUBLE PRECISION,
    avg_speed DOUBLE PRECISION,
    max_g_force DOUBLE PRECISION,
    data_points_count BIGINT DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Data Lifecycle

### Compression

Data older than 7 days is automatically compressed to save storage space:

```sql
ALTER TABLE telemetry SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id'
);

SELECT add_compression_policy('telemetry', INTERVAL '7 days');
```

Compression typically achieves **10-20x reduction** in storage space.

### Retention Policy (Optional)

To automatically delete old data:

```sql
SELECT add_retention_policy('telemetry', INTERVAL '1 year');
```

This will delete data older than 1 year.

## Common Queries

### Recent Telemetry

```sql
-- Get last 100 telemetry points
SELECT * FROM telemetry
ORDER BY recorded_at DESC
LIMIT 100;
```

### Device-Specific Data

```sql
-- Get telemetry for a specific device
SELECT * FROM telemetry
WHERE device_id = 'device-001'
  AND recorded_at > NOW() - INTERVAL '1 hour'
ORDER BY recorded_at DESC;
```

### Session Data

```sql
-- Get all telemetry for a session
SELECT * FROM telemetry
WHERE session_id = '550e8400-e29b-41d4-a716-446655440000'
ORDER BY recorded_at ASC;
```

### Time Range Query

```sql
-- Get telemetry between two timestamps
SELECT * FROM telemetry
WHERE recorded_at BETWEEN '2025-01-01' AND '2025-01-02'
ORDER BY recorded_at ASC;
```

### Geospatial Queries

```sql
-- Find telemetry within a radius (10km from a point)
SELECT *
FROM telemetry
WHERE ST_DWithin(
    location,
    ST_MakePoint(23.2887238, 42.6719035)::geography,
    10000  -- meters
)
AND recorded_at > NOW() - INTERVAL '1 day';
```

```sql
-- Find telemetry within a bounding box
SELECT *
FROM telemetry
WHERE ST_Within(
    location::geometry,
    ST_MakeEnvelope(23.0, 42.0, 24.0, 43.0, 4326)
);
```

### Aggregations

```sql
-- Average speed per hour
SELECT
    time_bucket('1 hour', recorded_at) AS hour,
    AVG(speed) as avg_speed,
    MAX(speed) as max_speed,
    COUNT(*) as data_points
FROM telemetry
WHERE device_id = 'device-001'
  AND recorded_at > NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour;
```

```sql
-- Maximum g-forces per session
SELECT
    session_id,
    MAX(GREATEST(ABS(g_force_x), ABS(g_force_y), ABS(g_force_z))) as max_g_force,
    MAX(speed) as top_speed
FROM telemetry
WHERE session_id IS NOT NULL
GROUP BY session_id;
```

## Continuous Aggregates (Optional)

For frequently accessed aggregations, create continuous aggregates:

```sql
CREATE MATERIALIZED VIEW telemetry_hourly
WITH (timescaledb.continuous) AS
SELECT
    device_id,
    time_bucket('1 hour', recorded_at) AS hour,
    AVG(speed) as avg_speed,
    MAX(speed) as max_speed,
    AVG(battery) as avg_battery,
    COUNT(*) as data_points
FROM telemetry
GROUP BY device_id, hour;

-- Auto-refresh policy
SELECT add_continuous_aggregate_policy('telemetry_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');
```

## Performance Optimization

### Batch Inserts

For better performance, insert telemetry data in batches:

```go
// Good: Batch insert
repo.SaveBatch(ctx, telemetrySlice)

// Less efficient: Individual inserts
for _, t := range telemetrySlice {
    repo.Save(ctx, t)
}
```

### Query Optimization

1. **Always filter by time**: TimescaleDB is optimized for time-based queries
2. **Use indexes**: Device ID, session ID, and location are indexed
3. **Limit results**: Always use `LIMIT` to avoid large result sets
4. **Use time_bucket**: For aggregations, use `time_bucket` instead of `date_trunc`

### Connection Pooling

The service uses connection pooling with these defaults:

- Max connections: 25
- Max idle connections: 5
- Connection max lifetime: 5 minutes

Adjust via environment variables:

```bash
DB_MAX_CONNECTIONS=50
DB_MAX_IDLE_CONNECTIONS=10
DB_CONNECTION_MAX_LIFETIME=10m
```

## Monitoring

### Table Size

```sql
-- Uncompressed size
SELECT pg_size_pretty(pg_total_relation_size('telemetry'));

-- Compression ratio
SELECT *
FROM timescaledb_information.compressed_chunk_stats
WHERE hypertable_name = 'telemetry';
```

### Performance Metrics

```sql
-- Chunk count
SELECT COUNT(*) FROM timescaledb_information.chunks
WHERE hypertable_name = 'telemetry';

-- Compression stats
SELECT
    pg_size_pretty(before_compression_total_bytes) as before,
    pg_size_pretty(after_compression_total_bytes) as after,
    ROUND(100.0 * after_compression_total_bytes / before_compression_total_bytes, 2) as ratio_percent
FROM timescaledb_information.hypertable_compression_stats
WHERE hypertable_name = 'telemetry';
```

## Backup and Restore

### Backup

```bash
# Full database backup
pg_dump -h localhost -U telemetry_user telemetry_dev > backup.sql

# Compressed backup
pg_dump -h localhost -U telemetry_user telemetry_dev | gzip > backup.sql.gz

# Backup specific time range
pg_dump -h localhost -U telemetry_user \
    -t telemetry \
    --where="recorded_at >= '2025-01-01' AND recorded_at < '2025-02-01'" \
    telemetry_dev > january_backup.sql
```

### Restore

```bash
# Restore from backup
psql -h localhost -U telemetry_user telemetry_dev < backup.sql

# Restore from compressed backup
gunzip -c backup.sql.gz | psql -h localhost -U telemetry_user telemetry_dev
```

## Troubleshooting

### High Disk Usage

```sql
-- Check compression status
SELECT * FROM timescaledb_information.compression_settings
WHERE hypertable = 'telemetry';

-- Manually compress chunks
SELECT compress_chunk(i)
FROM show_chunks('telemetry', older_than => INTERVAL '7 days') i;
```

### Slow Queries

```sql
-- Enable query timing
\timing on

-- Analyze query plan
EXPLAIN ANALYZE
SELECT * FROM telemetry
WHERE device_id = 'device-001'
  AND recorded_at > NOW() - INTERVAL '1 hour';

-- Update statistics
ANALYZE telemetry;
```

### Connection Issues

```sql
-- Check active connections
SELECT count(*) FROM pg_stat_activity
WHERE datname = 'telemetry_dev';

-- Kill idle connections
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = 'telemetry_dev'
  AND state = 'idle'
  AND state_change < NOW() - INTERVAL '1 hour';
```

## References

- [TimescaleDB Documentation](https://docs.timescale.com/)
- [PostGIS Documentation](https://postgis.net/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
