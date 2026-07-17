CREATE TABLE IF NOT EXISTS idempotency_records (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id         UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    operation            VARCHAR(50) NOT NULL,
    key                  VARCHAR(255) NOT NULL,
    request_fingerprint  VARCHAR(64) NOT NULL,
    response_body        JSONB NOT NULL,
    status_code          INTEGER NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at           TIMESTAMPTZ NOT NULL,
    UNIQUE (workspace_id, operation, key)
);

CREATE INDEX idx_idempotency_records_lookup ON idempotency_records(workspace_id, operation, key);
CREATE INDEX idx_idempotency_records_expires ON idempotency_records(expires_at);

ALTER TABLE idempotency_records ENABLE ROW LEVEL SECURITY;

CREATE POLICY idempotency_records_tenant ON idempotency_records
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));
