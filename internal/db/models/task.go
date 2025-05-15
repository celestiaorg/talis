package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// Field names for task model
const (
	// TaskStatusField is the field name for task status
	TaskStatusField = "status"
	// TaskIDField is the field name for task ID
	TaskIDField = "id"
	// TaskCreatedAtField is the field name for task created at
	TaskCreatedAtField = "created_at"
	// TaskLockedAtField is the field name for task locked at
	TaskLockedAtField = "locked_at"
	// TaskLockExpiryField is the field name for task lock expiry
	TaskLockExpiryField = "lock_expiry"
	// TaskAttemptsField is the field name for task attempts
	TaskAttemptsField = "attempts"
	// TaskLogsField is the field name for task logs
	TaskLogsField = "logs"
	// TaskErrorField is the field name for task error
	TaskErrorField = "error"
	// TaskPriorityField is the field name for task priority
	TaskPriorityField = "priority"

	// WebhookTimeoutSeconds is the timeout for webhook requests in seconds
	WebhookTimeout = 10 * time.Second
	// WebhookContentType is the content type for webhook requests
	WebhookContentType = "application/json"

	// TaskLockTimeout is the duration after which a task lock is considered expired
	TaskLockTimeout = 5 * time.Minute
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

// TaskAction represents the possible actions a task can perform.
type TaskAction string

const (
	// TaskActionCreateInstances represents the action to create instances.
	TaskActionCreateInstances TaskAction = "create_instances"
	// TaskActionTerminateInstances represents the action to terminate instances.
	TaskActionTerminateInstances TaskAction = "terminate_instances"
)

// TaskPriority represents the priority level of a task
type TaskPriority int

const (
	// TaskPriorityHigh represents high priority tasks (like create operations)
	TaskPriorityHigh TaskPriority = 1
	// TaskPriorityLow represents low priority tasks (like terminate operations)
	TaskPriorityLow TaskPriority = 2
)

// String returns the string representation of the task priority
func (p TaskPriority) String() string {
	switch p {
	case TaskPriorityHigh:
		return "high"
	case TaskPriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

// Task represents an asynchronous operation that can be tracked
type Task struct {
	gorm.Model
	ProjectID   uint            `json:"project_id" gorm:"not null; index"`
	OwnerID     uint            `json:"-" gorm:"not null; index"`
	InstanceID  uint            `json:"instance_id,omitempty" gorm:"index"` // Link to the specific instance, if applicable
	Action      TaskAction      `json:"action" gorm:"type:varchar(32)"`     // make sure this is long enough to handle all actions
	Status      TaskStatus      `json:"status" gorm:"not null; index"`
	Payload     json.RawMessage `json:"payload,omitempty" gorm:"type:jsonb"` // Data that is required for the task to be executed
	Result      json.RawMessage `json:"result,omitempty" gorm:"type:jsonb"`  // Result of the task
	Attempts    uint            `json:"attempts" gorm:"not null; default:0"`
	Logs        string          `json:"logs,omitempty" gorm:"type:text"`
	Error       string          `json:"error,omitempty" gorm:"type:text"`
	WebhookURL  string          `json:"webhook_url,omitempty" gorm:"type:text"`
	WebhookSent bool            `json:"webhook_sent" gorm:"not null;default:false;index"`
	CreatedAt   time.Time       `json:"created_at" gorm:"index"`
	LockedAt    *time.Time      `json:"locked_at,omitempty" gorm:"index"`   // When the task was locked for processing
	LockExpiry  *time.Time      `json:"lock_expiry,omitempty" gorm:""`      // When the lock expires
	Priority    TaskPriority    `json:"priority" gorm:"not null;default:1"` // Task priority (higher number = lower priority)
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
	// Validate Action field
	switch t.Action {
	case TaskActionCreateInstances, TaskActionTerminateInstances:
		// Valid actions
	default:
		return fmt.Errorf("invalid task action: %s", t.Action)
	}

	return nil
}

// BeforeCreate is a GORM hook that runs before creating a new task
func (t *Task) BeforeCreate(_ *gorm.DB) error {
	if t.Status == "" {
		t.Status = TaskStatusPending
	}

	// Set default priority based on action
	if t.Priority == 0 {
		switch t.Action {
		case TaskActionCreateInstances:
			t.Priority = TaskPriorityHigh
		case TaskActionTerminateInstances:
			t.Priority = TaskPriorityLow
		default:
			t.Priority = TaskPriorityHigh // Default to high priority
		}
	}

	return t.Validate()
}

// IsLocked checks if the task is currently locked for processing
func (t *Task) IsLocked() bool {
	return t.LockedAt != nil && t.LockExpiry != nil && time.Now().Before(*t.LockExpiry)
}

// Lock locks the task for processing
func (t *Task) Lock(duration time.Duration) {
	now := time.Now()
	expiry := now.Add(duration)
	t.LockedAt = &now
	t.LockExpiry = &expiry
}

// Unlock unlocks the task
func (t *Task) Unlock() {
	t.LockedAt = nil
	t.LockExpiry = nil
}

// SendWebhook sends a notification to the webhook URL if configured
func (t *Task) SendWebhook() error {
	if t.WebhookURL == "" {
		return nil // No webhook configured
	}

	// Create payload
	payload := map[string]interface{}{
		"task_id": t.ID,
		"status":  t.Status,
		"action":  t.Action,
	}

	if t.Error != "" {
		payload["error"] = t.Error
	}

	if t.Result != nil {
		payload["result"] = t.Result
	}

	// Convert payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Send HTTP request
	client := &http.Client{
		Timeout: WebhookTimeout,
	}

	resp, err := client.Post(t.WebhookURL, WebhookContentType, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	// Check the error returned by Close, as the linter suggests.
	closeErr := resp.Body.Close()
	if closeErr != nil {
		// Log the error or handle it as appropriate. Here we return it wrapped.
		return fmt.Errorf("failed to close response body: %w", closeErr)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status code: %d", resp.StatusCode)
	}

	return nil
}
