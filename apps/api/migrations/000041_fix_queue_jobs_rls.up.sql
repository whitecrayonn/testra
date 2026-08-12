-- Fix queue_jobs RLS so it fails closed when the worker session variable is not set.
-- The previous policy (tenant_id = app.current_tenant() OR app.current_tenant() IS NULL)
-- allowed any unauthenticated/unscoped connection to read all rows because the
-- right-hand side of the OR evaluated to TRUE whenever the session variable was unset.
--
-- This migration introduces an explicit worker-mode flag. Only the background worker
-- sets app.queue_worker = 'true'. Unset or false leaves the policy fail-closed,
-- matching every other tenant-scoped table.
DROP POLICY IF EXISTS tenant_isolation_queue_jobs ON queue_jobs;

CREATE POLICY tenant_isolation_queue_jobs ON queue_jobs
    USING (
        (current_setting('app.queue_worker', true) = 'true')
        OR
        (tenant_id = app.current_tenant())
    );
