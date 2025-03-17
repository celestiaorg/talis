//go:build !lint
// +build !lint

package commands

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/api/v1/client/mock"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// setupJobsTestCommand sets up a test command with a mock client
func setupJobsTestCommand(t *testing.T) (*cobra.Command, *mock.MockClient, *bytes.Buffer) {
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
	cmd := GetJobsCmd()

	// Set the output buffer for the root command and all subcommands
	cmd.SetOut(outputBuf)
	for _, subCmd := range cmd.Commands() {
		subCmd.SetOut(outputBuf)
	}

	return cmd, mockClient, outputBuf
}

func TestListJobsCommand(t *testing.T) {
	// Setup test command with mock client
	cmd, mockClient, outputBuf := setupJobsTestCommand(t)

	// Configure mock client to return a successful response
	mockClient.ListJobsFn = func(ctx context.Context, limit int, status string) ([]infrastructure.JobStatus, error) {
		// Verify parameters
		assert.Equal(t, 5, limit)
		assert.Equal(t, "running", status)

		// Return mock response
		return []infrastructure.JobStatus{
			{
				JobID:     "123",
				Status:    "running",
				CreatedAt: "2023-01-01T00:00:00Z",
			},
			{
				JobID:     "456",
				Status:    "running",
				CreatedAt: "2023-01-02T00:00:00Z",
			},
		}, nil
	}

	// Execute the command
	cmd.SetArgs([]string{"list", "-l", "5", "-s", "running"})
	err := cmd.Execute()
	require.NoError(t, err, "Command execution failed")

	// Verify that the mock client was called
	require.Len(t, mockClient.ListJobsCalls, 1, "ListJobs should be called once")

	// Verify command output
	output := outputBuf.String()
	assert.Contains(t, output, `"JobID": "123"`)
	assert.Contains(t, output, `"Status": "running"`)
	assert.Contains(t, output, `"JobID": "456"`)
}

func TestGetJobCommand(t *testing.T) {
	// Setup test command with mock client
	cmd, mockClient, outputBuf := setupJobsTestCommand(t)

	// Configure mock client to return a successful response
	mockClient.GetJobFn = func(ctx context.Context, id string) (*infrastructure.JobStatus, error) {
		// Verify parameters
		assert.Equal(t, "123", id)

		// Return mock response
		return &infrastructure.JobStatus{
			JobID:     "123",
			Status:    "completed",
			CreatedAt: "2023-01-01T00:00:00Z",
		}, nil
	}

	// Execute the command
	cmd.SetArgs([]string{"get", "-i", "123"})
	err := cmd.Execute()
	require.NoError(t, err, "Command execution failed")

	// Verify that the mock client was called
	require.Len(t, mockClient.GetJobCalls, 1, "GetJob should be called once")

	// Verify command output
	output := outputBuf.String()
	assert.Contains(t, output, `"JobID": "123"`)
	assert.Contains(t, output, `"Status": "completed"`)
}
