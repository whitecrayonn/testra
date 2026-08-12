DROP POLICY IF EXISTS tenant_isolation_queue_jobs ON queue_jobs;

CREATE POLICY tenant_isolation_queue_jobs ON queue_jobs
    USING (tenant_id = app.current_tenant() OR app.current_tenant() IS NULL);
