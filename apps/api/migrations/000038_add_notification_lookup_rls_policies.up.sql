-- Enable authenticated-user lookup for notification tables before tenant_id is known.
DROP POLICY IF EXISTS notifications_lookup_user ON notifications;
CREATE POLICY notifications_lookup_user ON notifications
    USING (organization_id IN (
        SELECT organization_id FROM organization_members
        WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
    ));

DROP POLICY IF EXISTS notification_preferences_lookup_user ON notification_preferences;
CREATE POLICY notification_preferences_lookup_user ON notification_preferences
    USING (organization_id IN (
        SELECT organization_id FROM organization_members
        WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
    ));

DROP POLICY IF EXISTS notification_channels_lookup_user ON notification_channels;
CREATE POLICY notification_channels_lookup_user ON notification_channels
    USING (organization_id IN (
        SELECT organization_id FROM organization_members
        WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
    ));

DROP POLICY IF EXISTS notification_templates_lookup_user ON notification_templates;
CREATE POLICY notification_templates_lookup_user ON notification_templates
    USING (organization_id IN (
        SELECT organization_id FROM organization_members
        WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
    ));

DROP POLICY IF EXISTS notification_history_lookup_user ON notification_history;
CREATE POLICY notification_history_lookup_user ON notification_history
    USING (organization_id IN (
        SELECT organization_id FROM organization_members
        WHERE user_id = NULLIF(current_setting('app.lookup_user_id', true), '')::uuid
    ));
