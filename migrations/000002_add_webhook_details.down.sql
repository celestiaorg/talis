DROP INDEX IF EXISTS idx_jobs_webhook_sent;

ALTER TABLE jobs
DROP COLUMN IF EXISTS webhook_headers,
DROP COLUMN IF EXISTS webhook_attempts,
DROP COLUMN IF EXISTS last_webhook_attempt; 