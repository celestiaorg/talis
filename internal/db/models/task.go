package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TaskStatus represents the current state of a task
type TaskStatus string

// Task status constants
const (
	// TaskStatusUnknown represents an unknown task status
	TaskStatusUnknown TaskStatus = "unknown"
	// TaskStatusPending represents a task that is waiting to be executed
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusRunning represents a task that is currently executing
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusComplete represents a task that has finished successfully
	TaskStatusComplete TaskStatus = "complete"
	// TaskStatusFailed represents a task that has failed to execute
	TaskStatusFailed TaskStatus = "failed"
)

// Task represents a unit of work within a project
type Task struct {
	gorm.Model
	ProjectID   uint            `json:"-" gorm:"not null; index"`
	OwnerID     uint            `json:"-" gorm:"not null; index"`
	Name        string          `json:"name" gorm:"not null; index"`
	Status      TaskStatus      `json:"status" gorm:"not null; index"`
	Result      json.RawMessage `json:"result,omitempty" gorm:"type:jsonb"`
	Error       string          `json:"error,omitempty" gorm:"type:text"`
	WebhookURL  string          `json:"webhook_url,omitempty" gorm:"type:text"`
	WebhookSent bool            `json:"webhook_sent" gorm:"not null;default:false;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"index"`
}

func (s TaskStatus) String() string {
	return string(s)
}


// ParseTaskStatus converts a string to a TaskStatus
func ParseTaskStatus(str string) (TaskStatus, error) {
	switch str {
	case string(TaskStatusUnknown):
		return TaskStatusUnknown, nil
	case string(TaskStatusPending):
		return TaskStatusPending, nil
	case string(TaskStatusRunning):
		return TaskStatusRunning, nil
	case string(TaskStatusComplete):
		return TaskStatusComplete, nil
	case string(TaskStatusFailed):
		return TaskStatusFailed, nil
	default:
		return TaskStatusUnknown, fmt.Errorf("invalid task status: %s", str)
	}
}

// UnmarshalJSON implements the json.Unmarshaler interface for TaskStatus
func (s *TaskStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status, err := ParseTaskStatus(str)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

// MarshalJSON implements the json.Marshaler interface for TaskStatus
func (s *TaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
