-- Notification Center schema, permissions and tenant isolation.
-- These tables support in-app notifications, per-user preferences, and
-- workspace-scoped notification channels (email, slack, teams, webhook).

CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            VARCHAR(50) NOT NULL DEFAULT 'system',
    title           VARCHAR(255) NOT NULL,
    body            TEXT DEFAULT '',
    link            VARCHAR(500) DEFAULT '',
    read            BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_org ON notifications(user_id, organization_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications(user_id, read, created_at DESC);

CREATE TABLE IF NOT EXISTS notification_preferences (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    in_app_enabled   BOOLEAN NOT NULL DEFAULT true,
    email_enabled    BOOLEAN NOT NULL DEFAULT false,
    slack_enabled    BOOLEAN NOT NULL DEFAULT false,
    teams_enabled    BOOLEAN NOT NULL DEFAULT false,
    webhook_enabled  BOOLEAN NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

CREATE TABLE IF NOT EXISTS notification_channels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    type            VARCHAR(50) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    config          JSONB NOT NULL DEFAULT '{}',
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_channels_workspace ON notification_channels(workspace_id);

-- Permissions for the notification center
INSERT INTO permissions (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000001301', 'notifications:read', 'View notifications'),
    ('00000000-0000-0000-0000-000000001302', 'notifications:create', 'Create notifications'),
    ('00000000-0000-0000-0000-000000001303', 'notifications:update', 'Update notifications'),
    ('00000000-0000-0000-0000-000000001304', 'notifications:delete', 'Delete notifications'),
    ('00000000-0000-0000-0000-000000001305', 'notification_preferences:read', 'View notification preferences'),
    ('00000000-0000-0000-0000-000000001306', 'notification_preferences:update', 'Update notification preferences'),
    ('00000000-0000-0000-0000-000000001307', 'notification_channels:read', 'View notification channels'),
    ('00000000-0000-0000-0000-000000001308', 'notification_channels:create', 'Create notification channels'),
    ('00000000-0000-0000-0000-000000001309', 'notification_channels:update', 'Update notification channels'),
    ('00000000-0000-0000-0000-000000001310', 'notification_channels:delete', 'Delete notification channels')
ON CONFLICT (name) DO NOTHING;

-- Owner gets full notification access
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001301'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001302'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001303'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001304'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001305'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001306'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001307'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001308'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001309'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000001310')
ON CONFLICT DO NOTHING;

-- Admin gets full notification access
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001301'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001302'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001303'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001304'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001305'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001306'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001307'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001308'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001309'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000001310')
ON CONFLICT DO NOTHING;

-- QA Engineer can view notifications, preferences and channels, and update their own notifications/preferences
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000001301'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000001303'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000001305'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000001306'),
    ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000001307')
ON CONFLICT DO NOTHING;

-- Viewer can view notifications, preferences and channels
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000001301'),
    ('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000001305'),
    ('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000001307')
ON CONFLICT DO NOTHING;

-- Tenant isolation (RLS)
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_channels ENABLE ROW LEVEL SECURITY;

CREATE POLICY notifications_tenant ON notifications
    USING (organization_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY notification_preferences_tenant ON notification_preferences
    USING (organization_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY notification_channels_tenant ON notification_channels
    USING (organization_id = current_setting('app.tenant_id', true)::uuid);
