package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/celestiaorg/talis/pkg/models"
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
func baseValidRequest(t *testing.T, payloadPath string, tarArchivePath string) InstanceRequest {
	t.Helper()
	return InstanceRequest{
		Name:              "valid-instance-name",
		OwnerID:           1,
		Provider:          models.ProviderID("mock"),
		Region:            "nyc1",
		Size:              "s-1vcpu-1gb",
		Image:             "ubuntu-20-04-x64",
		Tags:              []string{"test", "dev"},
		ProjectName:       "test-project",
		SSHKeyName:        "test-key",
		NumberOfInstances: 1,
		Provision:         true, // Assume provision is true for payload/volume/tar tests initially
		PayloadPath:       payloadPath,
		ExecutePayload:    false,
		TarArchivePath:    tarArchivePath,
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

func TestValidateUpload(t *testing.T) {
	tempDir := createTempDir(t, "upload-dir")
	validFile := createTempFile(t, "valid-upload.dat", maxUploadSize)
	oversizeFile := createTempFile(t, "oversize-upload.dat", maxUploadSize+1)
	nonexistentFile := filepath.Join(t.TempDir(), "nonexistent.dat") // Does not exist

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid file",
			path:    validFile,
			wantErr: false,
		},
		{
			name:    "Error: Relative path",
			path:    "not/absolute.dat",
			wantErr: true,
			errMsg:  "must be an absolute path",
		},
		{
			name:    "Error: Path does not exist",
			path:    nonexistentFile, // Absolute, non-existent
			wantErr: true,
			errMsg:  "file does not exist:",
		},
		{
			name:    "Error: Path is a directory",
			path:    tempDir,
			wantErr: true,
			errMsg:  "cannot be a directory",
		},
		{
			name:    "Error: File size exceeds limit",
			path:    oversizeFile,
			wantErr: true,
			errMsg:  "file size exceeds the limit of 2MB",
		},
		{
			name:    "Error: Empty path",
			path:    "",
			wantErr: true,
			errMsg:  "path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpload(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.True(t, strings.Contains(err.Error(), tt.errMsg),
						fmt.Sprintf("Expected error message to contain '%s', but got: %s", tt.errMsg, err.Error()))
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInstanceRequest_Validate(t *testing.T) {
	// Create test artifacts once for all test cases
	validPayloadFile := createTempFile(t, "valid-payload.sh", 100) // Use different sizes to differentiate
	validTarFile := createTempFile(t, "valid-archive.tar.gz", 100)

	// Base valid request to modify for failure cases - now takes two paths
	baseReq := baseValidRequest(t, validPayloadFile, validTarFile)

	tests := []struct {
		name    string
		request InstanceRequest
		wantErr bool
		errMsg  string // Substring to check for in the error message
	}{
		// --- Metadata Validations ---
		{
			name:    "Error: missing name",
			request: func() InstanceRequest { r := baseReq; r.Name = ""; return r }(),
			wantErr: true,
			errMsg:  "instance name is required",
		},
		{
			name:    "Error: invalid name (hostname)",
			request: func() InstanceRequest { r := baseReq; r.Name = "invalid_name!"; return r }(),
			wantErr: true,
			errMsg:  "invalid instance name",
		},
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
			name:    "Error: missing size",
			request: func() InstanceRequest { r := baseReq; r.Size = ""; return r }(),
			wantErr: true,
			errMsg:  "size is required",
		},
		{
			name:    "Error: missing image",
			request: func() InstanceRequest { r := baseReq; r.Image = ""; return r }(),
			wantErr: true,
			errMsg:  "image is required",
		},
		{
			name:    "Error: missing ssh_key_name",
			request: func() InstanceRequest { r := baseReq; r.SSHKeyName = ""; return r }(),
			wantErr: true,
			errMsg:  "ssh_key_name is required",
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

		// --- Payload/Tar Archive and Provisioning Validations ---
		{
			name: "Error: PayloadPath present, Provision=false",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = validPayloadFile
				r.TarArchivePath = "" // Clear tar path for this test
				r.Provision = false   // Invalid with payload
				r.ExecutePayload = false
				return r
			}(),
			wantErr: true,
			errMsg:  "provision must be true when payload_path is provided",
		},
		{
			name: "Error: ExecutePayload=true, PayloadPath empty",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = ""
				r.TarArchivePath = "" // Clear tar path
				r.Provision = true
				r.ExecutePayload = true // Requires PayloadPath
				return r
			}(),
			wantErr: true,
			errMsg:  "payload_path is required when execute_payload is true",
		},
		{
			name: "Error: TarArchivePath present, Provision=false",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = "" // Clear payload path
				r.TarArchivePath = validTarFile
				r.Provision = false // Invalid with tar archive
				r.ExecutePayload = false
				return r
			}(),
			wantErr: true,
			errMsg:  "provision must be true when tar_archive_path is provided",
		},
		// Specific path validity (existence, size, type) is tested in Test_validateUpload
		// InstanceRequest.Validate calls validateUpload, so we trust that function works based on its tests.
		// We just need to test the *logic* within InstanceRequest.Validate itself.
		{
			name: "Error: Invalid Payload Path (delegated check)",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = "relative/path.sh" // Let validateUpload handle the specific error
				r.TarArchivePath = ""
				r.Provision = true
				return r
			}(),
			wantErr: true,
			errMsg:  "invalid payload_path:", // Check that it reports the error comes from payload_path
		},
		{
			name: "Error: Invalid Tar Archive Path (delegated check)",
			request: func() InstanceRequest {
				r := baseReq
				r.PayloadPath = ""
				r.TarArchivePath = "relative/archive.tar.gz" // Let validateUpload handle the specific error
				r.Provision = true
				return r
			}(),
			wantErr: true,
			errMsg:  "invalid tar_archive_path:", // Check that it reports the error comes from tar_archive_path
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
			name: "Valid: Minimal request without payload (provision=false)",
			request: func() InstanceRequest {
				r := baseReq
				r.Provision = false
				r.PayloadPath = ""    // No payload
				r.TarArchivePath = "" // No tar archive
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
				r.TarArchivePath = "" // No tar archive
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
				r.TarArchivePath = "" // No tar archive
				r.ExecutePayload = true
				return r
			}(),
			wantErr: false,
		},
		{
			name: "Valid: Request with tar archive only (provision=true)",
			request: func() InstanceRequest {
				r := baseReq
				r.Provision = true
				r.PayloadPath = "" // No payload
				r.TarArchivePath = validTarFile
				r.ExecutePayload = false
				return r
			}(),
			wantErr: false,
		},
		{
			name: "Valid: Request with both payload and tar archive (provision=true)",
			request: func() InstanceRequest {
				r := baseReq // Already has both paths
				r.Provision = true
				r.ExecutePayload = false // Can't execute tar
				return r
			}(),
			wantErr: false,
		},
		{
			name: "Valid: Full base request (with payload and tar, execute=false)",
			request: func() InstanceRequest {
				r := baseReq             // baseReq now includes both valid paths
				r.ExecutePayload = false // Explicitly set for clarity
				return r
			}(),
			wantErr: false,
		},
		// Note: The baseReq function initializes with valid payload and tar paths.
		// Individual test cases modify it as needed.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clone the request to avoid modifying the original test case struct during the run
			// Needed because some fields are modified in-place (like SSHKeyName)
			reqCopy := tt.request // Copy the struct for this specific test run

			err := reqCopy.Validate()

			if tt.wantErr {
				require.Error(t, err, "Expected an error but got none")
				if tt.errMsg != "" {
					// Use Contains because some error messages might include dynamic paths or wrap underlying errors
					require.True(t, strings.Contains(err.Error(), tt.errMsg),
						fmt.Sprintf("Test '%s': Expected error message to contain '%s', but got: %s", tt.name, tt.errMsg, err.Error()))
				}
			} else {
				require.NoError(t, err, fmt.Sprintf("Test '%s': Expected no error, but got: %v", tt.name, err))

				// Specific post-validation checks for non-error cases
				if tt.name == "Check: ssh_key_name lowercased" {
					require.Equal(t, "test-key", reqCopy.SSHKeyName, "SSHKeyName should be lowercased")
				}
				// Add more post-validation checks if needed for other valid cases
			}
		})
	}
}
