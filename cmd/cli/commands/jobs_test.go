package commands

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/test"
)

func setupJobsCommand() *cobra.Command {
	// Create a new root command for testing
	cmd := &cobra.Command{
		Use:   "talis",
		Short: "Talis CLI tool",
	}

	// Add the jobs command and its subcommands
	jobsCmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage jobs",
	}
	cmd.AddCommand(jobsCmd)

	// Add list command
	listCmd := listJobsCmd
	listCmd.ResetFlags()
	listCmd.Flags().StringP("limit", "l", "", "Limit the number of jobs returned")
	listCmd.Flags().StringP("status", "s", "", "Filter jobs by status")
	jobsCmd.AddCommand(listCmd)

	// Add get command
	getCmd := getJobCmd
	getCmd.ResetFlags()
	getCmd.Flags().StringP("id", "i", "", "Job ID to fetch")
	_ = getCmd.MarkFlagRequired("id")
	jobsCmd.AddCommand(getCmd)

	return cmd
}

func TestListJobsCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponse   types.ListJobsResponse
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful list with no filters",
			args: []string{"jobs", "list"},
			mockResponse: types.ListJobsResponse{
				Jobs: []models.Job{
					{Name: "job1", Status: models.JobStatusCompleted},
					{Name: "job2", Status: models.JobStatusPending},
				},
			},
			expectedOutput: `{
  "jobs": [
    {
      "name": "job2",
      "status": "pending"
    },
    {
      "name": "job1",
      "status": "completed"
    }
  ]
}`,
		},
		{
			name: "successful list with limit",
			args: []string{"jobs", "list", "--limit", "1"},
			mockResponse: types.ListJobsResponse{
				Jobs: []models.Job{
					{Name: "job1", Status: models.JobStatusCompleted},
				},
			},
			expectedOutput: `{
  "jobs": [
    {
      "name": "job1",
      "status": "completed"
    }
  ]
}`,
		},
		{
			name:          "invalid limit value",
			args:          []string{"jobs", "list", "--limit", "invalid"},
			expectedError: "invalid limit value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			// Set up mock response in the database
			if tt.mockResponse.Jobs != nil {
				for _, job := range tt.mockResponse.Jobs {
					result := suite.DB.Create(&job)
					require.NoError(t, result.Error)
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
			cmd := setupJobsCommand()
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
			assert.Equal(t, tt.expectedOutput+"\n", buf.String())
		})
	}
}

func TestGetJobCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockResponse   models.Job
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful get",
			args: []string{"jobs", "get", "--id", "1"},
			mockResponse: models.Job{
				Name:         "job1",
				Status:       models.JobStatusCompleted,
				OwnerID:      1,
				InstanceName: "test-instance",
				ProjectName:  "test-project",
			},
			expectedOutput: `{
  "name": "job1",
  "status": "completed"
}`,
		},
		{
			name:          "missing job ID",
			args:          []string{"jobs", "get"},
			expectedError: "required flag(s) \"id\" not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			// Set up mock response in the database
			if tt.mockResponse.Name != "" {
				result := suite.DB.Create(&tt.mockResponse)
				require.NoError(t, result.Error)
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
			cmd := setupJobsCommand()
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
			assert.Equal(t, tt.expectedOutput+"\n", buf.String())
		})
	}
}
