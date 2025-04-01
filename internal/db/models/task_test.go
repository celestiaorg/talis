package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTaskStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        TaskStatus
		stringValue   string
		jsonValue     string
		validForParse bool
		validForJson  bool
	}{
		{
			name:          "Unknown status",
			status:        TaskStatusUnknown,
			stringValue:   "unknown",
			jsonValue:     `"unknown"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Pending status",
			status:        TaskStatusPending,
			stringValue:   "pending",
			jsonValue:     `"pending"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Running status",
			status:        TaskStatusRunning,
			stringValue:   "running",
			jsonValue:     `"running"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Completed status",
			status:        TaskStatusCompleted,
			stringValue:   "completed",
			jsonValue:     `"completed"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Failed status",
			status:        TaskStatusFailed,
			stringValue:   "failed",
			jsonValue:     `"failed"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Invalid status",
			stringValue:   "invalid_status",
			jsonValue:     `"invalid_status"`,
			validForParse: false,
			validForJson:  false,
		},
		{
			name:          "Invalid JSON",
			jsonValue:     `invalid`,
			validForParse: false,
			validForJson:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test String() method if we have a valid status
			if tt.status != "" {
				assert.Equal(t, tt.stringValue, tt.status.String(), "String() method failed")
			}

			// Test ParseTaskStatus
			parsedStatus, err := ParseTaskStatus(tt.stringValue)
			if tt.validForParse {
				assert.NoError(t, err, "ParseTaskStatus should not return error")
				assert.Equal(t, tt.status, parsedStatus, "ParseTaskStatus returned wrong status")
			} else {
				assert.Error(t, err, "ParseTaskStatus should return error for invalid status")
				assert.Equal(t, TaskStatusUnknown, parsedStatus, "Invalid status should return TaskStatusUnknown")
			}

			// Test JSON marshaling if we have a valid status
			if tt.status != "" {
				bytes, err := tt.status.MarshalJSON()
				assert.NoError(t, err, "Marshal should not return error")
				assert.Equal(t, tt.jsonValue, string(bytes), "Marshal produced incorrect JSON")
			}

			// Test JSON unmarshaling
			var unmarshaledStatus TaskStatus
			err = unmarshaledStatus.UnmarshalJSON([]byte(tt.jsonValue))
			if tt.validForJson {
				assert.NoError(t, err, "Unmarshal should not return error")
				assert.Equal(t, tt.status, unmarshaledStatus, "Unmarshal produced incorrect status")
			} else {
				assert.Error(t, err, "Unmarshal should return error for invalid JSON")
			}
		})
	}
}

func TestTask_Validation(t *testing.T) {
	now := time.Now()
	result := json.RawMessage(`{"key": "value"}`)

	validTask := Task{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		ProjectID:   1,
		Name:        "test-task",
		Status:      TaskStatusPending,
		Result:      result,
		Error:       "",
		WebhookURL:  "https://example.com/webhook",
		WebhookSent: false,
		CreatedAt:   now,
	}

	t.Run("Valid task", func(t *testing.T) {
		jsonData, err := json.Marshal(validTask)
		assert.NoError(t, err)

		var unmarshaledTask Task
		err = json.Unmarshal(jsonData, &unmarshaledTask)
		assert.NoError(t, err)

		// Verify fields were correctly marshaled/unmarshaled
		// Don't compare ID or ProjectID as they're not included in marshaled output
		// ProjectID has json:"-" tag
		assert.Equal(t, validTask.Name, unmarshaledTask.Name)
		assert.Equal(t, validTask.Status, unmarshaledTask.Status)

		// Compare Result by unmarshaling to map to ignore formatting differences
		var expectedResult, actualResult map[string]interface{}
		assert.NoError(t, json.Unmarshal(validTask.Result, &expectedResult))
		assert.NoError(t, json.Unmarshal(unmarshaledTask.Result, &actualResult))
		assert.Equal(t, expectedResult, actualResult)

		assert.Equal(t, validTask.Error, unmarshaledTask.Error)
		assert.Equal(t, validTask.WebhookURL, unmarshaledTask.WebhookURL)
		assert.Equal(t, validTask.WebhookSent, unmarshaledTask.WebhookSent)
		assert.Equal(t, validTask.CreatedAt.Unix(), unmarshaledTask.CreatedAt.Unix())
	})

	t.Run("Task with nil Result", func(t *testing.T) {
		task := validTask
		task.Result = nil

		jsonData, err := json.Marshal(task)
		assert.NoError(t, err)

		var unmarshaledTask Task
		err = json.Unmarshal(jsonData, &unmarshaledTask)
		assert.NoError(t, err)
		assert.Nil(t, unmarshaledTask.Result)
	})

	t.Run("Task with empty Result", func(t *testing.T) {
		task := validTask
		task.Result = json.RawMessage(`{}`)

		jsonData, err := json.Marshal(task)
		assert.NoError(t, err)

		var unmarshaledTask Task
		err = json.Unmarshal(jsonData, &unmarshaledTask)
		assert.NoError(t, err)
		assert.Equal(t, string(json.RawMessage(`{}`)), string(unmarshaledTask.Result))
	})

	t.Run("Task with different statuses", func(t *testing.T) {
		statuses := []TaskStatus{
			TaskStatusPending,
			TaskStatusRunning,
			TaskStatusCompleted,
			TaskStatusFailed,
		}

		for _, status := range statuses {
			task := validTask
			task.Status = status

			jsonData, err := json.Marshal(task)
			assert.NoError(t, err)

			var unmarshaledTask Task
			err = json.Unmarshal(jsonData, &unmarshaledTask)
			assert.NoError(t, err)
			assert.Equal(t, status, unmarshaledTask.Status)
		}
	})
}
