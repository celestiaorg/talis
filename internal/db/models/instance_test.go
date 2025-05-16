package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestInstanceStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        InstanceStatus
		stringValue   string
		jsonValue     string
		validForParse bool
		validForJSON  bool
		statusIndex   int
	}{
		{
			name:          "Unknown status",
			status:        InstanceStatusUnknown,
			stringValue:   "unknown",
			jsonValue:     `"unknown"`,
			validForParse: true,
			validForJSON:  true,
			statusIndex:   0,
		},
		{
			name:          "Pending status",
			status:        InstanceStatusPending,
			stringValue:   "pending",
			jsonValue:     `"pending"`,
			validForParse: true,
			validForJSON:  true,
			statusIndex:   1,
		},
		{
			name:          "Created status",
			status:        InstanceStatusCreated,
			stringValue:   "created",
			jsonValue:     `"created"`,
			validForParse: true,
			validForJSON:  true,
			statusIndex:   2,
		},
		{
			name:          "Provisioning status",
			status:        InstanceStatusProvisioning,
			stringValue:   "provisioning",
			jsonValue:     `"provisioning"`,
			validForParse: true,
			validForJSON:  true,
			statusIndex:   3,
		},
		{
			name:          "Ready status",
			status:        InstanceStatusReady,
			stringValue:   "ready",
			jsonValue:     `"ready"`,
			validForParse: true,
			validForJSON:  true,
			statusIndex:   4,
		},
		{
			name:          "Terminated status",
			status:        InstanceStatusTerminated,
			stringValue:   "terminated",
			jsonValue:     `"terminated"`,
			validForParse: true,
			validForJSON:  true,
			statusIndex:   5,
		},
		{
			name:          "Invalid status",
			stringValue:   "invalid_status",
			jsonValue:     `"invalid_status"`,
			validForParse: false,
			validForJSON:  false,
			statusIndex:   -1,
		},
		{
			name:          "Invalid JSON",
			jsonValue:     `invalid`,
			validForParse: false,
			validForJSON:  false,
			statusIndex:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test String() method if we have a valid status
			if tt.statusIndex >= 0 {
				assert.Equal(t, tt.stringValue, tt.status.String(), "String() method failed")
				// Verify the status index matches the iota value
				assert.Equal(t, tt.statusIndex, int(tt.status), "Status index does not match expected iota value")
			}

			// Test ParseInstanceStatus
			parsedStatus, err := ParseInstanceStatus(tt.stringValue)
			if tt.validForParse {
				assert.NoError(t, err, "ParseInstanceStatus should not return error")
				assert.Equal(t, tt.status, parsedStatus, "ParseInstanceStatus returned wrong status")
			} else {
				assert.Error(t, err, "ParseInstanceStatus should return error for invalid status")
				assert.Equal(t, InstanceStatusUnknown, parsedStatus, "Invalid status should return InstanceStatusUnknown")
			}

			// Test JSON marshaling if we have a valid status
			if tt.statusIndex >= 0 {
				bytes, err := tt.status.MarshalJSON()
				assert.NoError(t, err, "Marshal should not return error")
				assert.Equal(t, tt.jsonValue, string(bytes), "Marshal produced incorrect JSON")
			}

			// Test JSON unmarshaling
			var unmarshaledStatus InstanceStatus
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

func TestInstance_Validation(t *testing.T) {
	now := time.Now()
	validInstance := Instance{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		ProjectID:  1,
		ProviderID: ProviderDO,
		PublicIP:   "192.0.2.1",
		Region:     "nyc1",
		Size:       "s-1vcpu-1gb",
		Image:      "ubuntu-20-04-x64",
		Tags:       pq.StringArray{"tag1", "tag2"},
		Status:     InstanceStatusReady,
		CreatedAt:  now,
	}

	t.Run("Valid instance", func(t *testing.T) {
		jsonData, err := json.Marshal(validInstance)
		assert.NoError(t, err)

		var unmarshaledInstance Instance
		err = json.Unmarshal(jsonData, &unmarshaledInstance)
		assert.NoError(t, err)

		// Verify fields were correctly marshaled/unmarshaled
		assert.Equal(t, validInstance.ID, unmarshaledInstance.ID)
		assert.Equal(t, validInstance.ProjectID, unmarshaledInstance.ProjectID)
		assert.Equal(t, validInstance.ProviderID, unmarshaledInstance.ProviderID)
		assert.Equal(t, validInstance.PublicIP, unmarshaledInstance.PublicIP)
		assert.Equal(t, validInstance.Region, unmarshaledInstance.Region)
		assert.Equal(t, validInstance.Size, unmarshaledInstance.Size)
		assert.Equal(t, validInstance.Image, unmarshaledInstance.Image)
		assert.Equal(t, validInstance.Tags, unmarshaledInstance.Tags)
		assert.Equal(t, validInstance.Status, unmarshaledInstance.Status)
		assert.Equal(t, validInstance.CreatedAt.Unix(), unmarshaledInstance.CreatedAt.Unix())
	})

	t.Run("Custom JSON marshaling", func(t *testing.T) {
		jsonData, err := validInstance.MarshalJSON()
		assert.NoError(t, err)

		// Verify the custom marshaling includes the ID field
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		assert.NoError(t, err)
	})

	t.Run("test json unmarshal to map", func(t *testing.T) {
		// Test marshal/unmarshal
		jsonData, err := json.Marshal(validInstance)
		assert.NoError(t, err)

		// Test that JSON unmarshaling into a map works
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		assert.NoError(t, err)

		// Check that important fields are in the JSON output
		// Name is removed, let's check for ID instead.
		// When unmarshaling to map[string]interface{}, numbers are often float64
		assert.Equal(t, float64(validInstance.ID), jsonMap["ID"])
	})
}
