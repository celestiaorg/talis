package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestJobStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        JobStatus
		stringValue   string
		jsonValue     string
		validForParse bool
		validForJson  bool
	}{
		{
			name:          "Unknown status",
			status:        JobStatusUnknown,
			stringValue:   "unknown",
			jsonValue:     `"unknown"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Pending status",
			status:        JobStatusPending,
			stringValue:   "pending",
			jsonValue:     `"pending"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Initializing status",
			status:        JobStatusInitializing,
			stringValue:   "initializing",
			jsonValue:     `"initializing"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Provisioning status",
			status:        JobStatusProvisioning,
			stringValue:   "provisioning",
			jsonValue:     `"provisioning"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Configuring status",
			status:        JobStatusConfiguring,
			stringValue:   "configuring",
			jsonValue:     `"configuring"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Deleting status",
			status:        JobStatusDeleting,
			stringValue:   "deleting",
			jsonValue:     `"deleting"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Completed status",
			status:        JobStatusCompleted,
			stringValue:   "completed",
			jsonValue:     `"completed"`,
			validForParse: true,
			validForJson:  true,
		},
		{
			name:          "Failed status",
			status:        JobStatusFailed,
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

			// Test ParseJobStatus
			parsedStatus, err := ParseJobStatus(tt.stringValue)
			if tt.validForParse {
				assert.NoError(t, err, "ParseJobStatus should not return error")
				assert.Equal(t, tt.status, parsedStatus, "ParseJobStatus returned wrong status")
			} else {
				assert.Error(t, err, "ParseJobStatus should return error for invalid status")
				assert.Equal(t, JobStatusUnknown, parsedStatus, "Invalid status should return JobStatusUnknown")
			}

			// Test JSON marshaling if we have a valid status
			if tt.status != "" {
				bytes, err := tt.status.MarshalJSON()
				assert.NoError(t, err, "Marshal should not return error")
				assert.Equal(t, tt.jsonValue, string(bytes), "Marshal produced incorrect JSON")
			}

			// Test JSON unmarshaling
			var unmarshaledStatus JobStatus
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

func TestJob_Validation(t *testing.T) {
	now := time.Now()
	validJob := Job{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Name:         "test-job",
		InstanceName: "test-instance",
		ProjectName:  "test-project",
		OwnerID:      1,
		Status:       JobStatusPending,
		SSHKeys:      []string{"key1", "key2"},
		WebhookURL:   "https://example.com/webhook",
		WebhookSent:  false,
	}

	t.Run("Valid job", func(t *testing.T) {
		jsonData, err := json.Marshal(validJob)
		assert.NoError(t, err)

		var unmarshaledJob Job
		err = json.Unmarshal(jsonData, &unmarshaledJob)
		assert.NoError(t, err)

		// Verify fields were correctly marshaled/unmarshaled
		assert.Equal(t, validJob.Name, unmarshaledJob.Name)
		assert.Equal(t, validJob.InstanceName, unmarshaledJob.InstanceName)
		assert.Equal(t, validJob.ProjectName, unmarshaledJob.ProjectName)
		assert.Equal(t, validJob.OwnerID, unmarshaledJob.OwnerID)
		assert.Equal(t, validJob.Status, unmarshaledJob.Status)
		assert.Equal(t, validJob.SSHKeys, unmarshaledJob.SSHKeys)
		assert.Equal(t, validJob.WebhookURL, unmarshaledJob.WebhookURL)
		assert.Equal(t, validJob.WebhookSent, unmarshaledJob.WebhookSent)
	})
}
