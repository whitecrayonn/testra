-- Deterministic (no LLM) test case generation from an imported OpenAPI spec.
-- Per docs/BIBLICAL_TESTRA.md "No External LLM" / "Transparent ML" principles and
-- docs/AI_MEMORY.md, generation must be rule-based and explainable, not model-based.
--
-- generation_runs records one generation job (which spec, how many endpoints/cases,
-- who ran it). test_cases gains source/generation_run_id/reviewed_by so generated
-- cases share the same table and API as manually created ones, distinguished only
-- by these columns, and require explicit human approval before status = 'active'.

CREATE TABLE IF NOT EXISTS generation_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source          VARCHAR(20) NOT NULL DEFAULT 'generated_spec',
    spec_filename   VARCHAR(500) DEFAULT '',
    endpoint_count  INTEGER NOT NULL DEFAULT 0,
    case_count      INTEGER NOT NULL DEFAULT 0,
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_generation_runs_workspace ON generation_runs(workspace_id);
CREATE INDEX idx_generation_runs_project ON generation_runs(project_id);

ALTER TABLE test_cases ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'manual';
ALTER TABLE test_cases ADD COLUMN IF NOT EXISTS generation_run_id UUID REFERENCES generation_runs(id) ON DELETE SET NULL;
ALTER TABLE test_cases ADD COLUMN IF NOT EXISTS reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX idx_test_cases_source ON test_cases(source);
CREATE INDEX idx_test_cases_generation_run ON test_cases(generation_run_id);

-- Tenant isolation, consistent with migration 000014's model.
ALTER TABLE generation_runs ENABLE ROW LEVEL SECURITY;

CREATE POLICY generation_runs_tenant ON generation_runs
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));
