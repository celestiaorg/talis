package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSSHKeys(t *testing.T) {
	tests := []struct {
		name         string
		keys         SSHKeys
		jsonValue    string
		validForJSON bool
		expectedKeys SSHKeys
		dbValue      interface{}
		validForScan bool
	}{
		{
			name:         "Valid SSH keys",
			keys:         SSHKeys{"key1", "key2"},
			jsonValue:    `["key1","key2"]`,
			validForJSON: true,
			expectedKeys: SSHKeys{"key1", "key2"},
			dbValue:      []byte(`["key1","key2"]`),
			validForScan: true,
		},
		{
			name:         "Empty SSH keys",
			keys:         SSHKeys{},
			jsonValue:    `[]`,
			validForJSON: true,
			expectedKeys: SSHKeys{},
			dbValue:      []byte(`[]`),
			validForScan: true,
		},
		{
			name:         "Nil SSH keys",
			keys:         nil,
			jsonValue:    `null`,
			validForJSON: true,
			expectedKeys: nil,
			dbValue:      nil,
			validForScan: true,
		},
		{
			name:         "Invalid JSON",
			jsonValue:    `invalid`,
			validForJSON: false,
			dbValue:      []byte(`invalid`),
			validForScan: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			if tt.keys != nil {
				bytes, err := tt.keys.MarshalJSON()
				assert.NoError(t, err, "Marshal should not return error")
				assert.Equal(t, tt.jsonValue, string(bytes), "Marshal produced incorrect JSON")
			}

			// Test JSON unmarshaling
			var unmarshaledKeys SSHKeys
			err := unmarshaledKeys.UnmarshalJSON([]byte(tt.jsonValue))
			if tt.validForJSON {
				assert.NoError(t, err, "Unmarshal should not return error")
				assert.Equal(t, tt.expectedKeys, unmarshaledKeys, "Unmarshal produced incorrect keys")
			} else {
				assert.Error(t, err, "Unmarshal should return error for invalid JSON")
			}

			// Test database Value method
			if tt.keys != nil {
				value, err := tt.keys.Value()
				assert.NoError(t, err, "Value should not return error")
				assert.JSONEq(t, tt.jsonValue, string(value.([]byte)), "Value produced incorrect database value")
			}

			// Test database Scan method
			var scannedKeys SSHKeys
			err = scannedKeys.Scan(tt.dbValue)
			if tt.validForScan {
				assert.NoError(t, err, "Scan should not return error")
				assert.Equal(t, tt.expectedKeys, scannedKeys, "Scan produced incorrect keys")
			} else {
				assert.Error(t, err, "Scan should return error for invalid value")
			}
		})
	}
}

func TestJobStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        JobStatus
		stringValue   string
		jsonValue     string
		validForParse bool
		validForJSON  bool
	}{
		{
			name:          "Unknown status",
			status:        JobStatusUnknown,
			stringValue:   "unknown",
			jsonValue:     `"unknown"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Pending status",
			status:        JobStatusPending,
			stringValue:   "pending",
			jsonValue:     `"pending"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Initializing status",
			status:        JobStatusInitializing,
			stringValue:   "initializing",
			jsonValue:     `"initializing"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Provisioning status",
			status:        JobStatusProvisioning,
			stringValue:   "provisioning",
			jsonValue:     `"provisioning"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Configuring status",
			status:        JobStatusConfiguring,
			stringValue:   "configuring",
			jsonValue:     `"configuring"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Deleting status",
			status:        JobStatusDeleting,
			stringValue:   "deleting",
			jsonValue:     `"deleting"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Completed status",
			status:        JobStatusCompleted,
			stringValue:   "completed",
			jsonValue:     `"completed"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Failed status",
			status:        JobStatusFailed,
			stringValue:   "failed",
			jsonValue:     `"failed"`,
			validForParse: true,
			validForJSON:  true,
		},
		{
			name:          "Invalid status",
			stringValue:   "invalid_status",
			jsonValue:     `"invalid_status"`,
			validForParse: false,
			validForJSON:  false,
		},
		{
			name:          "Invalid JSON",
			jsonValue:     `invalid`,
			validForParse: false,
			validForJSON:  false,
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
			if tt.validForJSON {
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
		SSHKeys:      SSHKeys{"key1", "key2"},
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

	t.Run("Job with nil SSHKeys", func(t *testing.T) {
		job := validJob
		job.SSHKeys = nil

		jsonData, err := json.Marshal(job)
		assert.NoError(t, err)

		var unmarshaledJob Job
		err = json.Unmarshal(jsonData, &unmarshaledJob)
		assert.NoError(t, err)
		assert.Nil(t, unmarshaledJob.SSHKeys)
	})

	t.Run("Job with empty SSHKeys", func(t *testing.T) {
		job := validJob
		job.SSHKeys = SSHKeys{}

		jsonData, err := json.Marshal(job)
		assert.NoError(t, err)

		var unmarshaledJob Job
		err = json.Unmarshal(jsonData, &unmarshaledJob)
		assert.NoError(t, err)
		assert.Empty(t, unmarshaledJob.SSHKeys)
	})
}
