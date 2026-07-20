-- Revert idempotency scope to workspace-only. Non-null workspace rows are preserved;
-- organization-only rows remain orphaned until cleaned up manually.
ALTER TABLE idempotency_records
DROP CONSTRAINT IF EXISTS idempotency_records_org_workspace_operation_key_key;

ALTER TABLE idempotency_records
ALTER COLUMN workspace_id SET NOT NULL,
DROP COLUMN IF EXISTS organization_id;

ALTER TABLE idempotency_records
ADD CONSTRAINT idempotency_records_workspace_id_operation_key_key
UNIQUE (workspace_id, operation, key);

DROP POLICY IF EXISTS idempotency_records_tenant ON idempotency_records;

CREATE POLICY idempotency_records_tenant ON idempotency_records
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));
