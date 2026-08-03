CREATE TABLE IF NOT EXISTS flaky_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    test_case_id UUID REFERENCES test_cases(id) ON DELETE CASCADE,
    test_case_title VARCHAR(255) NOT NULL DEFAULT '',
    flakiness_score NUMERIC(5,4) NOT NULL DEFAULT 0,
    confidence NUMERIC(5,4) NOT NULL DEFAULT 0,
    features JSONB NOT NULL DEFAULT '{}',
    predicted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flaky_predictions_workspace ON flaky_predictions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_flaky_predictions_score ON flaky_predictions(workspace_id, flakiness_score DESC);

CREATE TABLE IF NOT EXISTS failure_clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    cluster_label VARCHAR(255) NOT NULL,
    pattern VARCHAR(255) NOT NULL DEFAULT '',
    sample_error TEXT DEFAULT '',
    count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_failure_clusters_workspace ON failure_clusters(workspace_id);
CREATE INDEX IF NOT EXISTS idx_failure_clusters_label ON failure_clusters(workspace_id, cluster_label);

ALTER TABLE flaky_predictions ENABLE ROW LEVEL SECURITY;
ALTER TABLE failure_clusters ENABLE ROW LEVEL SECURITY;

CREATE POLICY flaky_predictions_tenant ON flaky_predictions
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY failure_clusters_tenant ON failure_clusters
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));
