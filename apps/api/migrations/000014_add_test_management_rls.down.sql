DROP POLICY IF EXISTS test_case_versions_tenant ON test_case_versions;
DROP POLICY IF EXISTS test_cases_tenant ON test_cases;
DROP POLICY IF EXISTS test_suites_tenant ON test_suites;
DROP POLICY IF EXISTS test_folders_tenant ON test_folders;

ALTER TABLE test_case_versions DISABLE ROW LEVEL SECURITY;
ALTER TABLE test_cases DISABLE ROW LEVEL SECURITY;
ALTER TABLE test_suites DISABLE ROW LEVEL SECURITY;
ALTER TABLE test_folders DISABLE ROW LEVEL SECURITY;
