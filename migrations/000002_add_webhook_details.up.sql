ALTER TABLE jobs
ADD COLUMN webhook_headers JSONB,
ADD COLUMN webhook_attempts INT DEFAULT 0,
ADD COLUMN last_webhook_attempt TIMESTAMP;

CREATE INDEX idx_jobs_webhook_sent ON jobs(webhook_sent) WHERE NOT webhook_sent; 