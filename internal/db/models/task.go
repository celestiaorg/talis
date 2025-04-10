package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Field names for task model
const (
	// TaskStatusField is the field name for task status
	TaskStatusField = "status"
	// TaskNameField is the field name for task name
	TaskNameField = "name"
)

// TaskStatus represents the current state of a task
type TaskStatus string

// Task status constants
const (
	// TaskStatusUnknown represents an unknown or invalid task status
	TaskStatusUnknown TaskStatus = "unknown"
	// TaskStatusPending indicates the task is waiting to be processed
	TaskStatusPending TaskStatus = "pending"
	// TaskStatusRunning indicates the task is currently being processed
	TaskStatusRunning TaskStatus = "running"
	// TaskStatusCompleted indicates the task has been successfully completed
	TaskStatusCompleted TaskStatus = "completed"
	// TaskStatusFailed indicates the task has failed
	TaskStatusFailed TaskStatus = "failed"
	// TaskStatusTerminated indicates the task was manually aborted
	TaskStatusTerminated TaskStatus = "terminated"
)

// Task represents an asynchronous operation that can be tracked
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

// MarshalJSON implements the json.Marshaler interface for Task
func (t Task) MarshalJSON() ([]byte, error) {
	type Alias Task // Create an alias to avoid infinite recursion
	return json.Marshal(Alias(t))
}

// String returns the string representation of the task status
func (s TaskStatus) String() string {
	return string(s)
}

// ParseTaskStatus converts a string to a TaskStatus type
func ParseTaskStatus(str string) (TaskStatus, error) {
	switch str {
	case string(TaskStatusUnknown):
		return TaskStatusUnknown, nil
	case string(TaskStatusPending):
		return TaskStatusPending, nil
	case string(TaskStatusRunning):
		return TaskStatusRunning, nil
	case string(TaskStatusCompleted):
		return TaskStatusCompleted, nil
	case string(TaskStatusFailed):
		return TaskStatusFailed, nil
	case string(TaskStatusTerminated):
		return TaskStatusTerminated, nil
	default:
		return TaskStatusUnknown, fmt.Errorf("invalid task status: %s", str)
	}
}

// UnmarshalJSON implements json.Unmarshaler for TaskStatus
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

// MarshalJSON implements json.Marshaler for TaskStatus
func (s *TaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Validate ensures that the task data is valid
func (t *Task) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	return nil
}

// BeforeCreate is a GORM hook that runs before creating a new task
func (t *Task) BeforeCreate(_ *gorm.DB) error {
	if t.Status == "" {
		t.Status = TaskStatusPending
	}
	return t.Validate()
}
