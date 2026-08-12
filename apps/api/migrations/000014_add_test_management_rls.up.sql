-- Enable RLS on all test management tables
-- Consistent with migration 000009 tenant isolation model:
--   app.tenant_id is set per-transaction by the application layer
--   tables are scoped through workspace_id → workspaces.organization_id

ALTER TABLE test_folders ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_suites ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_cases ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_case_versions ENABLE ROW LEVEL SECURITY;

CREATE POLICY test_folders_tenant ON test_folders
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY test_suites_tenant ON test_suites
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY test_cases_tenant ON test_cases
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY test_case_versions_tenant ON test_case_versions
    USING (test_case_id IN (
        SELECT id FROM test_cases WHERE workspace_id IN (
            SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
        )
    ));
