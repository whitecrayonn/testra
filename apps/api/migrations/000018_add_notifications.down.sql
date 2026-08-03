DROP POLICY IF EXISTS notifications_tenant ON notifications;
DROP POLICY IF EXISTS notification_preferences_tenant ON notification_preferences;
DROP POLICY IF EXISTS notification_channels_tenant ON notification_channels;

ALTER TABLE notifications DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_preferences DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_channels DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS notification_channels;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS notifications;

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE name LIKE 'notifications:%' OR name LIKE 'notification_preferences:%' OR name LIKE 'notification_channels:%'
);

DELETE FROM permissions
WHERE name LIKE 'notifications:%' OR name LIKE 'notification_preferences:%' OR name LIKE 'notification_channels:%';
