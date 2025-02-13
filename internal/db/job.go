package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Job represents a job in the database
type Job struct {
	ID          string          `json:"id"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Result      json.RawMessage `json:"result,omitempty"`
	Error       string          `json:"error,omitempty"`
	WebhookURL  string          `json:"webhook_url,omitempty"`
	WebhookSent bool            `json:"webhook_sent"`
}

// JobStore handles job-related database operations
type JobStore struct {
	db *DB
}

// NewJobStore creates a new JobStore
func NewJobStore(db *DB) *JobStore {
	return &JobStore{db: db}
}

// CreateJob creates a new job in the database
func (s *JobStore) CreateJob(ctx context.Context, job *Job) error {
	query := `
        INSERT INTO jobs (id, status, created_at, webhook_url)
        VALUES ($1, $2, $3, $4)
    `
	_, err := s.db.ExecContext(ctx, query,
		job.ID,
		job.Status,
		job.CreatedAt,
		job.WebhookURL,
	)
	return err
}

// UpdateJobStatus updates the status of a job
func (s *JobStore) UpdateJobStatus(
	ctx context.Context,
	id string,
	status string,
	result interface{},
	errMsg string,
) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	query := `
        UPDATE jobs 
        SET status = $1, result = $2, error = $3, updated_at = CURRENT_TIMESTAMP
        WHERE id = $4
    `
	_, err = s.db.ExecContext(ctx, query, status, resultJSON, errMsg, id)
	return err
}
