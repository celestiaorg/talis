package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/stretchr/testify/require"
)

// Helper function to create a temporary file or directory and make it absolute
func createTempPath(t *testing.T, name string, isDir bool, size int) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, name)

	if isDir {
		err := os.Mkdir(path, 0750)
		require.NoError(t, err)
	} else {
		content := make([]byte, size)
		err := os.WriteFile(path, content, 0600)
		require.NoError(t, err)
	}

	absPath, err := filepath.Abs(path)
	require.NoError(t, err)
	return absPath
}

// Helper function to create a temporary file with specific size and make it absolute
func createTempFile(t *testing.T, name string, size int) string {
	return createTempPath(t, name, false, size)
}

// Helper function to create a temporary directory and make it absolute
func createTempDir(t *testing.T, name string) string {
	return createTempPath(t, name, true, 0)
}

// Helper function to create a base valid InstanceRequest for incremental testing
func baseValidRequest(t *testing.T, payloadPath string) InstanceRequest {
	t.Helper()

	// Set environment variable for SSH key name
	err := os.Setenv(constants.EnvTalisSSHKeyName, "test-key")
	require.NoError(t, err)

	return InstanceRequest{
		OwnerID:           1,
		Provider:          models.ProviderID("mock"),
		Region:            "nyc1",
		Size:              "s-1vcpu-1gb",
		Image:             "ubuntu-20-04-x64",
		Tags:              []string{"test", "dev"},
		ProjectName:       "test-project",
		NumberOfInstances: 1,
		Provision:         true, // Assume provision is true for payload/volume tests initially
		PayloadPath:       payloadPath,
		ExecutePayload:    false,
		Volumes: []VolumeConfig{
			{
				Name:       "test-volume",
				SizeGB:     10,
				MountPoint: "/mnt/data",
			},
		},
		Action: "create", // Add a default valid action
	}
}

