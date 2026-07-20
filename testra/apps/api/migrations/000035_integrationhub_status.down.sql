-- Revert integration hub status columns.

DROP INDEX IF EXISTS idx_integration_events_status_retry;
DROP INDEX IF EXISTS idx_integrations_health;

ALTER TABLE integration_events
    DROP COLUMN IF EXISTS retry_count;

ALTER TABLE integrations
    DROP COLUMN IF EXISTS retry_count,
    DROP COLUMN IF EXISTS sync_status,
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS last_tested_at,
    DROP COLUMN IF EXISTS health_status;
