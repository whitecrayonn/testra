-- Widen idempotency scope to organization so it can be applied to any mutating
-- endpoint, not only workspace-scoped create calls.
ALTER TABLE idempotency_records
ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Backfill organization_id from the existing workspace relationship.
UPDATE idempotency_records
SET organization_id = (SELECT organization_id FROM workspaces WHERE id = workspace_id)
WHERE organization_id IS NULL AND workspace_id IS NOT NULL;

ALTER TABLE idempotency_records
ALTER COLUMN organization_id SET NOT NULL,
ALTER COLUMN workspace_id DROP NOT NULL;

-- Replace the workspace-scoped unique key with an org+workspace scoped key.
ALTER TABLE idempotency_records
DROP CONSTRAINT IF EXISTS idempotency_records_workspace_id_operation_key_key;

ALTER TABLE idempotency_records
ADD CONSTRAINT idempotency_records_org_workspace_operation_key_key
UNIQUE (organization_id, workspace_id, operation, key);

-- Replace RLS with organization ownership.
DROP POLICY IF EXISTS idempotency_records_tenant ON idempotency_records;

CREATE POLICY idempotency_records_tenant ON idempotency_records
    USING (organization_id = current_setting('app.tenant_id', true)::uuid);