func TestInstanceRequest_Validate(t *testing.T) {
	// Create test artifacts once for all test cases
	tempDir := createTempDir(t, "payload-dir")
	validPayloadFile := createTempFile(t, "valid-payload.sh", maxPayloadSize)
	oversizePayloadFile := createTempFile(t, "oversize-payload.sh", maxPayloadSize+1)
	nonexistentPayloadFile := filepath.Join(t.TempDir(), "nonexistent.sh") // Does not exist

	// Set up environment variable for SSH key
	err := os.Setenv(constants.EnvTalisSSHKeyName, "test-key")
	require.NoError(t, err)

	// Cleanup after test
	defer func() {
		err := os.Unsetenv(constants.EnvTalisSSHKeyName)
		if err != nil {
			t.Logf("Failed to unset %s: %v", constants.EnvTalisSSHKeyName, err)
		}
	}()

	// Base valid request to modify for failure cases
	baseReq := baseValidRequest(t, validPayloadFile)

	tests := []struct {
		name    string
		request InstanceRequest
		wantErr bool
		errMsg  string // Substring to check for in the error message
	}{
		// --- Metadata Validations ---
		{
			name:    "Error: missing project_name",
			request: func() InstanceRequest { r := baseReq; r.ProjectName = ""; return r }(),
			wantErr: true,
			errMsg:  "project_name is required",
		},
		{
			name:    "Error: missing owner_id",
			request: func() InstanceRequest { r := baseReq; r.OwnerID = 0; return r }(),
			wantErr: true,
			errMsg:  "owner_id is required",
		},

		// --- User Defined Configs Validations ---
		{
			name:    "Error: missing provider",
			request: func() InstanceRequest { r := baseReq; r.Provider = ""; return r }(),
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name:    "Error: invalid provider",
			request: func() InstanceRequest { r := baseReq; r.Provider = "invalid"; return r }(),
			wantErr: true,
			errMsg:  "unsupported provider",
		},
		{
			name:    "Error: missing region",
			request: func() InstanceRequest { r := baseReq; r.Region = ""; return r }(),
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name:    "Error: missing size, memory, and cpu",
			request: func() InstanceRequest { r := baseReq; r.Size = ""; return r }(),
			wantErr: true,
			errMsg:  "either size, or both memory and cpu, must be provided",
		},
		{
			name: "Error: only memory provided (without cpu)",
			request: func() InstanceRequest {
				r := baseReq
				r.Size = ""
				r.Memory = 2048
				r.CPU = 0
				return r
			}(),
			wantErr: true,
			errMsg:  "either size, or both memory and cpu, must be provided",
		},
		{
			name: "Error: only cpu provided (without memory)",
			request: func() InstanceRequest {
				r := baseReq
				r.Size = ""
				r.Memory = 0
				r.CPU = 2
				return r
			}(),
			wantErr: true,
			errMsg:  "either size, or both memory and cpu, must be provided",
		},
		{
			name:    "Error: missing image",
			request: func() InstanceRequest { r := baseReq; r.Image = ""; return r }(),
			wantErr: true,
			errMsg:  "image is required",
		},
		{
			name:    "Error: number_of_instances less than 1",
			request: func() InstanceRequest { r := baseReq; r.NumberOfInstances = 0; return r }(),
			wantErr: true,
			errMsg:  "number_of_instances must be greater than 0",
		},

		// --- Volume Validations ---
		{
			name:    "Error: missing volumes",
			request: func() InstanceRequest { r := baseReq; r.Volumes = nil; return r }(),
			wantErr: true,
			errMsg:  "at least one volume configuration is required",
		},
		{
			name: "Error: invalid volume configuration",
			request: func() InstanceRequest {
				r := baseReq
				r.Volumes = []VolumeConfig{{Region: "nyc2"}} // Invalid volume
				return r
			}(),
			wantErr: true,
			errMsg:  "invalid volume configuration",
		},

		// --- Payload Validations ---
		{
			name: "Error: Payload path is relative",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = "not/absolute.sh"
				r.Provision = true // Ensure provision is true for payload tests
				r.ExecutePayload = false
				return r
			}(),
			wantErr: true,
			errMsg:  "payload_path must be an absolute path",
		},
		{
			name: "Error: Payload path does not exist",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = nonexistentPayloadFile // Absolute, non-existent
				r.Provision = true
				r.ExecutePayload = false
				return r
			}(),
			wantErr: true,
			errMsg:  "payload_path file does not exist:",
		},
		{
			name: "Error: Payload path is a directory",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = tempDir
				r.Provision = true
				r.ExecutePayload = false
				return r
			}(),
			wantErr: true,
			errMsg:  "payload_path cannot be a directory",
		},
		{
			name: "Error: Payload size exceeds limit",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = oversizePayloadFile
				r.Provision = true
				r.ExecutePayload = false
				return r
			}(),
			wantErr: true,
			errMsg:  "payload file size exceeds the limit of 2MB",
		},
		{
			name: "Error: ExecutePayload=true, PayloadPath empty",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = ""
				r.Provision = true
				r.ExecutePayload = true // Requires PayloadPath
				return r
			}(),
			wantErr: true,
			errMsg:  "payload_path is required when execute_payload is true",
		},
		{
			name: "Error: PayloadPath present, Provision=false",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = validPayloadFile
				r.Provision = false // Invalid with payload
				r.ExecutePayload = false
				return r
			}(),
			wantErr: true,
			errMsg:  "provision must be true when payload_path is provided",
		},

		// --- Action Validation ---
		{
			name:    "Error: missing action",
			request: func() InstanceRequest { r := baseReq; r.Action = ""; return r }(),
			wantErr: true,
			errMsg:  "action is required",
		},

		// --- Valid Cases --- //
		{
			name: "Valid: Request with memory and cpu (no size)",
			request: func() InstanceRequest {
				r := baseReq
				r.Size = ""
				r.Memory = 2048
				r.CPU = 2
				return r
			}(),
			wantErr: false,
		},
		{
			name: "Valid: Request with size only (explicit test)",
			request: func() InstanceRequest {
				r := baseReq
				r.Memory = 0
				r.CPU = 0
				return r
			}(),
			wantErr: false,
		},
		{
			name: "Valid: Minimal request without payload (provision=false)",
			request: func() InstanceRequest {
				r := baseReq
				r.Provision = false
				r.PayloadPath = "" // No payload
				r.ExecutePayload = false
				return r
			}(),
			wantErr: false,
		},
		{
			name: "Valid: Request with payload copy only (provision=true)",
			request: func() InstanceRequest {
				r := baseReq
				r.Provision = true
				r.PayloadPath = validPayloadFile
				r.ExecutePayload = false
				return r
			}(),
			wantErr: false,
		},
		{
			name: "Valid: Request with payload copy and execute (provision=true)",
			request: func() InstanceRequest {
				r := baseReq
				r.Provision = true
				r.PayloadPath = validPayloadFile
				r.ExecutePayload = true
				return r
			}(),
			wantErr: false,
		},
		{
			name:    "Valid: Full base request (already tested implicitly)",
			request: baseReq,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clone the request to avoid modifying the original test case struct during the run
			reqCopy := tt.request // Copy the struct for this specific test run

			err := reqCopy.Validate()

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none")
				if tt.errMsg != "" {
					// Use Contains because some error messages might include dynamic paths or wrap underlying errors
					require.True(t, strings.Contains(err.Error(), tt.errMsg),
						fmt.Sprintf("Expected error message to contain '%s', but got: %s", tt.errMsg, err.Error()))
				}
			} else {
				require.NoError(t, err, fmt.Sprintf("Expected no error, but got: %v", err))
			}
		})
	}
}
