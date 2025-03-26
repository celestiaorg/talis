//go:build !lint
// +build !lint

package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/api/v1/client/mock"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// setupTestCommand sets up a test command with a mock client
func setupTestCommand(t *testing.T) (*cobra.Command, *mock.MockClient, *bytes.Buffer) {
	// Create a mock client
	mockClient := &mock.MockClient{}

	// Save the original client instance and restore it after the test
	originalClientInstance := clientInstance
	t.Cleanup(func() {
		clientInstance = originalClientInstance
	})

	// Set the mock client as the client instance
	clientInstance = mockClient

	// Create a buffer to capture command output
	outputBuf := &bytes.Buffer{}

	// Get the command to test
	cmd := GetInfraCmd()

	// Set the output buffer for the root command and all subcommands
	cmd.SetOut(outputBuf)
	cmd.SetErr(outputBuf)
	for _, subCmd := range cmd.Commands() {
		subCmd.SetOut(outputBuf)
		subCmd.SetErr(outputBuf)
	}

	return cmd, mockClient, outputBuf
}

// createTempJSONFile creates a temporary JSON file for testing
func createTempJSONFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	err := os.WriteFile(filePath, []byte(content), 0600)
	require.NoError(t, err, "Failed to create temp JSON file")

	return filePath
}

func TestCreateInfrastructureCommand(t *testing.T) {
	// Setup test command with mock client
	cmd, mockClient, outputBuf := setupTestCommand(t)

	// Create a sample JSON file
	jsonContent := `{
		"name": "test-infra",
		"project_name": "test-project",
		"instances": [
			{
				"provider": "aws",
				"number_of_instances": 1,
				"region": "us-west-2",
				"size": "t2.micro",
				"image": "ami-12345",
				"tags": ["test"],
				"ssh_key_name": "test-key"
			}
		]
	}`
	jsonFile := createTempJSONFile(t, jsonContent)

	// Configure mock client to return a successful response for create
	mockClient.CreateJobFn = func(ctx context.Context, req infrastructure.InstanceCreateRequest) (*infrastructure.Response, error) {
		// Verify request
		assert.Equal(t, "test-infra", req.InstanceName)
		assert.Equal(t, "test-project", req.ProjectName)
		assert.Len(t, req.Instances, 1)
		assert.Equal(t, "aws", req.Instances[0].Provider)

		// Return mock response
		return &infrastructure.Response{
			ID:     123,
			Status: "created",
		}, nil
	}

	// Execute the create command
	cmd.SetArgs([]string{"create", "-f", jsonFile})
	err := cmd.Execute()
	require.NoError(t, err, "Create command execution failed")

	// Verify that the mock client was called
	require.Len(t, mockClient.CreateJobCalls, 1, "CreateJob should be called once")

	// Verify command output
	output := outputBuf.String()
	assert.Contains(t, output, `"ID": 123`)
	assert.Contains(t, output, `"Status": "created"`)

	// Verify that the delete file was created
	deleteFilePath := filepath.Join(filepath.Dir(jsonFile), "delete_test.json")
	assert.FileExists(t, deleteFilePath)

	// Verify the content of the delete file
	// #nosec G304 -- This is a test file and the path is constructed from a temporary directory
	deleteFileContent, err := os.ReadFile(deleteFilePath)
	require.NoError(t, err)

	var deleteReq infrastructure.DeleteInstanceRequest
	err = json.Unmarshal(deleteFileContent, &deleteReq)
	require.NoError(t, err)

	assert.Equal(t, uint(123), deleteReq.ID)
	assert.Equal(t, "test-infra", deleteReq.InstanceName)
	assert.Equal(t, "test-project", deleteReq.ProjectName)
	assert.Len(t, deleteReq.Instances, 1)

	// Now test the delete command with the generated delete file
	// Reset the output buffer
	outputBuf.Reset()

	// Configure mock client to return a successful response for delete
	mockClient.DeleteJobInstanceFn = func(ctx context.Context, jobID string, req infrastructure.DeleteInstanceRequest) (*infrastructure.Response, error) {
		// Verify request
		assert.Equal(t, "123", jobID)
		assert.Equal(t, uint(123), req.ID)
		assert.Equal(t, "test-infra", req.InstanceName)
		assert.Equal(t, "test-project", req.ProjectName)
		assert.Len(t, req.Instances, 1)

		// Return mock response
		return &infrastructure.Response{
			ID:     123,
			Status: "deleted",
		}, nil
	}

	// Execute the delete command with the generated delete file
	cmd.SetArgs([]string{"delete", "-f", deleteFilePath})
	err = cmd.Execute()
	require.NoError(t, err, "Delete command execution failed")

	// Verify that the mock client was called
	require.Len(t, mockClient.DeleteJobInstanceCalls, 1, "DeleteJobInstance should be called once")

	// Verify command output
	output = outputBuf.String()
	assert.Contains(t, output, `"ID": 123`)
	assert.Contains(t, output, `"Status": "deleted"`)
}

func TestDeleteInfrastructureCommand(t *testing.T) {
	// Setup test command with mock client
	cmd, mockClient, outputBuf := setupTestCommand(t)

	// Create a sample JSON file
	jsonContent := `{
		"id": 123,
		"name": "test-infra",
		"project_name": "test-project",
		"instances": [
			{
				"provider": "aws",
				"number_of_instances": 1,
				"region": "us-west-2",
				"size": "t2.micro",
				"image": "ami-12345",
				"tags": ["test"],
				"ssh_key_name": "test-key"
			}
		]
	}`
	jsonFile := createTempJSONFile(t, jsonContent)

	// Configure mock client to return a successful response
	mockClient.DeleteJobInstanceFn = func(ctx context.Context, jobID string, req infrastructure.DeleteInstanceRequest) (*infrastructure.Response, error) {
		// Verify request
		assert.Equal(t, "123", jobID)
		assert.Equal(t, uint(123), req.ID)
		assert.Equal(t, "test-infra", req.InstanceName)
		assert.Equal(t, "test-project", req.ProjectName)
		assert.Len(t, req.Instances, 1)

		// Return mock response
		return &infrastructure.Response{
			ID:     123,
			Status: "deleted",
		}, nil
	}

	// Execute the command
	cmd.SetArgs([]string{"delete", "-f", jsonFile})
	err := cmd.Execute()
	require.NoError(t, err, "Command execution failed")

	// Verify that the mock client was called
	require.Len(t, mockClient.DeleteJobInstanceCalls, 1, "DeleteJobInstance should be called once")

	// Verify command output
	output := outputBuf.String()
	assert.Contains(t, output, `"ID": 123`)
	assert.Contains(t, output, `"Status": "deleted"`)
}
