// Package job provides database operations for job management.
package job

import (
	"encoding/json"
	"time"
)

// Status represents the possible states of a job
type Status string

const (
	StatusPending      Status = "pending"
	StatusInitializing Status = "initializing"
	StatusProvisioning Status = "provisioning"
	StatusConfiguring  Status = "configuring"
	StatusCompleted    Status = "completed"
	StatusFailed       Status = "failed"
	StatusDeleted      Status = "deleted"
)

// Job represents a job in the database
type Job struct {
	ID          string          `json:"id"`
	Status      Status          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Result      json.RawMessage `json:"result,omitempty"`
	Error       string          `json:"error,omitempty"`
	WebhookURL  string          `json:"webhook_url,omitempty"`
	WebhookSent bool            `json:"webhook_sent"`
}

// ListOptions defines filters for listing jobs
type ListOptions struct {
	Limit  string // Maximum number of jobs to return
	Status string // Filter by status
}
