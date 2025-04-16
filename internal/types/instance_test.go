package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/stretchr/testify/require"
)

// TestInstancesRequest_Validate tests the Validate method for InstancesRequest. It does not test the Validate method for InstanceRequest.
func TestInstancesRequest_Validate(t *testing.T) {
	// create a default valid InstanceRequest
	defaultInstanceRequest := InstanceRequest{
		Provider:          "do",
		NumberOfInstances: 1,
		Region:            "nyc1",
		Size:              "s-1vcpu-1gb",
		Image:             "ubuntu-20-04-x64",
		SSHKeyName:        "test-key",
		Volumes: []VolumeConfig{
			{
				Name:       "test-volume",
				SizeGB:     10,
				MountPoint: "/mnt/data",
			},
		},
	}
	tests := []struct {
		name    string
		request *InstancesRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with instance name",
			request: &InstancesRequest{
				JobName:      "test-job",
				ProjectName:  "test-project",
				InstanceName: "valid-instance",
				Instances:    []InstanceRequest{defaultInstanceRequest},
			},
			wantErr: false,
		},
		{
			name: "valid request with multiple instances",
			request: &InstancesRequest{
				JobName:      "test-job",
				ProjectName:  "test-project",
				InstanceName: "valid-instance",
				Instances:    []InstanceRequest{defaultInstanceRequest, defaultInstanceRequest},
			},
			wantErr: false,
		},
		{
			name:    "missing job name",
			request: &InstancesRequest{},
			wantErr: true,
			errMsg:  "job_name is required",
		},
		{
			name: "missing project name",
			request: &InstancesRequest{
				JobName: "test-job",
			},
			wantErr: true,
			errMsg:  "project_name is required",
		},
		{
			name: "missing instances",
			request: &InstancesRequest{
				JobName:     "test-job",
				ProjectName: "test-project",
			},
			wantErr: true,
			errMsg:  "at least one instance configuration is required",
		},
		{
			name: "empty instances",
			request: &InstancesRequest{
				JobName:     "test-job",
				ProjectName: "test-project",
				Instances:   []InstanceRequest{},
			},
			wantErr: true,
			errMsg:  "at least one instance configuration is required",
		},
		{
			name: "invalid hostname",
			request: &InstancesRequest{
				JobName:      "test-job",
				ProjectName:  "test-project",
				InstanceName: "invalid_hostname$123",
				Instances:    []InstanceRequest{defaultInstanceRequest},
			},
			wantErr: true,
			errMsg:  "invalid hostname format",
		},
		{
			name: "missing instance name and instance name in request",
			request: &InstancesRequest{
				JobName:     "test-job",
				ProjectName: "test-project",
				Instances:   []InstanceRequest{defaultInstanceRequest},
			},
			wantErr: true,
			errMsg:  "instance_name or instance.name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("InstancesRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("InstancesRequest.Validate() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}

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

func TestInstanceRequest_Validate(t *testing.T) {
	// Create test artifacts once for all test cases
	tempDir := createTempDir(t, "payload-dir")
	validPayloadFile := createTempFile(t, "valid-payload.sh", maxPayloadSize)
	oversizePayloadFile := createTempFile(t, "oversize-payload.sh", maxPayloadSize+1)

	defaultVolumeConfig := VolumeConfig{
		Name:       "test-volume",
		SizeGB:     10,
		MountPoint: "/mnt/data",
	}
	tests := []struct {
		name    string
		request InstanceRequest
		wantErr bool
		errMsg  string // Substring to check for in the error message
	}{
		{
			name:    "Error: missing provider",
			request: InstanceRequest{}, // Start with empty
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "Error: invalid number of instances",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"), // Pass previous check
				NumberOfInstances: 0,                         // Invalid
			},
			wantErr: true,
			errMsg:  "number_of_instances must be greater than 0",
		},
		{
			name: "Error: missing region",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1, // Pass previous check
				// Region missing
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "Error: missing size",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1", // Pass previous check
				// Size missing
			},
			wantErr: true,
			errMsg:  "size is required",
		},
		{
			name: "Error: missing image",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb", // Pass previous check
				// Image missing
			},
			wantErr: true,
			errMsg:  "image is required",
		},
		{
			name: "Error: missing ssh key name",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu", // Pass previous check
				// SSHKeyName missing
			},
			wantErr: true,
			errMsg:  "ssh_key_name is required",
		},
		// --- Payload Validations --- //
		{
			name: "Error: Payload path is relative",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key", // Pass previous check
				Provision:         true,       // Required for payload tests
				PayloadPath:       "not/absolute.sh",
			},
			wantErr: true,
			errMsg:  "payload_path must be an absolute path",
		},
		{
			name: "Error: Payload path does not exist",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key",
				Provision:         true,
				PayloadPath:       filepath.Join(t.TempDir(), "nonexistent.sh"), // Absolute, non-existent
			},
			wantErr: true,
			errMsg:  "payload_path file does not exist:",
		},
		{
			name: "Error: Payload path is a directory",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key",
				Provision:         true,
				PayloadPath:       tempDir,
			},
			wantErr: true,
			errMsg:  "payload_path cannot be a directory",
		},
		{
			name: "Error: Payload size exceeds limit",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key",
				Provision:         true,
				PayloadPath:       oversizePayloadFile,
			},
			wantErr: true,
			errMsg:  "payload file size exceeds the limit of 2MB",
		},
		{
			name: "Error: ExecutePayload=true, PayloadPath empty",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key",
				Provision:         true,
				ExecutePayload:    true, // Requires PayloadPath
				PayloadPath:       "",   // Is empty
			},
			wantErr: true,
			errMsg:  "payload_path is required when execute_payload is true",
		},
		{
			name: "Error: PayloadPath present, Provision=false",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key",
				Provision:         false, // Invalid with payload
				PayloadPath:       validPayloadFile,
			},
			wantErr: true,
			errMsg:  "provision must be true when payload_path is provided",
		},
		// --- Valid Cases --- //
		{
			name: "Valid: Minimal request without payload",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key",
				// Provision defaults to false, which is valid here
			},
			wantErr: false,
		},
		{
			name: "Valid: Request with payload copy only",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key",
				Provision:         true, // Required for payload
				ExecutePayload:    false,
				PayloadPath:       validPayloadFile,
			},
			wantErr: false,
		},
		{
			name: "Valid: Request with payload copy and execute",
			request: InstanceRequest{
				Provider:          models.ProviderID("mock"),
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu",
				SSHKeyName:        "test-key",
				Provision:         true, // Required for payload
				ExecutePayload:    true,
				PayloadPath:       validPayloadFile,
			},
			wantErr: false,
		},
		{
			name: "missing volumes",
			request: InstanceRequest{
				Name:              "valid-instance",
				Provider:          "do",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
			},
			wantErr: true,
			errMsg:  "at least one volume configuration is required",
		},
		{
			name: "invalid instance name",
			request: InstanceRequest{
				Name:              "invalid_hostname$123",
				Provider:          "do",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				SSHKeyName:        "test-key",
				Volumes:           []VolumeConfig{defaultVolumeConfig},
			},
			wantErr: true,
			errMsg:  "invalid instance name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clone the request to avoid modifying the original test case struct
			reqCopy := tt.request

			err := reqCopy.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					// Use Contains because some error messages might include dynamic paths
					require.True(t, strings.Contains(err.Error(), tt.errMsg),
						fmt.Sprintf("Expected error message '%s' to contain '%s'", err.Error(), tt.errMsg))
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request *DeleteRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: &DeleteRequest{
				InstanceName: "test-instance",
				ProjectName:  "test-project",
				Instances: []DeleteInstance{
					{
						Provider:          "do",
						Name:              "instance-1",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing instance name",
			request: &DeleteRequest{
				ProjectName: "test-project",
				Instances: []DeleteInstance{
					{
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
					},
				},
			},
			wantErr: true,
			errMsg:  "instance_name is required",
		},
		{
			name: "missing project name",
			request: &DeleteRequest{
				InstanceName: "test-instance",
				Instances: []DeleteInstance{
					{
						Provider:          "do",
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
					},
				},
			},
			wantErr: true,
			errMsg:  "project_name is required",
		},
		{
			name: "empty instances array",
			request: &DeleteRequest{
				InstanceName: "test-instance",
				ProjectName:  "test-project",
				Instances:    []DeleteInstance{},
			},
			wantErr: true,
			errMsg:  "at least one instance configuration is required",
		},
		{
			name: "invalid instance configuration",
			request: &DeleteRequest{
				InstanceName: "test-instance",
				ProjectName:  "test-project",
				Instances: []DeleteInstance{
					{
						NumberOfInstances: 1,
						Region:            "nyc1",
						Size:              "s-1vcpu-1gb",
					},
				},
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("DeleteRequest.Validate() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}

func TestDeleteInstance_Validate(t *testing.T) {
	tests := []struct {
		name     string
		instance *DeleteInstance
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid instance",
			instance: &DeleteInstance{
				Provider:          "do",
				Name:              "instance-1",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			instance: &DeleteInstance{
				Name:              "instance-1",
				NumberOfInstances: 1,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "invalid number of instances",
			instance: &DeleteInstance{
				Provider:          "do",
				Name:              "instance-1",
				NumberOfInstances: 0,
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
			},
			wantErr: true,
			errMsg:  "number_of_instances must be greater than 0",
		},
		{
			name: "missing region",
			instance: &DeleteInstance{
				Provider:          "do",
				Name:              "instance-1",
				NumberOfInstances: 1,
				Size:              "s-1vcpu-1gb",
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "missing size",
			instance: &DeleteInstance{
				Provider:          "do",
				Name:              "instance-1",
				NumberOfInstances: 1,
				Region:            "nyc1",
			},
			wantErr: true,
			errMsg:  "size is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.instance.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteInstance.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("DeleteInstance.Validate() error message = %v, want to contain %v", err, tt.errMsg)
			}
		})
	}
}
