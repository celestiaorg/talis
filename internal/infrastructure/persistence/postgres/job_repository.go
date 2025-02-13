package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/celestiaorg/talis/internal/domain/job"
)

// JobRepository implements job.Repository
type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Query executes a raw SQL query
func (r *JobRepository) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return r.db.QueryContext(ctx, query, args...)
}

// Create creates a new job
func (r *JobRepository) Create(ctx context.Context, job *job.Job) error {
	query := `
		INSERT INTO jobs (id, status, created_at, webhook_url)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		job.ID,
		job.Status,
		job.CreatedAt,
		job.WebhookURL,
	)
	return err
}

// GetByID returns a job by its ID
func (r *JobRepository) GetByID(ctx context.Context, id string) (*job.Job, error) {
	query := `
		SELECT id, status, created_at, updated_at, result, error, webhook_url, webhook_sent
		FROM jobs WHERE id = $1
	`
	j := &job.Job{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&j.ID,
		&j.Status,
		&j.CreatedAt,
		&j.UpdatedAt,
		&j.Result,
		&j.Error,
		&j.WebhookURL,
		&j.WebhookSent,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	}
	return j, err
}

// UpdateStatus updates the status of a job
func (r *JobRepository) UpdateStatus(ctx context.Context, id string, status string, result interface{}, errMsg string) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	query := `
		UPDATE jobs 
		SET status = $1, result = $2, error = $3, updated_at = CURRENT_TIMESTAMP, webhook_sent = false
		WHERE id = $4
	`
	_, err = r.db.ExecContext(ctx, query, status, resultJSON, errMsg, id)
	return err
}

// List returns a list of jobs with optional filtering
func (r *JobRepository) List(ctx context.Context, opts *job.ListOptions) ([]*job.Job, error) {
	query := `
		SELECT id, status, created_at, updated_at, result, error, webhook_url, webhook_sent
		FROM jobs
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC
	`
	if opts.Limit != "" {
		query += fmt.Sprintf(" LIMIT %s", opts.Limit)
	}

	rows, err := r.db.QueryContext(ctx, query, opts.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %v", err)
	}
	defer rows.Close()

	var jobs []*job.Job
	for rows.Next() {
		j := &job.Job{}
		err := rows.Scan(
			&j.ID,
			&j.Status,
			&j.CreatedAt,
			&j.UpdatedAt,
			&j.Result,
			&j.Error,
			&j.WebhookURL,
			&j.WebhookSent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %v", err)
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}
