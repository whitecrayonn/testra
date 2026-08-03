DROP POLICY IF EXISTS generation_runs_tenant ON generation_runs;

DROP INDEX IF EXISTS idx_test_cases_generation_run;
DROP INDEX IF EXISTS idx_test_cases_source;

ALTER TABLE test_cases DROP COLUMN IF EXISTS reviewed_by;
ALTER TABLE test_cases DROP COLUMN IF EXISTS generation_run_id;
ALTER TABLE test_cases DROP COLUMN IF EXISTS source;

DROP TABLE IF EXISTS generation_runs;
