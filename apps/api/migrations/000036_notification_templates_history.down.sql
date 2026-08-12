-- Revert notification templates and history.

DROP POLICY IF EXISTS notification_history_tenant ON notification_history;
DROP POLICY IF EXISTS notification_templates_tenant ON notification_templates;

ALTER TABLE notification_history DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates DISABLE ROW LEVEL SECURITY;

DROP INDEX IF EXISTS idx_notification_history_status;
DROP INDEX IF EXISTS idx_notification_history_notification;

DROP TABLE IF EXISTS notification_history;
DROP TABLE IF EXISTS notification_templates;
