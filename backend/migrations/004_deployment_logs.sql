-- ============================================
-- CloudX Migration 004: Deployment Logs Table
-- ============================================

-- Log levels for deployment logs
CREATE TYPE log_level AS ENUM (
    'INFO',
    'WARNING',
    'ERROR',
    'DEBUG'
);

-- deployment_logs stores streaming build and runtime logs per deployment
CREATE TABLE deployment_logs (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    deployment_id UUID NOT NULL,

    phase VARCHAR(50) NOT NULL, -- e.g. 'build', 'runtime'

    level log_level NOT NULL DEFAULT 'INFO',

    message TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_log_deployment
        FOREIGN KEY(deployment_id)
        REFERENCES deployments(id)
        ON DELETE CASCADE

);

-- Index for fast log fetching per deployment ordered by time
CREATE INDEX idx_deployment_logs_deployment_id ON deployment_logs(deployment_id, created_at DESC);
