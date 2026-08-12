-- Performance and operational indexes for high-traffic tables.
-- Addresses SBL-079 (missing indexes) and SBL-083 (queue_jobs dequeue composite index).

CREATE INDEX IF NOT EXISTS idx_audit_events_resource_created
    ON audit_events(resource, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_revoked
    ON refresh_tokens(user_id, revoked_at);

CREATE INDEX IF NOT EXISTS idx_notification_channels_org
    ON notification_channels(organization_id);

CREATE INDEX IF NOT EXISTS idx_test_cases_workspace_status
    ON test_cases(workspace_id, status);

-- Dequeue-optimized composite index for queue_jobs (SBL-083).
CREATE INDEX IF NOT EXISTS idx_queue_jobs_dequeue
    ON queue_jobs(queue_name, status, scheduled_at, created_at);
