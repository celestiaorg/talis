package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// JobCreatedAtField is the database field name for the job creation timestamp
const (
	// JobCreatedAtField is the database field name for the job creation timestamp
	JobCreatedAtField = "created_at"
	// JobUpdatedAtField is the database field name for the job update timestamp
	JobUpdatedAtField = "updated_at"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	// JobStatusUnknown represents an unknown job status
	JobStatusUnknown JobStatus = "unknown"
	// JobStatusPending represents a pending job
	JobStatusPending JobStatus = "pending"
	// JobStatusInitializing represents a job that is being initialized
	JobStatusInitializing JobStatus = "initializing"
	// JobStatusProvisioning represents a job that is provisioning infrastructure
	JobStatusProvisioning JobStatus = "provisioning"
	// JobStatusConfiguring represents a job that is configuring the infrastructure
	JobStatusConfiguring JobStatus = "configuring"
	// JobStatusDeleting represents a job that is deleting infrastructure
	JobStatusDeleting JobStatus = "deleting"
	// JobStatusCompleted represents a completed job
	JobStatusCompleted JobStatus = "completed"
	// JobStatusFailed represents a failed job
	JobStatusFailed JobStatus = "failed"
)

// Job represents a task or operation in the system
type Job struct {
	gorm.Model
	Name        string          `json:"name" gorm:"not null; index"`
	ProjectName string          `json:"project_name" gorm:"not null; index"`
	OwnerID     uint            `json:"owner_id" gorm:"not null;index"` // ID from the users table
	SSHKeys     []string        `json:"ssh_keys" gorm:"type:jsonb"`
	Status      JobStatus       `json:"status" gorm:"index"`
	Result      json.RawMessage `json:"result,omitempty" gorm:"type:jsonb"`
	Error       string          `json:"error,omitempty" gorm:"type:text"`
	WebhookURL  string          `json:"webhook_url,omitempty" gorm:"type:text"`
	WebhookSent bool            `json:"webhook_sent" gorm:"not null;default:false;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"index"`
}

// String returns the string representation of the job status
func (s JobStatus) String() string {
	return string(s)
}

// ParseJobStatus converts a string representation of a job status to JobStatus type
func ParseJobStatus(str string) (JobStatus, error) {
	switch str {
	case string(JobStatusUnknown):
		return JobStatusUnknown, nil
	case string(JobStatusPending):
		return JobStatusPending, nil
	case string(JobStatusInitializing):
		return JobStatusInitializing, nil
	case string(JobStatusProvisioning):
		return JobStatusProvisioning, nil
	case string(JobStatusConfiguring):
		return JobStatusConfiguring, nil
	case string(JobStatusDeleting):
		return JobStatusDeleting, nil
	case string(JobStatusCompleted):
		return JobStatusCompleted, nil
	case string(JobStatusFailed):
		return JobStatusFailed, nil
	default:
		return JobStatusUnknown, fmt.Errorf("invalid job status: %s", str)
	}
}

// MarshalJSON implements the json.Marshaler interface for JobStatus
func (s JobStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for JobStatus
func (s *JobStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status, err := ParseJobStatus(str)
	if err != nil {
		return err
	}

	*s = status
	return nil
}
