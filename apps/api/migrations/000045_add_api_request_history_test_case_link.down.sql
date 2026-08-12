DROP INDEX IF EXISTS idx_api_request_history_test_case;

ALTER TABLE api_request_history DROP COLUMN IF EXISTS test_case_id;
