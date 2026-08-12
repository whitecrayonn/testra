-- Remove ADR-013 authenticated-user lookup policies.

DROP POLICY IF EXISTS defects_lookup_user ON defects;
DROP POLICY IF EXISTS test_run_items_lookup_user ON test_run_items;
DROP POLICY IF EXISTS test_runs_lookup_user ON test_runs;
DROP POLICY IF EXISTS test_case_versions_lookup_user ON test_case_versions;
DROP POLICY IF EXISTS test_cases_lookup_user ON test_cases;
DROP POLICY IF EXISTS test_suites_lookup_user ON test_suites;
DROP POLICY IF EXISTS test_folders_lookup_user ON test_folders;
DROP POLICY IF EXISTS api_keys_lookup_user ON api_keys;
DROP POLICY IF EXISTS projects_lookup_user ON projects;
DROP POLICY IF EXISTS workspaces_lookup_user ON workspaces;
DROP POLICY IF EXISTS organization_members_lookup_user ON organization_members;

DROP INDEX IF EXISTS idx_organization_members_user_id;
