DROP POLICY IF EXISTS test_run_items_tenant ON test_run_items;
DROP POLICY IF EXISTS test_runs_tenant ON test_runs;

ALTER TABLE test_run_items DISABLE ROW LEVEL SECURITY;
ALTER TABLE test_runs DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS test_run_items;
DROP TABLE IF EXISTS test_runs;
