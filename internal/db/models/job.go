package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	JobCreatedAtField = "created_at"
	JobUpdatedAtField = "updated_at"
)

type JobStatus int

const (
	// we need unknown to be the first status to avoid conflicts with the default value
	// Also allow us to search for all jobs no matter their status
	JobStatusUnknown JobStatus = iota
	JobStatusPending
	JobStatusInitializing
	JobStatusProvisioning
	JobStatusConfiguring
	JobStatusCompleted
	JobStatusFailed
	JobStatusTerminated
)

type Job struct {
	gorm.Model
	Name        string          `json:"name" gorm:"not null; index"`
	ProjectName string          `json:"project_name" gorm:"not null; index"`
	OwnerID     uint            `json:"owner_id" gorm:"not null;index"` // ID from the users table
	Status      JobStatus       `json:"status" gorm:"index"`
	Result      json.RawMessage `json:"result,omitempty" gorm:"type:jsonb"`
	Error       string          `json:"error,omitempty" gorm:"type:text"`
	WebhookURL  string          `json:"webhook_url,omitempty" gorm:"type:text"`
	WebhookSent bool            `json:"webhook_sent" gorm:"not null;default:false;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

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

func (s JobStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
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
