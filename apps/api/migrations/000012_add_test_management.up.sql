CREATE TABLE IF NOT EXISTS test_folders (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    parent_id    UUID REFERENCES test_folders(id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_folders_workspace ON test_folders(workspace_id);
CREATE INDEX idx_test_folders_parent ON test_folders(parent_id);

CREATE TABLE IF NOT EXISTS test_suites (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    folder_id    UUID REFERENCES test_folders(id) ON DELETE SET NULL,
    name         VARCHAR(255) NOT NULL,
    description  TEXT DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_suites_workspace ON test_suites(workspace_id);
CREATE INDEX idx_test_suites_folder ON test_suites(folder_id);

CREATE TABLE IF NOT EXISTS test_cases (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id    UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    suite_id      UUID REFERENCES test_suites(id) ON DELETE SET NULL,
    title         VARCHAR(500) NOT NULL,
    description   TEXT DEFAULT '',
    preconditions TEXT DEFAULT '',
    steps         JSONB DEFAULT '[]',
    status        VARCHAR(20) NOT NULL DEFAULT 'draft',
    priority      VARCHAR(20) NOT NULL DEFAULT 'medium',
    tags          TEXT[] DEFAULT '{}',
    version       INTEGER NOT NULL DEFAULT 1,
    created_by    UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    search_tsv    TSVECTOR
);

CREATE INDEX idx_test_cases_workspace ON test_cases(workspace_id);
CREATE INDEX idx_test_cases_project ON test_cases(project_id);
CREATE INDEX idx_test_cases_suite ON test_cases(suite_id);
CREATE INDEX idx_test_cases_status ON test_cases(status);
CREATE INDEX idx_test_cases_search ON test_cases USING GIN(search_tsv);

CREATE TRIGGER test_cases_search_tsv_insert
    AFTER INSERT ON test_cases
    FOR EACH ROW EXECUTE FUNCTION
    tsvector_update_trigger(search_tsv, 'pg_catalog.english', title, description);

CREATE TRIGGER test_cases_search_tsv_update
    AFTER UPDATE OF title, description ON test_cases
    FOR EACH ROW EXECUTE FUNCTION
    tsvector_update_trigger(search_tsv, 'pg_catalog.english', title, description);

CREATE TABLE IF NOT EXISTS test_case_versions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_case_id  UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    version       INTEGER NOT NULL,
    title         VARCHAR(500) NOT NULL,
    description   TEXT DEFAULT '',
    preconditions TEXT DEFAULT '',
    steps         JSONB DEFAULT '[]',
    changed_by    UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_case_versions_case ON test_case_versions(test_case_id);
CREATE INDEX idx_test_case_versions_version ON test_case_versions(test_case_id, version DESC);
