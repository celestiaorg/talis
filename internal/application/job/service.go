package job

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/celestiaorg/talis/internal/domain/job"
)

// Constants for job status
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusDeleted   = "deleted"
)

// service implements the job.Service interface
type service struct {
	repo job.Repository
}

// NewService creates a new job service instance
func NewService(repo job.Repository) job.Service {
	return &service{repo: repo}
}

// scanJob scans a database row into a Job
func scanJob(rows *sql.Rows) (*job.Job, error) {
	var j job.Job
	var result sql.NullString  // Handle NULL for result
	var updatedAt sql.NullTime // Handle NULL for updated_at
	var error sql.NullString   // Handle NULL for error

	err := rows.Scan(
		&j.ID,
		&j.Status,
		&j.CreatedAt,
		&updatedAt,
		&result,
		&error, // Use 'error' column instead of 'error_message'
		&j.WebhookURL,
		&j.WebhookSent,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan job: %v", err)
	}

	// Convert sql.NullString to json.RawMessage if not NULL
	if result.Valid {
		j.Result = json.RawMessage(result.String)
	}

	// Convert sql.NullTime to time.Time if not NULL
	if updatedAt.Valid {
		j.UpdatedAt = updatedAt.Time
	}

	// Assign error_message from error column
	if error.Valid {
		j.ErrorMessage = error.String
	}

	return &j, nil
}

// ListJobs returns a list of jobs with optional filtering
func (s *service) ListJobs(ctx context.Context, opts *job.ListOptions) ([]*job.Job, error) {
	query := `SELECT id, status, created_at, updated_at, result, error, webhook_url, webhook_sent 
             FROM jobs`

	// Add filters if provided
	var conditions []string
	var args []interface{}

	if opts != nil {
		if opts.Status != "" {
			conditions = append(conditions, "status = ?")
			args = append(args, opts.Status)
		}
		// Add more filters as needed
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if opts != nil && opts.Limit != "" {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := s.repo.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %v", err)
	}
	defer rows.Close()

	var jobs []*job.Job
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// CreateJob creates a new job with the given webhook URL
func (s *service) CreateJob(ctx context.Context, webhookURL string) (*job.Job, error) {
	j := &job.Job{
		ID:         fmt.Sprintf("job-%s", time.Now().Format("20060102-150405")),
		Status:     job.StatusPending,
		CreatedAt:  time.Now(),
		WebhookURL: webhookURL,
	}

	if err := s.repo.Create(ctx, j); err != nil {
		return nil, fmt.Errorf("failed to create job: %v", err)
	}

	return j, nil
}

func (s *service) GetJobStatus(ctx context.Context, id string) (*job.Job, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) UpdateJobStatus(ctx context.Context, id string, status string, result interface{}, errMsg string) error {
	// Validate status
	switch status {
	case job.StatusPending,
		job.StatusInitializing,
		job.StatusProvisioning,
		job.StatusConfiguring,
		job.StatusCompleted,
		job.StatusFailed,
		job.StatusDeleted:
		// Valid status
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	// If we're deleting and get a "not found" error, consider it a success
	if status == job.StatusFailed && result != nil {
		if resultMap, ok := result.(map[string]string); ok {
			if resultMap["note"] == "some resources were already deleted" {
				status = job.StatusDeleted
				errMsg = "" // Clear error message as this is actually a success case
			}
		}
	}

	return s.repo.UpdateStatus(ctx, id, status, result, errMsg)
}
