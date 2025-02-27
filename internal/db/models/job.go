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

// JobStatus represents the current state of a job in the system
type JobStatus int

// Job status constants
const (
	// JobStatusUnknown represents an unknown or invalid job status
	JobStatusUnknown JobStatus = iota
	// JobStatusPending indicates the job is waiting to be processed
	JobStatusPending
	// JobStatusInitializing indicates the job is currently being processed
	JobStatusInitializing
	// JobStatusProvisioning indicates the job is currently being processed
	JobStatusProvisioning
	// JobStatusConfiguring indicates the job is currently being processed
	JobStatusConfiguring
	// JobStatusCompleted indicates the job has finished successfully
	JobStatusCompleted
	// JobStatusFailed indicates the job has failed to complete
	JobStatusFailed
	// JobStatusTerminated indicates the job has been terminated
	JobStatusTerminated
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

// ParseJobStatus converts a string representation of a job status to JobStatus type
func ParseJobStatus(str string) (JobStatus, error) {
	for i, status := range []string{
		"unknown",
		"pending",
		"initializing",
		"provisioning",
		"configuring",
		"completed",
		"failed",
		"terminated",
	} {
		if status == str {
			return JobStatus(i), nil
		}
	}

	return JobStatus(0), fmt.Errorf("invalid job status: %s", str)
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

func (s JobStatus) String() string {
	return []string{
		"unknown",
		"pending",
		"initializing",
		"provisioning",
		"configuring",
		"completed",
		"failed",
		"terminated",
	}[s]
}
