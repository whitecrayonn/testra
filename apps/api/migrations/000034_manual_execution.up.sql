-- Adds the schema needed for manual, step-by-step test execution and test plans.

-- Step-level execution state on existing test run items.
ALTER TABLE test_run_items
    ADD COLUMN IF NOT EXISTS step_results JSONB DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS executed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS executed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS comment TEXT DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_test_run_items_executed_by ON test_run_items(executed_by);

-- Snapshots of a run item each time execution is saved.
CREATE TABLE IF NOT EXISTS test_run_item_history (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_item_id   UUID NOT NULL REFERENCES test_run_items(id) ON DELETE CASCADE,
    status        VARCHAR(20) NOT NULL,
    step_results  JSONB DEFAULT '[]',
    comment       TEXT DEFAULT '',
    duration_ms   BIGINT NOT NULL DEFAULT 0,
    executed_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_test_run_item_history_run_item ON test_run_item_history(run_item_id);
CREATE INDEX IF NOT EXISTS idx_test_run_item_history_created_at ON test_run_item_history(run_item_id, created_at DESC);

-- Evidence metadata attached to run items (and optionally a step).
CREATE TABLE IF NOT EXISTS test_run_item_evidence (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_item_id   UUID NOT NULL REFERENCES test_run_items(id) ON DELETE CASCADE,
    step_order    INTEGER NOT NULL DEFAULT 0,
    file_name     VARCHAR(500) NOT NULL DEFAULT '',
    content_type  VARCHAR(255) NOT NULL DEFAULT '',
    storage_path  VARCHAR(1000) NOT NULL DEFAULT '',
    uploaded_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_test_run_item_evidence_run_item ON test_run_item_evidence(run_item_id);

-- Link defects to run items (failed steps).
CREATE TABLE IF NOT EXISTS test_run_item_defects (
    run_item_id   UUID NOT NULL REFERENCES test_run_items(id) ON DELETE CASCADE,
    defect_id     UUID NOT NULL REFERENCES defects(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (run_item_id, defect_id)
);

CREATE INDEX IF NOT EXISTS idx_test_run_item_defects_defect ON test_run_item_defects(defect_id);

-- Test plans: reusable collections of test cases that can be turned into runs.
CREATE TABLE IF NOT EXISTS test_plans (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    suite_id      UUID REFERENCES test_suites(id) ON DELETE SET NULL,
    name          VARCHAR(255) NOT NULL,
    description   TEXT DEFAULT '',
    status        VARCHAR(20) NOT NULL DEFAULT 'active',
    configuration JSONB DEFAULT '{}',
    created_by    UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_test_plans_workspace ON test_plans(workspace_id);
CREATE INDEX IF NOT EXISTS idx_test_plans_project ON test_plans(project_id);
CREATE INDEX IF NOT EXISTS idx_test_plans_suite ON test_plans(suite_id);
CREATE INDEX IF NOT EXISTS idx_test_plans_status ON test_plans(status);
CREATE INDEX IF NOT EXISTS idx_test_plans_created_by ON test_plans(created_by);

CREATE TABLE IF NOT EXISTS test_plan_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id       UUID NOT NULL REFERENCES test_plans(id) ON DELETE CASCADE,
    test_case_id  UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    sort_order    INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_test_plan_items_plan ON test_plan_items(plan_id);
CREATE INDEX IF NOT EXISTS idx_test_plan_items_case ON test_plan_items(test_case_id);
CREATE INDEX IF NOT EXISTS idx_test_plan_items_sort ON test_plan_items(plan_id, sort_order);

-- Row-level security for new tables. They live in the same tenant context as test_runs.
ALTER TABLE test_run_item_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_run_item_evidence ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_run_item_defects ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_plans ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_plan_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY test_run_item_history_tenant ON test_run_item_history
    USING (run_item_id IN (
        SELECT id FROM test_run_items WHERE run_id IN (
            SELECT id FROM test_runs WHERE workspace_id IN (
                SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
            )
        )
    ));

CREATE POLICY test_run_item_evidence_tenant ON test_run_item_evidence
    USING (run_item_id IN (
        SELECT id FROM test_run_items WHERE run_id IN (
            SELECT id FROM test_runs WHERE workspace_id IN (
                SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
            )
        )
    ));

CREATE POLICY test_run_item_defects_tenant ON test_run_item_defects
    USING (run_item_id IN (
        SELECT id FROM test_run_items WHERE run_id IN (
            SELECT id FROM test_runs WHERE workspace_id IN (
                SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
            )
        )
    ));

CREATE POLICY test_plans_tenant ON test_plans
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY test_plan_items_tenant ON test_plan_items
    USING (plan_id IN (
        SELECT id FROM test_plans WHERE workspace_id IN (
            SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
        )
    ));
