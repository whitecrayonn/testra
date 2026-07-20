-- Automation Hub: projects, executions, artifacts, logs, and permissions.

CREATE TABLE IF NOT EXISTS automation_projects (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    framework VARCHAR(50),
    repository_url TEXT,
    branch VARCHAR(255),
    command TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_automation_projects_workspace_id
    ON automation_projects(workspace_id);
CREATE INDEX IF NOT EXISTS idx_automation_projects_project_id
    ON automation_projects(project_id);

CREATE TABLE IF NOT EXISTS automation_executions (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES automation_projects(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    test_run_id UUID REFERENCES test_runs(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    report_format VARCHAR(50),
    report_path TEXT,
    retry_of UUID REFERENCES automation_executions(id) ON DELETE SET NULL,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    total INTEGER NOT NULL DEFAULT 0,
    passed INTEGER NOT NULL DEFAULT 0,
    failed INTEGER NOT NULL DEFAULT 0,
    skipped INTEGER NOT NULL DEFAULT 0,
    blocked INTEGER NOT NULL DEFAULT 0,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    triggered_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_automation_executions_workspace_id
    ON automation_executions(workspace_id);
CREATE INDEX IF NOT EXISTS idx_automation_executions_project_id
    ON automation_executions(project_id);
CREATE INDEX IF NOT EXISTS idx_automation_executions_test_run_id
    ON automation_executions(test_run_id);
CREATE INDEX IF NOT EXISTS idx_automation_executions_status
    ON automation_executions(status);
CREATE INDEX IF NOT EXISTS idx_automation_executions_created_at
    ON automation_executions(workspace_id, created_at DESC);

CREATE TABLE IF NOT EXISTS automation_artifacts (
    id UUID PRIMARY KEY,
    execution_id UUID NOT NULL REFERENCES automation_executions(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    test_run_item_id UUID REFERENCES test_run_items(id) ON DELETE SET NULL,
    kind VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    mime_type VARCHAR(255),
    file_size BIGINT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_automation_artifacts_workspace_id
    ON automation_artifacts(workspace_id);
CREATE INDEX IF NOT EXISTS idx_automation_artifacts_execution_id
    ON automation_artifacts(execution_id);
CREATE INDEX IF NOT EXISTS idx_automation_artifacts_kind
    ON automation_artifacts(kind);
CREATE INDEX IF NOT EXISTS idx_automation_artifacts_test_run_item_id
    ON automation_artifacts(test_run_item_id);

CREATE TABLE IF NOT EXISTS automation_logs (
    id UUID PRIMARY KEY,
    execution_id UUID NOT NULL REFERENCES automation_executions(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL DEFAULT 'info',
    message TEXT NOT NULL,
    logged_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_automation_logs_workspace_id
    ON automation_logs(workspace_id);
CREATE INDEX IF NOT EXISTS idx_automation_logs_execution_id
    ON automation_logs(execution_id);
CREATE INDEX IF NOT EXISTS idx_automation_logs_logged_at
    ON automation_logs(execution_id, logged_at);

-- Row-level security
ALTER TABLE automation_projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_artifacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY automation_projects_tenant ON automation_projects
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY automation_executions_tenant ON automation_executions
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY automation_artifacts_tenant ON automation_artifacts
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY automation_logs_tenant ON automation_logs
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

-- Lookup policies for tenant resolution before app.tenant_id is known.
CREATE POLICY automation_projects_lookup_user ON automation_projects
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

CREATE POLICY automation_executions_lookup_user ON automation_executions
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

CREATE POLICY automation_artifacts_lookup_user ON automation_artifacts
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

CREATE POLICY automation_logs_lookup_user ON automation_logs
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
        )
    ));

-- Permissions
INSERT INTO permissions (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000002101', 'automation:read', 'View automation projects, executions, artifacts, and logs'),
    ('00000000-0000-0000-0000-000000002102', 'automation:create', 'Create automation projects and executions'),
    ('00000000-0000-0000-0000-000000002103', 'automation:update', 'Update automation projects and executions'),
    ('00000000-0000-0000-0000-000000002104', 'automation:delete', 'Delete automation projects and executions'),
    ('00000000-0000-0000-0000-000000002105', 'automation:execute', 'Import reports, rerun executions, and upload artifacts')
ON CONFLICT (name) DO NOTHING;

-- Owner: all automation permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002101'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002102'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002103'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002104'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002105')
ON CONFLICT DO NOTHING;

-- Admin: all automation permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002101'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002102'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002103'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002104'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002105')
ON CONFLICT DO NOTHING;

-- QA Engineer: create/read/update/execute
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000002101'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000002102'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000002103'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000002105')
ON CONFLICT DO NOTHING;

-- Viewer: read only
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000002101')
ON CONFLICT DO NOTHING;
