-- Rename historical failed-but-exhausted jobs to dead_letter status and add an
-- index for efficient cleanup of terminal (completed/dead_letter) jobs.
UPDATE queue_jobs
SET status = 'dead_letter'
WHERE status = 'failed' AND attempts >= max_attempts;

CREATE INDEX IF NOT EXISTS idx_queue_jobs_status_updated ON queue_jobs (status, updated_at);
