package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusUnknown   TaskStatus = "unknown"
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	gorm.Model
	ID          uint            `json:"-" gorm:"primaryKey"`
	ProjectID   uint            `json:"-" gorm:"not null"`
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
	default:
		return TaskStatusUnknown, fmt.Errorf("invalid task status: %s", str)
	}
}

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

func (s *TaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
