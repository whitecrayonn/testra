DROP POLICY IF EXISTS role_assignments_tenant ON role_assignments;
DROP POLICY IF EXISTS api_keys_tenant ON api_keys;
DROP POLICY IF EXISTS projects_tenant ON projects;
DROP POLICY IF EXISTS workspace_members_tenant ON workspace_members;
DROP POLICY IF EXISTS workspaces_tenant ON workspaces;
DROP POLICY IF EXISTS org_members_tenant ON organization_members;
DROP POLICY IF EXISTS org_tenant_isolation ON organizations;

ALTER TABLE role_assignments DISABLE ROW LEVEL SECURITY;
ALTER TABLE api_keys DISABLE ROW LEVEL SECURITY;
ALTER TABLE projects DISABLE ROW LEVEL SECURITY;
ALTER TABLE workspace_members DISABLE ROW LEVEL SECURITY;
ALTER TABLE workspaces DISABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members DISABLE ROW LEVEL SECURITY;
ALTER TABLE organizations DISABLE ROW LEVEL SECURITY;
