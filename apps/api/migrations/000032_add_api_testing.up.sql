-- API Testing Engine: collections, folders, environments, requests, and execution history.

CREATE TABLE IF NOT EXISTS api_collections (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_collections_workspace_id
    ON api_collections(workspace_id);

CREATE TABLE IF NOT EXISTS api_folders (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    collection_id UUID NOT NULL REFERENCES api_collections(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES api_folders(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_folders_workspace_id
    ON api_folders(workspace_id);
CREATE INDEX IF NOT EXISTS idx_api_folders_collection_id
    ON api_folders(collection_id);
CREATE INDEX IF NOT EXISTS idx_api_folders_parent_id
    ON api_folders(parent_id);

CREATE TABLE IF NOT EXISTS api_environments (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_environments_workspace_id
    ON api_environments(workspace_id);

CREATE TABLE IF NOT EXISTS api_requests (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    collection_id UUID NOT NULL REFERENCES api_collections(id) ON DELETE CASCADE,
    folder_id UUID REFERENCES api_folders(id) ON DELETE SET NULL,
    environment_id UUID REFERENCES api_environments(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    method VARCHAR(20) NOT NULL DEFAULT 'GET',
    url TEXT NOT NULL DEFAULT '',
    headers JSONB NOT NULL DEFAULT '[]'::jsonb,
    query_params JSONB NOT NULL DEFAULT '[]'::jsonb,
    auth_type VARCHAR(50) NOT NULL DEFAULT 'none',
    auth_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    body_type VARCHAR(50) NOT NULL DEFAULT 'none',
    body_content TEXT,
    variables JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_requests_workspace_id
    ON api_requests(workspace_id);
CREATE INDEX IF NOT EXISTS idx_api_requests_collection_id
    ON api_requests(collection_id);
CREATE INDEX IF NOT EXISTS idx_api_requests_folder_id
    ON api_requests(folder_id);

CREATE TABLE IF NOT EXISTS api_request_history (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    request_id UUID REFERENCES api_requests(id) ON DELETE SET NULL,
    environment_id UUID REFERENCES api_environments(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    method VARCHAR(20) NOT NULL,
    url TEXT NOT NULL,
    request_headers JSONB NOT NULL DEFAULT '[]'::jsonb,
    request_body TEXT,
    response_status INTEGER,
    response_status_text VARCHAR(100),
    response_headers JSONB NOT NULL DEFAULT '[]'::jsonb,
    response_body TEXT,
    response_time_ms INTEGER,
    error TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_request_history_workspace_id
    ON api_request_history(workspace_id);
CREATE INDEX IF NOT EXISTS idx_api_request_history_request_id
    ON api_request_history(request_id);
CREATE INDEX IF NOT EXISTS idx_api_request_history_created_at
    ON api_request_history(workspace_id, created_at DESC);

-- RLS policies
ALTER TABLE api_collections ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_folders ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_environments ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE api_request_history ENABLE ROW LEVEL SECURITY;

CREATE POLICY api_collections_tenant ON api_collections
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY api_folders_tenant ON api_folders
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY api_environments_tenant ON api_environments
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY api_requests_tenant ON api_requests
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

CREATE POLICY api_request_history_tenant ON api_request_history
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));

-- Permissions
INSERT INTO permissions (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000002001', 'api_tests:read', 'View API collections, requests, environments, and execution history'),
    ('00000000-0000-0000-0000-000000002002', 'api_tests:create', 'Create API collections, requests, folders, and environments'),
    ('00000000-0000-0000-0000-000000002003', 'api_tests:update', 'Update API collections, requests, folders, and environments'),
    ('00000000-0000-0000-0000-000000002004', 'api_tests:delete', 'Delete API collections, requests, folders, and environments'),
    ('00000000-0000-0000-0000-000000002005', 'api_tests:execute', 'Execute API requests and view response details')
ON CONFLICT (name) DO NOTHING;

-- Owner: all api_tests permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002001'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002002'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002003'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002004'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000002005')
ON CONFLICT DO NOTHING;

-- Admin: all api_tests permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002001'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002002'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002003'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002004'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000002005')
ON CONFLICT DO NOTHING;

-- QA Engineer: full access except delete
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000002001'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000002002'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000002003'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000002005')
ON CONFLICT DO NOTHING;

-- Viewer: read only
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000002001')
ON CONFLICT DO NOTHING;
