CREATE TABLE IF NOT EXISTS test_runs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    suite_id      UUID REFERENCES test_suites(id) ON DELETE SET NULL,
    name          VARCHAR(255) NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    total         INTEGER NOT NULL DEFAULT 0,
    passed        INTEGER NOT NULL DEFAULT 0,
    failed        INTEGER NOT NULL DEFAULT 0,
    skipped       INTEGER NOT NULL DEFAULT 0,
    blocked       INTEGER NOT NULL DEFAULT 0,
    duration_ms   BIGINT NOT NULL DEFAULT 0,
    source        VARCHAR(20) NOT NULL DEFAULT 'manual',
    metadata      JSONB DEFAULT '{}',
    created_by    UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_runs_workspace ON test_runs(workspace_id);
CREATE INDEX idx_test_runs_project ON test_runs(project_id);
CREATE INDEX idx_test_runs_suite ON test_runs(suite_id);
CREATE INDEX idx_test_runs_status ON test_runs(status);
CREATE INDEX idx_test_runs_created_by ON test_runs(created_by);
CREATE INDEX idx_test_runs_created_at ON test_runs(created_at DESC);

CREATE TABLE IF NOT EXISTS test_run_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id        UUID NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
    test_case_id  UUID REFERENCES test_cases(id) ON DELETE SET NULL,
    title         VARCHAR(500) NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    duration_ms   BIGINT NOT NULL DEFAULT 0,
    error_message TEXT DEFAULT '',
    stack_trace   TEXT DEFAULT '',
    artifacts     JSONB DEFAULT '[]',
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_run_items_run ON test_run_items(run_id);
CREATE INDEX idx_test_run_items_case ON test_run_items(test_case_id);
CREATE INDEX idx_test_run_items_status ON test_run_items(status);
CREATE INDEX idx_test_run_items_sort ON test_run_items(run_id, sort_order);

-- Enable RLS consistent with migration 000009 and 000014
ALTER TABLE test_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_run_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY test_runs_tenant ON test_runs
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY test_run_items_tenant ON test_run_items
    USING (run_id IN (
        SELECT id FROM test_runs WHERE workspace_id IN (
            SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
        )
    ));
