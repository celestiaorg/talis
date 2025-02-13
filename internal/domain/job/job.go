package job

import (
	"context"
	"encoding/json"
	"time"
)

// Job status constants
const (
	StatusPending      = "pending"
	StatusInitializing = "initializing" // Setting up Pulumi
	StatusProvisioning = "provisioning" // Creating infrastructure
	StatusConfiguring  = "configuring"  // Setting up Nix
	StatusCompleted    = "completed"
	StatusFailed       = "failed"
	StatusDeleted      = "deleted"
)

// Job represents an infrastructure task
type Job struct {
	ID           string          `json:"id"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Result       json.RawMessage `json:"result,omitempty"`
	Error        string          `json:"error,omitempty"`
	WebhookURL   string          `json:"webhook_url,omitempty"`
	WebhookSent  bool            `json:"webhook_sent"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

// ListOptions defines options for listing jobs
type ListOptions struct {
	Limit  string
	Status string
}

// Service defines business operations for jobs
type Service interface {
	CreateJob(ctx context.Context, webhookURL string) (*Job, error)
	UpdateJobStatus(ctx context.Context, id string, status string, result interface{}, errMsg string) error
	GetJobStatus(ctx context.Context, id string) (*Job, error)
	ListJobs(ctx context.Context, opts *ListOptions) ([]*Job, error)
}
