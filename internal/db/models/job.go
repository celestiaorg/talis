package models

import (
	"encoding/json"
	"time"
)

const (
	JobCreatedAtField = "created_at"
	JobUpdatedAtField = "updated_at"
)

// JobStatus represents the possible states of a job
type JobStatus string

const (
	JobStatusPending      JobStatus = "pending"
	JobStatusInitializing JobStatus = "initializing"
	JobStatusProvisioning JobStatus = "provisioning"
	JobStatusConfiguring  JobStatus = "configuring"
	JobStatusCompleted    JobStatus = "completed"
	JobStatusFailed       JobStatus = "failed"
	JobStatusDeleted      JobStatus = "deleted"
)

type Job struct {
	ID          string          `json:"id" gorm:"primaryKey;varchar(50)"`
	OwnerID     string          `json:"owner_id" gorm:"not null;varchar(50);index"` // TODO: For future use
	Status      JobStatus       `json:"status" gorm:"not null;varchar(20);index"`
	Result      json.RawMessage `json:"result,omitempty" gorm:"type:jsonb"`
	Error       string          `json:"error,omitempty" gorm:"type:text"`
	WebhookURL  string          `json:"webhook_url,omitempty" gorm:"type:text"`
	WebhookSent bool            `json:"webhook_sent" gorm:"not null;default:false;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ListOptions struct {
	Limit  int
	Offset int
	Status JobStatus
}
