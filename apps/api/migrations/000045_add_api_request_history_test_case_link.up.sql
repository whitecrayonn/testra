-- Lets an API-testing execution record which test case it was run for, so
-- the "quick generate" flow (apps/api/internal/testgen) can show a test
-- case's most recent ad hoc execution result. Mirrors the existing
-- test_run_items.test_case_id link (migration 000015) but for the
-- apitesting module's own history table rather than automation-hub-ingested
-- runs.

ALTER TABLE api_request_history ADD COLUMN IF NOT EXISTS test_case_id UUID REFERENCES test_cases(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_api_request_history_test_case ON api_request_history(test_case_id);
