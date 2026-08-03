-- Notification templates and delivery history for the Notification Center.

CREATE TABLE IF NOT EXISTS notification_templates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    event_type      VARCHAR(50) NOT NULL,
    channel_type    VARCHAR(50) NOT NULL,
    subject         VARCHAR(500) DEFAULT '',
    body            TEXT DEFAULT '',
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, event_type, channel_type)
);

CREATE INDEX IF NOT EXISTS idx_notification_templates_org ON notification_templates(organization_id);

CREATE TABLE IF NOT EXISTS notification_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    notification_id UUID REFERENCES notifications(id) ON DELETE CASCADE,
    channel_id      UUID REFERENCES notification_channels(id) ON DELETE SET NULL,
    channel_type    VARCHAR(50) NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    retry_count     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_history_notification ON notification_history(notification_id);
CREATE INDEX IF NOT EXISTS idx_notification_history_status ON notification_history(organization_id, status);

ALTER TABLE notification_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_history ENABLE ROW LEVEL SECURITY;

CREATE POLICY notification_templates_tenant ON notification_templates
    USING (organization_id = current_setting('app.tenant_id', true)::uuid);

CREATE POLICY notification_history_tenant ON notification_history
    USING (organization_id = current_setting('app.tenant_id', true)::uuid);
