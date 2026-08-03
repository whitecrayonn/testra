DROP POLICY IF EXISTS integrations_tenant ON integrations;
DROP POLICY IF EXISTS integration_events_tenant ON integration_events;
ALTER TABLE integrations DISABLE ROW LEVEL SECURITY;
ALTER TABLE integration_events DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS integration_events;
DROP TABLE IF EXISTS integrations;
