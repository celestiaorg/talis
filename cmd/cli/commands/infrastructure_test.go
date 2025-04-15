package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/test"
)

func setupInfraCommand() *cobra.Command {
	// Create a new root command for testing
	cmd := &cobra.Command{
		Use:   "talis",
		Short: "Talis CLI tool",
	}

	// Add the infra command and its subcommands
	infraCmd := &cobra.Command{
		Use:   "infra",
		Short: "Manage infrastructure",
	}
	cmd.AddCommand(infraCmd)

	// Add create command
	createCmd := createInfraCmd
	createCmd.ResetFlags()
	createCmd.Flags().StringP("file", "f", "", "JSON file containing infrastructure configuration")
	_ = createCmd.MarkFlagRequired("file")
	infraCmd.AddCommand(createCmd)

	// Add delete command
	deleteCmd := deleteInfraCmd
	deleteCmd.ResetFlags()
	deleteCmd.Flags().StringP("file", "f", "", "JSON file containing infrastructure configuration")
	_ = deleteCmd.MarkFlagRequired("file")
	infraCmd.AddCommand(deleteCmd)

	return cmd
}

func setupTestJob(t *testing.T, suite *test.Suite) *models.Job {
	// Create a test user first
	user := &models.User{
		Username: "test-user",
		Email:    "test@example.com",
		Role:     1,
	}
	result := suite.DB.Create(user)
	require.NoError(t, result.Error)

	// Create a test job
	job := &models.Job{
		Name:        "test-job",
		ProjectName: "test-project",
		OwnerID:     user.ID,
	}
	result = suite.DB.Create(job)
	require.NoError(t, result.Error)

	return job
}

func TestCreateInfraCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		inputFile      string
		inputContent   string
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name:      "successful create",
			args:      []string{"infra", "create", "--file", "test.json"},
			inputFile: "test.json",
			inputContent: `{
  "job_name": "test-job",
  "project_name": "test-project",
  "instance_name": "test-instance",
  "instances": [
    {
      "name": "instance-1",
      "number_of_instances": 1,
      "region": "nyc1",
      "provider": "do",
      "size": "s-1vcpu-1gb",
      "image": "ubuntu-20-04-x64",
      "ssh_key_name": "test-key-1",
      "tags": ["test", "dev"],
      "volumes": [
        {
          "name": "test-volume",
          "size_gb": 10,
          "mount_point": "/mnt/data"
        }
      ]
    }
  ]
}`,
			expectedOutput: "Infrastructure creation request submitted successfully\nDelete file generated: ",
		},
		{
			name:          "missing file flag",
			args:          []string{"infra", "create"},
			expectedError: "required flag(s) \"file\" not set",
		},
		{
			name:          "file not found",
			args:          []string{"infra", "create", "--file", "nonexistent.json"},
			expectedError: "error validating file path",
		},
		{
			name:      "invalid JSON",
			args:      []string{"infra", "create", "--file", "invalid.json"},
			inputFile: "invalid.json",
			inputContent: `{
  "invalid": json
}`,
			expectedError: "error parsing JSON file",
		},
		{
			name:      "empty instances array",
			args:      []string{"infra", "create", "--file", "empty.json"},
			inputFile: "empty.json",
			inputContent: `{
  "job_name": "test-job",
  "project_name": "test-project",
  "instance_name": "test-instance",
  "instances": []
}`,
			expectedError: "no instances specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			// Create test job if needed
			if tt.name == "successful create" {
				setupTestJob(t, suite)
			}

			// Create temporary directory for test files
			tmpDir := t.TempDir()

			// Create input file if specified
			var filePath string
			if tt.inputFile != "" {
				filePath = filepath.Join(tmpDir, tt.inputFile)
				err := os.WriteFile(filePath, []byte(tt.inputContent), 0644) //nolint:gosec
				require.NoError(t, err)

				// Update args to use the temporary file path
				for i, arg := range tt.args {
					if arg == tt.inputFile {
						tt.args[i] = filePath
					}
				}
			}

			// Store the original client and restore it after the test
			originalClient := apiClient
			apiClient = suite.APIClient
			defer func() { apiClient = originalClient }()

			// Create a buffer to capture output
			buf := new(bytes.Buffer)
			// Store the original stdout and restore it after the test
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Use WaitGroup to ensure we capture all output
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = io.Copy(buf, r)
			}()

			// Execute command
			cmd := setupInfraCommand()
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Close the write end of the pipe and restore stdout
			_ = w.Close()
			os.Stdout = originalStdout

			// Wait for output to be copied
			wg.Wait()
			_ = r.Close()

			// Check error
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			output := buf.String()
			assert.Contains(t, output, tt.expectedOutput)

			// Check if delete file was created
			if tt.expectedOutput != "" {
				deleteFilePath := filepath.Join(tmpDir, "delete_"+tt.inputFile)
				_, err = os.Stat(deleteFilePath)
				assert.NoError(t, err)

				// Read and verify delete file content
				content, err := os.ReadFile(deleteFilePath) //nolint:gosec
				assert.NoError(t, err)

				var deleteReq types.DeleteInstanceRequest
				err = json.Unmarshal(content, &deleteReq)
				assert.NoError(t, err)
				assert.Equal(t, "test-job", deleteReq.JobName)
				assert.Contains(t, deleteReq.InstanceNames, "instance-1")
			}
		})
	}
}

func TestDeleteInfraCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		inputFile      string
		inputContent   string
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name:      "successful delete",
			args:      []string{"infra", "delete", "--file", "delete.json"},
			inputFile: "delete.json",
			inputContent: `{
  "job_name": "test-job",
  "project_name": "test-project",
  "instance_names": ["instance-1"]
}`,
			expectedOutput: "Infrastructure deletion request submitted successfully\n",
		},
		{
			name:          "missing file flag",
			args:          []string{"infra", "delete"},
			expectedError: "required flag(s) \"file\" not set",
		},
		{
			name:          "file not found",
			args:          []string{"infra", "delete", "--file", "nonexistent.json"},
			expectedError: "error validating file path",
		},
		{
			name:      "invalid JSON",
			args:      []string{"infra", "delete", "--file", "invalid.json"},
			inputFile: "invalid.json",
			inputContent: `{
  "invalid": json
}`,
			expectedError: "error parsing JSON file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			// Create test job if needed
			if tt.name == "successful delete" {
				setupTestJob(t, suite)
			}

			// Create temporary directory for test files
			tmpDir := t.TempDir()

			// Create input file if specified
			var filePath string
			if tt.inputFile != "" {
				filePath = filepath.Join(tmpDir, tt.inputFile)
				err := os.WriteFile(filePath, []byte(tt.inputContent), 0644) //nolint:gosec
				require.NoError(t, err)

				// Update args to use the temporary file path
				for i, arg := range tt.args {
					if arg == tt.inputFile {
						tt.args[i] = filePath
					}
				}
			}

			// Store the original client and restore it after the test
			originalClient := apiClient
			apiClient = suite.APIClient
			defer func() { apiClient = originalClient }()

			// Create a buffer to capture output
			buf := new(bytes.Buffer)
			// Store the original stdout and restore it after the test
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Use WaitGroup to ensure we capture all output
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = io.Copy(buf, r)
			}()

			// Execute command
			cmd := setupInfraCommand()
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Close the write end of the pipe and restore stdout
			_ = w.Close()
			os.Stdout = originalStdout

			// Wait for output to be copied
			wg.Wait()
			_ = r.Close()

			// Check error
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}
