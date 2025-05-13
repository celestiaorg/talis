package handlers

import (
	"strings"
	"testing"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/stretchr/testify/assert"
)

func TestTaskListByInstanceParams_Validate(t *testing.T) {
	tests := []struct {
		name        string
		params      TaskListByInstanceParams
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_params_no_optional_fields",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
			},
			expectError: false,
		},
		{
			name: "valid_params_with_action",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
				Action:     string(models.TaskActionCreateInstances),
				Limit:      10,
			},
			expectError: false,
		},
		{
			name: "valid_params_with_pagination",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
				Limit:      10,
				Offset:     5,
			},
			expectError: false,
		},
		{
			name: "missing_instance_id",
			params: TaskListByInstanceParams{
				OwnerID: 2,
			},
			expectError: true,
			errorMsg:    "instance_id is required and must be a positive number",
		},
		{
			name: "missing_owner_id",
			params: TaskListByInstanceParams{
				InstanceID: 1,
			},
			expectError: true,
			errorMsg:    strings.ToLower(ErrMsgTaskOwnerIDRequired),
		},
		{
			name: "negative_limit",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
				Limit:      -1,
			},
			expectError: true,
			errorMsg:    "limit must be a non-negative number",
		},
		{
			name: "negative_offset",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
				Offset:     -1,
			},
			expectError: true,
			errorMsg:    "offset must be a non-negative number",
		},
		{
			name: "offset_without_limit",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
				Offset:     5,
				Limit:      0,
			},
			expectError: true,
			errorMsg:    "limit must be set when offset is used",
		},
		{
			name: "invalid_action",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
				Action:     "invalid_action",
			},
			expectError: true,
			errorMsg:    "invalid task action: invalid_action",
		},
		{
			name: "valid_create_instances_action",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
				Action:     string(models.TaskActionCreateInstances),
			},
			expectError: false,
		},
		{
			name: "valid_terminate_instances_action",
			params: TaskListByInstanceParams{
				InstanceID: 1,
				OwnerID:    2,
				Action:     string(models.TaskActionTerminateInstances),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
