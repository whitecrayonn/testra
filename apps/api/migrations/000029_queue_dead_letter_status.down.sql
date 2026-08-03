-- Revert dead_letter rows back to failed and remove the cleanup index.
UPDATE queue_jobs
SET status = 'failed'
WHERE status = 'dead_letter';

DROP INDEX IF EXISTS idx_queue_jobs_status_updated;
