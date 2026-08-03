-- Fix role_assignments RLS policy so it respects scope_type.
-- The original policy compared scope_id directly to the tenant organization id,
-- which is only correct for 'organization' scope assignments. Workspace and project
-- assignments have workspace/project UUIDs in scope_id and would leak or be hidden.

DROP POLICY IF EXISTS role_assignments_tenant ON role_assignments;

CREATE POLICY role_assignments_tenant ON role_assignments
    USING (
        (scope_type = 'organization' AND scope_id = current_setting('app.tenant_id', true)::uuid)
        OR (scope_type = 'workspace' AND scope_id IN (
            SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
        ))
        OR (scope_type = 'project' AND scope_id IN (
            SELECT id FROM projects WHERE workspace_id IN (
                SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
            )
        ))
    );
