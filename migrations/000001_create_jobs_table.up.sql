CREATE TABLE jobs (
    id VARCHAR(50) PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    result JSONB,
    error TEXT,
    webhook_url TEXT,
    webhook_sent BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at); 