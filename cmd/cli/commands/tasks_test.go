package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/test"
)

func setupTasksCommand() *cobra.Command {
	// Create a new root command for testing
	cmd := &cobra.Command{
		Use:   "talis",
		Short: "Talis CLI tool",
	}

	// Add the owner-id flag
	cmd.PersistentFlags().StringP(flagOwnerID, "o", "", "Owner ID for resources")

	// Add the tasks command and its subcommands
	tasksCmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks",
	}
	cmd.AddCommand(tasksCmd)

	// Add get command
	getCmd := getTaskCmd
	getCmd.ResetFlags()
	getCmd.Flags().StringP("name", "n", "", "Task name")
	_ = getCmd.MarkFlagRequired("name")
	tasksCmd.AddCommand(getCmd)

	// Add list command
	listCmd := listTasksCmd
	listCmd.ResetFlags()
	listCmd.Flags().StringP("project", "p", "", "Project name")
	listCmd.Flags().IntP("page", "g", 1, "Page number for pagination")
	_ = listCmd.MarkFlagRequired("project")
	tasksCmd.AddCommand(listCmd)

	// Add terminate command
	terminateCmd := terminateTaskCmd
	terminateCmd.ResetFlags()
	terminateCmd.Flags().StringP("name", "n", "", "Task name")
	_ = terminateCmd.MarkFlagRequired("name")
	tasksCmd.AddCommand(terminateCmd)

	return cmd
}

func TestTasksCommand(t *testing.T) {
	cmd := tasksCmd

	// Test that the tasks command has the expected subcommands
	subCmds := cmd.Commands()
	assert.Equal(t, 3, len(subCmds), "Expected 3 subcommands")

	// Verify the subcommand names
	var subCmdNames []string
	for _, c := range subCmds {
		subCmdNames = append(subCmdNames, c.Name())
	}

	// Expect list, get, and terminate subcommands
	assert.Contains(t, subCmdNames, "list")
	assert.Contains(t, subCmdNames, "get")
	assert.Contains(t, subCmdNames, "terminate")

	// Verify flags for list command
	listCmd := findCommand(subCmds, "list")
	assert.NotNil(t, listCmd)
	assert.True(t, listCmd.Flags().HasFlags())
	projectFlag, _ := listCmd.Flags().GetString("project")
	assert.Equal(t, "", projectFlag)
	pageFlag, _ := listCmd.Flags().GetInt("page")
	assert.Equal(t, 1, pageFlag)

	// Verify flags for get command
	getCmd := findCommand(subCmds, "get")
	assert.NotNil(t, getCmd)
	assert.True(t, getCmd.Flags().HasFlags())
	nameFlag, _ := getCmd.Flags().GetString("name")
	assert.Equal(t, "", nameFlag)

	// Verify flags for terminate command
	terminateCmd := findCommand(subCmds, "terminate")
	assert.NotNil(t, terminateCmd)
	assert.True(t, terminateCmd.Flags().HasFlags())
	taskNameFlag, _ := terminateCmd.Flags().GetString("name")
	assert.Equal(t, "", taskNameFlag)
}

