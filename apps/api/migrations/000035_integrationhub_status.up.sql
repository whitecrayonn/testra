-- Add connection health and lifecycle status columns to integrations and retry tracking to integration_events.

ALTER TABLE integrations
    ADD COLUMN IF NOT EXISTS health_status VARCHAR(20) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS last_tested_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS last_error TEXT,
    ADD COLUMN IF NOT EXISTS sync_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0;

ALTER TABLE integration_events
    ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_integrations_health ON integrations(workspace_id, health_status);
CREATE INDEX IF NOT EXISTS idx_integration_events_status_retry ON integration_events(workspace_id, status, retry_count);
