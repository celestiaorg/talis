// Package job provides database operations for job management.
package job

// This file implements the database storage layer for jobs.
// It provides a Store type that handles all job-related database operations
// following the Repository pattern.
//
// Key components:
//   - Store: Implements database operations for jobs
//   - Create: Creates new jobs in the database
//   - UpdateStatus: Updates job status and results
//
// The Store type implements the job.Repository interface from domain/job/repository.go,
// providing concrete PostgreSQL implementations for:
//   - Creating jobs
//   - Updating job status
//   - Retrieving job information
//   - Listing jobs with filters
//
// Usage example:
//   db := db.NewDB(connStr)
//   store := job.NewStore(db)
//   err := store.Create(ctx, &job.Job{...})
//   err = store.UpdateStatus(ctx, "job-id", job.StatusCompleted, result, "")

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/celestiaorg/talis/internal/db"
)

// Store handles job-related database operations
type Store struct {
	db *db.DB
}

// NewStore creates a new job store
func NewStore(db *db.DB) *Store {
	return &Store{db: db}
}

// Create creates a new job in the database
func (s *Store) Create(ctx context.Context, job *Job) error {
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

// UpdateStatus updates the status of a job
func (s *Store) UpdateStatus(ctx context.Context, id string, status Status, result interface{}, errMsg string) error {
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

// GetByID retrieves a job by its ID
func (s *Store) GetByID(ctx context.Context, id string) (*Job, error) {
	query := `
		SELECT id, status, created_at, updated_at, result, error, webhook_url, webhook_sent
		FROM jobs
		WHERE id = $1
	`
	var job Job
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.Status,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.Result,
		&job.Error,
		&job.WebhookURL,
		&job.WebhookSent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	return &job, nil
}

// List returns a list of jobs with optional filters
func (s *Store) List(ctx context.Context, opts *ListOptions) ([]*Job, error) {
	query := `
		SELECT id, status, created_at, updated_at, result, error, webhook_url, webhook_sent
		FROM jobs
		WHERE ($1::text IS NULL OR status = $1)
		ORDER BY created_at DESC
		LIMIT CASE WHEN $2::integer > 0 THEN $2 ELSE NULL END
	`
	rows, err := s.db.QueryContext(ctx, query, opts.Status, opts.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %v", err)
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		var job Job
		err := rows.Scan(
			&job.ID,
			&job.Status,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.Result,
			&job.Error,
			&job.WebhookURL,
			&job.WebhookSent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %v", err)
		}
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

// Query executes a custom query and returns the rows
func (s *Store) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}
