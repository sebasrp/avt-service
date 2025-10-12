-- Sessions table for grouping telemetry data
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    name VARCHAR(255),
    location VARCHAR(255),
    notes TEXT,

    -- Cached aggregates (can be updated via triggers or application logic)
    total_distance DOUBLE PRECISION,
    max_speed DOUBLE PRECISION,
    avg_speed DOUBLE PRECISION,
    max_g_force DOUBLE PRECISION,
    data_points_count BIGINT DEFAULT 0,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_sessions_device ON sessions (device_id, started_at DESC);
CREATE INDEX idx_sessions_started_at ON sessions (started_at DESC);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_sessions_updated_at BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

