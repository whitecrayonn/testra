DROP POLICY IF EXISTS analytics_dashboards_tenant ON analytics_dashboards;
DROP POLICY IF EXISTS analytics_daily_metrics_tenant ON analytics_daily_metrics;
ALTER TABLE analytics_dashboards DISABLE ROW LEVEL SECURITY;
ALTER TABLE analytics_daily_metrics DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS analytics_daily_metrics;
DROP TABLE IF EXISTS analytics_dashboards;
