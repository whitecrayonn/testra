-- Documents the new test_cases.source value used by the LLM-backed
-- Excel/CSV upload generator (apps/api/internal/testgen GenerateFromFile).
-- test_cases.source has no CHECK constraint (see migration 000042), so this
-- is a comment-only migration — the Go enum in testmanagement/domain.go is
-- the actual source of truth for allowed values ('manual', 'generated_spec',
-- 'generated_file').

COMMENT ON COLUMN test_cases.source IS
    'manual | generated_spec (deterministic, from an OpenAPI spec) | generated_file (LLM-backed, from an uploaded Excel/CSV via the ML service — see docs/BIBLICAL_TESTRA.md''s "No External LLM" exception)';
