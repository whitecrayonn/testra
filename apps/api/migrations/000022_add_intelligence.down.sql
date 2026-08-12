DROP POLICY IF EXISTS flaky_predictions_tenant ON flaky_predictions;
DROP POLICY IF EXISTS failure_clusters_tenant ON failure_clusters;
ALTER TABLE flaky_predictions DISABLE ROW LEVEL SECURITY;
ALTER TABLE failure_clusters DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS failure_clusters;
DROP TABLE IF EXISTS flaky_predictions;