func TestGetTaskCmd(t *testing.T) {
	createdAt := time.Now()

	tests := []struct {
		name           string
		args           []string
		mockTask       models.Task
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful get",
			args: []string{"tasks", "get", "--name", "test-task", "-o", fmt.Sprintf("%v", models.AdminID)},
			mockTask: models.Task{
				Name:      "test-task",
				Status:    models.TaskStatusCompleted,
				Action:    models.TaskActionCreateInstances,
				Logs:      "Task is running...",
				OwnerID:   models.AdminID,
				CreatedAt: createdAt,
			},
			expectedOutput: `{
  "name": "test-task",
  "status": "completed",
  "action": "create_instances",
  "logs": "Task is running...",
  "created_at": "` + createdAt.Format("2006-01-02 15:04:05") + `"
}`,
		},
		{
			name:          "missing task name",
			args:          []string{"tasks", "get"},
			expectedError: "required flag(s) \"name\" not set",
		},
		{
			name:          "missing owner-id",
			args:          []string{"tasks", "get", "--name", "test-task"},
			expectedError: `required flag(s) "owner-id" not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			// Set up mock response in the database
			if tt.mockTask.Name != "" {
				result := suite.DB.Create(&tt.mockTask)
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
			cmd := setupTasksCommand()
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
			if tt.expectedOutput != "" {
				// Normalize the JSON for comparison
				var expected, actual interface{}
				err = json.Unmarshal([]byte(tt.expectedOutput), &expected)
				require.NoError(t, err)
				err = json.Unmarshal(buf.Bytes(), &actual)
				require.NoError(t, err)

				expectedJSON, err := json.Marshal(expected)
				require.NoError(t, err)
				actualJSON, err := json.Marshal(actual)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func TestListTasksCmd(t *testing.T) {
	createdAt := time.Now()

	tests := []struct {
		name           string
		args           []string
		mockProject    models.Project
		mockTasks      []models.Task
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful list",
			args: []string{"tasks", "list", "--project", "test-project", "-o", fmt.Sprintf("%d", models.AdminID)},
			mockProject: models.Project{
				Name:    "test-project",
				OwnerID: models.AdminID,
			},
			mockTasks: []models.Task{
				{
					Name:      "task1",
					Status:    models.TaskStatusCompleted,
					Action:    models.TaskActionCreateInstances,
					OwnerID:   models.AdminID,
					CreatedAt: createdAt,
				},
				{
					Name:      "task2",
					Status:    models.TaskStatusCompleted,
					Action:    models.TaskActionTerminateInstances,
					OwnerID:   models.AdminID,
					CreatedAt: createdAt,
				},
			},
			expectedOutput: `{
  "tasks": [
    {
      "name": "task1",
      "status": "completed",
      "action": "create_instances",
      "created_at": "` + createdAt.Format("2006-01-02 15:04:05") + `"
    },
    {
      "name": "task2",
      "status": "completed",
      "action": "terminate_instances",
      "created_at": "` + createdAt.Format("2006-01-02 15:04:05") + `"
    }
  ]
}`,
		},
		{
			name:          "missing project name",
			args:          []string{"tasks", "list"},
			expectedError: "required flag(s) \"project\" not set",
		},
		{
			name:          "invalid page value",
			args:          []string{"tasks", "list", "--project", "test-project", "--page", "invalid"},
			expectedError: "invalid argument \"invalid\" for \"-g, --page\" flag: strconv.ParseInt: parsing \"invalid\": invalid syntax",
		},
		{
			name:          "missing owner-id",
			args:          []string{"tasks", "list", "--project", "test-project"},
			expectedError: `required flag(s) "owner-id" not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			// Set up mock response in the database
			if tt.mockProject.Name != "" {
				result := suite.DB.Create(&tt.mockProject)
				require.NoError(t, result.Error)

				if tt.mockTasks != nil {
					for i := range tt.mockTasks {
						tt.mockTasks[i].ProjectID = tt.mockProject.ID
						result := suite.DB.Create(&tt.mockTasks[i])
						require.NoError(t, result.Error)
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
			cmd := setupTasksCommand()
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
			if tt.expectedOutput != "" {
				// Normalize the JSON for comparison
				var expected, actual interface{}
				err = json.Unmarshal([]byte(tt.expectedOutput), &expected)
				require.NoError(t, err)
				err = json.Unmarshal(buf.Bytes(), &actual)
				require.NoError(t, err)

				expectedJSON, err := json.Marshal(expected)
				require.NoError(t, err)
				actualJSON, err := json.Marshal(actual)
				require.NoError(t, err)

				assert.JSONEq(t, string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func TestTerminateTaskCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockTask       models.Task
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful terminate",
			args: []string{"tasks", "terminate", "--name", "test-task", "-o", fmt.Sprintf("%d", models.AdminID)},
			mockTask: models.Task{
				Name:    "test-task",
				Status:  models.TaskStatusCompleted,
				Action:  models.TaskActionCreateInstances,
				OwnerID: models.AdminID,
			},
			expectedOutput: "Task 'test-task' termination request submitted successfully",
		},
		{
			name:          "missing task name",
			args:          []string{"tasks", "terminate"},
			expectedError: "required flag(s) \"name\" not set",
		},
		{
			name:          "missing owner-id",
			args:          []string{"tasks", "terminate", "--name", "test-task"},
			expectedError: `required flag(s) "owner-id" not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			// Set up mock response in the database
			if tt.mockTask.Name != "" {
				result := suite.DB.Create(&tt.mockTask)
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
			cmd := setupTasksCommand()
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
			if tt.expectedOutput != "" {
				assert.Contains(t, buf.String(), tt.expectedOutput)
			}
		})
	}
}
