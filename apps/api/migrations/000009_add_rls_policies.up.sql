-- Enable RLS on all tenant-scoped tables
-- Per ADR-004: tenant_id (organization_id) isolation with RLS policies

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspace_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_assignments ENABLE ROW LEVEL SECURITY;

-- Create policies: app.tenant_id is set per-transaction by the application layer.
-- The application sets app.tenant_id after authenticated scope resolution.
-- API roles do NOT bypass RLS (SUPERUSER not used by the app).

-- Organizations: only visible if tenant_id matches owner_id's org membership
CREATE POLICY org_tenant_isolation ON organizations
    USING (id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY org_members_tenant ON organization_members
    USING (organization_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY workspaces_tenant ON workspaces
    USING (organization_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY workspace_members_tenant ON workspace_members
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY projects_tenant ON projects
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY api_keys_tenant ON api_keys
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY role_assignments_tenant ON role_assignments
    USING (scope_id = current_setting('app.tenant_id', true)::uuid);

-- Users table: not tenant-scoped (users exist across orgs), no RLS needed
-- password_reset_tokens: not tenant-scoped, no RLS needed
-- roles, permissions, role_permissions: system tables, no RLS needed
