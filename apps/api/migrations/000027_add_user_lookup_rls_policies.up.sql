-- ADR-013: enable authenticated-user lookup before tenant_id is known.
-- These permissive OR policies allow a connection to resolve the tenant
-- of a resource that belongs to an organization the current user is a
-- member of. The application MUST reset app.lookup_user_id after tenant
-- resolution is complete.

CREATE INDEX IF NOT EXISTS idx_organization_members_user_id ON organization_members(user_id);

-- organization_members needs a lookup policy so subqueries below can see
-- the user's membership rows while app.tenant_id is not yet set.
DROP POLICY IF EXISTS organization_members_lookup_user ON organization_members;
CREATE POLICY organization_members_lookup_user ON organization_members
    USING (user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid);

-- workspaces
DROP POLICY IF EXISTS workspaces_lookup_user ON workspaces;
CREATE POLICY workspaces_lookup_user ON workspaces
    USING (organization_id IN (
        SELECT organization_id FROM organization_members
        WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
    ));

-- projects
DROP POLICY IF EXISTS projects_lookup_user ON projects;
CREATE POLICY projects_lookup_user ON projects
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

-- api_keys
DROP POLICY IF EXISTS api_keys_lookup_user ON api_keys;
CREATE POLICY api_keys_lookup_user ON api_keys
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

-- test management tables
DROP POLICY IF EXISTS test_folders_lookup_user ON test_folders;
CREATE POLICY test_folders_lookup_user ON test_folders
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

DROP POLICY IF EXISTS test_suites_lookup_user ON test_suites;
CREATE POLICY test_suites_lookup_user ON test_suites
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

DROP POLICY IF EXISTS test_cases_lookup_user ON test_cases;
CREATE POLICY test_cases_lookup_user ON test_cases
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

DROP POLICY IF EXISTS test_case_versions_lookup_user ON test_case_versions;
CREATE POLICY test_case_versions_lookup_user ON test_case_versions
    USING (test_case_id IN (
        SELECT id FROM test_cases WHERE workspace_id IN (
            SELECT id FROM workspaces WHERE organization_id IN (
                SELECT organization_id FROM organization_members
                WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
            )
        )
    ));

-- test runs
DROP POLICY IF EXISTS test_runs_lookup_user ON test_runs;
CREATE POLICY test_runs_lookup_user ON test_runs
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

DROP POLICY IF EXISTS test_run_items_lookup_user ON test_run_items;
CREATE POLICY test_run_items_lookup_user ON test_run_items
    USING (run_id IN (
        SELECT id FROM test_runs WHERE workspace_id IN (
            SELECT id FROM workspaces WHERE organization_id IN (
                SELECT organization_id FROM organization_members
                WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
            )
        )
    ));

-- defects
DROP POLICY IF EXISTS defects_lookup_user ON defects;
CREATE POLICY defects_lookup_user ON defects
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));
