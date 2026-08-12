-- Revert role_assignments RLS policy to the original tenant-only match on scope_id.
-- This is less precise for workspace/project scopes but is the prior behavior.

DROP POLICY IF EXISTS role_assignments_tenant ON role_assignments;

CREATE POLICY role_assignments_tenant ON role_assignments
    USING (scope_id = current_setting('app.tenant_id', true)::uuid);
