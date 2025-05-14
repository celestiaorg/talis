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
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/test"
)

// setupTasksCommandTest reinitializes flags for task commands for each test run.
func setupTasksCommandTest(_ *testing.T) *cobra.Command {
	// Reset flags on the actual command variables from tasks.go
	getTaskCmd.ResetFlags()
	getTaskCmd.Flags().UintP(flagTaskID, "i", 0, "Task ID")
	_ = getTaskCmd.MarkFlagRequired(flagTaskID)

	listTasksCmd.ResetFlags()
	listTasksCmd.Flags().StringP(flagProjectName, "p", "", "Project name")
	listTasksCmd.Flags().IntP(flagTaskPage, "g", 1, "Page number for pagination")
	_ = listTasksCmd.MarkFlagRequired(flagProjectName)

	terminateTaskCmd.ResetFlags()
	terminateTaskCmd.Flags().UintP(flagTaskID, "i", 0, "Task ID")
	_ = terminateTaskCmd.MarkFlagRequired(flagTaskID)

	// Reset flags for the new command
	listInstanceTasksCmd.ResetFlags()
	listInstanceTasksCmd.Flags().UintP(flagInstanceID, "I", 0, "Instance ID to list tasks for")
	listInstanceTasksCmd.Flags().StringP(flagTaskAction, "a", "", "Filter tasks by action (e.g., create_instances, terminate_instances)")
	listInstanceTasksCmd.Flags().Int(flagTaskLimit, 0, "Limit the number of tasks returned")
	listInstanceTasksCmd.Flags().Int(flagTaskOffset, 0, "Offset for paginating tasks")
	_ = listInstanceTasksCmd.MarkFlagRequired(flagInstanceID)

	// Create a new root command for testing & attach the actual tasksCmd
	rootCmd := &cobra.Command{Use: "talis"}
	// Ensure the global ownerID flag is available on the root for subcommands that need it via getOwnerID
	rootCmd.PersistentFlags().StringP(flagOwnerID, "o", "", "Owner ID for resources")
	// tasksCmd should already have listInstanceTasksCmd added via its init()
	rootCmd.AddCommand(tasksCmd)
	return rootCmd
}

func TestTasksCommand(t *testing.T) {
	cmd := setupTasksCommandTest(t)                                 // This sets up the main rootCmd with tasksCmd added
	tasksSubCmds := findCommand(cmd.Commands(), "tasks").Commands() // Get subcommands of the actual tasksCmd

	assert.Equal(t, 4, len(tasksSubCmds), "Expected 4 subcommands for tasks")

	var subCmdNames []string
	for _, c := range tasksSubCmds {
		subCmdNames = append(subCmdNames, c.Name())
	}

	assert.Contains(t, subCmdNames, "list")
	assert.Contains(t, subCmdNames, "get")
	assert.Contains(t, subCmdNames, "terminate")
	assert.Contains(t, subCmdNames, "list-by-instance")

	listCmd := findCommand(tasksSubCmds, "list")
	assert.NotNil(t, listCmd)
	require.NotNil(t, listCmd.Flags(), "list command flags should not be nil")
	projectFlag, _ := listCmd.Flags().GetString(flagProjectName)
	assert.Equal(t, "", projectFlag)
	pageFlag, _ := listCmd.Flags().GetInt(flagTaskPage)
	assert.Equal(t, 1, pageFlag)

	getCmd := findCommand(tasksSubCmds, "get")
	assert.NotNil(t, getCmd)
	require.NotNil(t, getCmd.Flags(), "get command flags should not be nil")
	idFlag, _ := getCmd.Flags().GetUint(flagTaskID)
	assert.Equal(t, uint(0), idFlag)

	terminateCmd := findCommand(tasksSubCmds, "terminate")
	assert.NotNil(t, terminateCmd)
	require.NotNil(t, terminateCmd.Flags(), "terminate command flags should not be nil")
	idFlag, _ = terminateCmd.Flags().GetUint(flagTaskID)
	assert.Equal(t, uint(0), idFlag)

	listByInstanceCmd := findCommand(tasksSubCmds, "list-by-instance")
	assert.NotNil(t, listByInstanceCmd)
	require.NotNil(t, listByInstanceCmd.Flags(), "list-by-instance command flags should not be nil")
	instanceIDFlag, _ := listByInstanceCmd.Flags().GetUint(flagInstanceID)
	assert.Equal(t, uint(0), instanceIDFlag)
	actionFlag, _ := listByInstanceCmd.Flags().GetString(flagTaskAction)
	assert.Equal(t, "", actionFlag)
}

func TestGetTaskCmd(t *testing.T) {
	createdAt := time.Now()
	ownerID := models.AdminID

	tests := []struct {
		name      string
		args      []string
		setupTask func(s *test.Suite) models.Task // Func to create the task using the suite
		// expectedOutput will be set dynamically based on the created task
		expectedError string
	}{
		{
			name: "successful get",
			// Args will be updated dynamically if setupTask is present
			args: []string{"tasks", "get", "--id", "1", "-o", fmt.Sprintf("%d", ownerID)},
			setupTask: func(s *test.Suite) models.Task {
				task := models.Task{
					Model:     gorm.Model{CreatedAt: createdAt}, // ID will be set by DB
					Status:    models.TaskStatusCompleted,
					Action:    models.TaskActionCreateInstances,
					Logs:      "Task is running...",
					OwnerID:   ownerID,
					ProjectID: 1, // Example ProjectID, ensure project exists if task needs it
				}
				err := s.TaskRepo.Create(s.Context(), &task) // Use TaskRepo from suite
				s.Require().NoError(err)
				return task // task now has its ID
			},
		},
		{
			name:          "missing task id",
			args:          []string{"tasks", "get", "-o", fmt.Sprintf("%d", ownerID)},
			expectedError: "required flag(s) \"id\" not set",
		},
		{
			name:          "missing owner-id",
			args:          []string{"tasks", "get", "--id", "1"},
			expectedError: `required flag(s) "owner-id" not set`,
		},
		{
			name:          "invalid task id format",
			args:          []string{"tasks", "get", "--id", "abc", "-o", fmt.Sprintf("%d", ownerID)},
			expectedError: "invalid argument \"abc\" for \"-i, --id\" flag: strconv.ParseUint: parsing \"abc\": invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			currentArgs := make([]string, len(tt.args))
			copy(currentArgs, tt.args)
			var expectedOutput string

			if tt.setupTask != nil {
				createdTask := tt.setupTask(suite)
				// Dynamically set expected output based on created task ID and CreatedAt
				expectedOutput = fmt.Sprintf(`{
  "id": %d,
  "status": "%s",
  "action": "%s",
  "logs": "%s",
  "created_at": "%s"
}`, createdTask.ID, createdTask.Status, createdTask.Action, createdTask.Logs, createdTask.CreatedAt.Format("2006-01-02 15:04:05"))

				// Update args to use the dynamic ID
				foundIDFlag := false
				for i, arg := range currentArgs {
					if arg == "--id" && i+1 < len(currentArgs) {
						currentArgs[i+1] = fmt.Sprintf("%d", createdTask.ID)
						foundIDFlag = true
						break
					}
				}
				require.True(t, foundIDFlag, "--id flag not found in args for dynamic update")
			}

			originalClient := apiClient
			apiClient = suite.APIClient // Use the suite's actual API client
			defer func() { apiClient = originalClient }()

			buf := new(bytes.Buffer)
			originalStdout := os.Stdout
			rPipe, wPipe, _ := os.Pipe()
			os.Stdout = wPipe

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = io.Copy(buf, rPipe)
			}()

			cmd := setupTasksCommandTest(t)
			cmd.SetArgs(currentArgs)
			err := cmd.Execute()

			_ = wPipe.Close()
			os.Stdout = originalStdout
			wg.Wait()
			_ = rPipe.Close()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, expectedOutput, "expectedOutput should be set for successful test")
				assert.JSONEq(t, expectedOutput, buf.String())
			}
		})
	}
}

func TestListTasksCmd(t *testing.T) {
	createdAt := time.Now()
	ownerID := models.AdminID
	projectName := "list-tasks-project-cli"

	tests := []struct {
		name         string
		args         []string
		setupProject bool
		setupTasks   func(s *test.Suite, projectID uint) []models.Task
		// expectedOutput will be set dynamically based on created tasks
		expectedError string
	}{
		{
			name:         "successful list",
			args:         []string{"tasks", "list", "--project", projectName, "-o", fmt.Sprintf("%d", ownerID)},
			setupProject: true,
			setupTasks: func(s *test.Suite, projectID uint) []models.Task {
				// Create a proper payload for the terminate instance task
				terminatePayload, err := json.Marshal(types.DeleteInstanceRequest{
					InstanceID: 1, // Using a dummy instance ID since this is just for testing
				})
				s.Require().NoError(err)

				tasksToCreate := []models.Task{
					{Model: gorm.Model{CreatedAt: createdAt}, ProjectID: projectID, OwnerID: ownerID, Status: models.TaskStatusCompleted, Action: models.TaskActionCreateInstances, Logs: "Log for task 1"},
					{Model: gorm.Model{CreatedAt: createdAt.Add(time.Second)}, ProjectID: projectID, OwnerID: ownerID, Status: models.TaskStatusRunning, Action: models.TaskActionTerminateInstances, Error: "Some error for task 2", Payload: terminatePayload},
				}
				createdTasks := make([]models.Task, len(tasksToCreate))
				for i, task := range tasksToCreate {
					localTask := task // Create a local copy for the pointer
					err := s.TaskRepo.Create(s.Context(), &localTask)
					s.Require().NoError(err)
					createdTasks[i] = localTask // localTask now has ID and accurate CreatedAt from DB
				}
				return createdTasks
			},
		},
		{
			name:          "missing project name",
			args:          []string{"tasks", "list", "-o", fmt.Sprintf("%d", ownerID)},
			expectedError: "required flag(s) \"project\" not set",
		},
		{
			name:          "invalid page value",
			args:          []string{"tasks", "list", "--project", projectName, "--page", "invalid", "-o", fmt.Sprintf("%d", ownerID)},
			expectedError: "invalid argument \"invalid\" for \"-g, --page\" flag: strconv.ParseInt: parsing \"invalid\": invalid syntax",
		},
		{
			name:          "missing owner-id",
			args:          []string{"tasks", "list", "--project", projectName},
			expectedError: `required flag(s) "owner-id" not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			var projectID uint
			if tt.setupProject {
				project := models.Project{Name: projectName, OwnerID: ownerID, Description: "Test project for list tasks CLI"}
				err := suite.ProjectRepo.Create(suite.Context(), &project)
				suite.Require().NoError(err)
				projectID = project.ID
			}

			var expectedOutput string // Declare here to be set by setup
			if tt.setupTasks != nil {
				createdTasks := tt.setupTasks(suite, projectID)
				if tt.name == "successful list" {
					require.NotEmpty(t, createdTasks, "Should have created tasks for dynamic output")
					// Match the taskOutput struct from tasks.go (ID, Status, Action, Error, Created)
					taskOutputs := make([]taskOutput, len(createdTasks))
					for i, task := range createdTasks {
						taskOutputs[i] = taskOutput{
							ID:      task.ID,
							Status:  string(task.Status),
							Action:  string(task.Action),
							Error:   task.Error,
							Logs:    task.Logs, // Ensure logs are included if present
							Created: task.CreatedAt.Format("2006-01-02 15:04:05"),
						}
					}
					listOutput := taskListOutput{Tasks: taskOutputs}
					jsonBytes, err := json.MarshalIndent(listOutput, "", "  ")
					suite.Require().NoError(err)
					expectedOutput = string(jsonBytes)
				}
			}

			originalClient := apiClient
			apiClient = suite.APIClient
			defer func() { apiClient = originalClient }()

			buf := new(bytes.Buffer)
			originalStdout := os.Stdout
			rPipe, wPipe, _ := os.Pipe()
			os.Stdout = wPipe

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = io.Copy(buf, rPipe)
			}()

			cmd := setupTasksCommandTest(t)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			_ = wPipe.Close()
			os.Stdout = originalStdout
			wg.Wait()
			_ = rPipe.Close()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if expectedOutput != "" { // Only assert JSONEq if expectedOutput was set
					assert.JSONEq(t, expectedOutput, buf.String())
				}
			}
		})
	}
}

func TestTerminateTaskCmd(t *testing.T) {
	ownerID := models.AdminID
	tests := []struct {
		name          string
		args          []string
		setupTask     func(s *test.Suite) models.Task // Task to be terminated
		expectedError string
	}{
		{
			name: "successful terminate",
			// Args will be updated dynamically if setupTask is present
			args: []string{"tasks", "terminate", "--id", "1", "-o", fmt.Sprintf("%d", ownerID)},
			setupTask: func(s *test.Suite) models.Task {
				// Minimal valid payload for InstanceRequest to avoid worker unmarshal errors
				minimalPayload := `{"owner_id": %d, "project_name": "test-proj-terminate", "provider": "mock", "region": "mock", "size": "mock", "image": "mock", "ssh_key_name": "mock", "number_of_instances": 1, "action": "create"}`
				payloadBytes := []byte(fmt.Sprintf(minimalPayload, ownerID))

				task := models.Task{
					OwnerID:   ownerID,
					ProjectID: 1, // Assuming project 1 exists or is not strictly checked for this termination test
					Status:    models.TaskStatusRunning,
					Action:    models.TaskActionCreateInstances, // Keep action for consistency if needed
					Payload:   payloadBytes,                     // Add payload
				}
				err := s.TaskRepo.Create(s.Context(), &task) // Use TaskRepo from suite
				s.Require().NoError(err)

				// Ensure a project "test-proj-terminate" exists if the worker processing this task tries to fetch it.
				// This might be overkill if the worker's payload processing fails before project lookup.
				_ = s.ProjectRepo.Create(s.Context(), &models.Project{Name: "test-proj-terminate", OwnerID: ownerID, Description: "for terminate test"})

				return task // task now has its ID
			},
		},
		{
			name:          "missing task id",
			args:          []string{"tasks", "terminate", "-o", fmt.Sprintf("%d", ownerID)},
			expectedError: "required flag(s) \"id\" not set",
		},
		{
			name:          "missing owner-id",
			args:          []string{"tasks", "terminate", "--id", "1"},
			expectedError: `required flag(s) "owner-id" not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			var taskIDToTerminate uint = 1 // Default for args, will be dynamic if task is created
			currentArgs := make([]string, len(tt.args))
			copy(currentArgs, tt.args)

			if tt.setupTask != nil {
				createdTask := tt.setupTask(suite)
				taskIDToTerminate = createdTask.ID
				// Update args to use the dynamic ID
				foundIDFlag := false
				for i, arg := range currentArgs {
					if arg == "--id" && i+1 < len(currentArgs) {
						currentArgs[i+1] = fmt.Sprintf("%d", taskIDToTerminate)
						foundIDFlag = true
						break
					}
				}
				require.True(t, foundIDFlag, "--id flag not found in args for dynamic update")
			}

			originalClient := apiClient
			apiClient = suite.APIClient // Use the suite's actual API client
			defer func() { apiClient = originalClient }()

			buf := new(bytes.Buffer)
			originalStdout := os.Stdout
			rPipe, wPipe, _ := os.Pipe()
			os.Stdout = wPipe

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = io.Copy(buf, rPipe)
			}()

			cmd := setupTasksCommandTest(t)
			cmd.SetArgs(currentArgs)
			err := cmd.Execute()

			_ = wPipe.Close()
			os.Stdout = originalStdout
			wg.Wait()
			_ = rPipe.Close()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, buf.String(), fmt.Sprintf("Task ID %d termination request submitted successfully", taskIDToTerminate))
			}
		})
	}
}

func TestListInstanceTasksCmd(t *testing.T) {
	ownerID := models.AdminID // Assuming AdminID for CLI tests for simplicity
	projectName := "cli-list-inst-tasks-project"
	// instanceName1 := "cli-instance-1" // Removed as it's no longer used after removing Name field from Instance model
	// instanceName2 := "cli-instance-2" // Not used in this specific test plan yet

	tests := []struct {
		name                     string
		args                     []string
		setupProjectAndInstances func(s *test.Suite) (projectID uint, instanceID1 uint, instanceID2 uint)
		setupTasks               func(s *test.Suite, projectID, instanceID1, instanceID2 uint) []models.Task
		expectedOutputContains   []string // For partial output checks
		expectedTaskCount        int      // For verifying number of tasks returned
		noTasksMessage           string   // For asserting "No tasks found" message
		expectedError            string
	}{
		{
			name: "successful list for instance with specific action",
			// Args will be set dynamically
			setupProjectAndInstances: func(s *test.Suite) (uint, uint, uint) {
				proj := models.Project{Name: projectName, OwnerID: ownerID}
				s.Require().NoError(s.ProjectRepo.Create(s.Context(), &proj))
				inst1 := models.Instance{OwnerID: ownerID, ProjectID: proj.ID, ProviderID: "mock", Status: models.InstanceStatusReady}
				_, err := s.InstanceRepo.Create(s.Context(), &inst1)
				s.Require().NoError(err)
				inst2 := models.Instance{OwnerID: ownerID, ProjectID: proj.ID, ProviderID: "mock", Status: models.InstanceStatusReady}
				_, err = s.InstanceRepo.Create(s.Context(), &inst2)
				s.Require().NoError(err)
				return proj.ID, inst1.ID, inst2.ID
			},
			setupTasks: func(s *test.Suite, projectID, instanceID1, instanceID2 uint) []models.Task {
				tasksToCreate := []models.Task{
					{ProjectID: projectID, OwnerID: ownerID, InstanceID: instanceID1, Action: models.TaskActionCreateInstances, Status: models.TaskStatusCompleted, CreatedAt: time.Now().Add(-5 * time.Minute)},
					{ProjectID: projectID, OwnerID: ownerID, InstanceID: instanceID1, Action: models.TaskActionTerminateInstances, Status: models.TaskStatusPending, CreatedAt: time.Now().Add(-4 * time.Minute)},
					{ProjectID: projectID, OwnerID: ownerID, InstanceID: instanceID2, Action: models.TaskActionCreateInstances, Status: models.TaskStatusRunning, CreatedAt: time.Now().Add(-3 * time.Minute)},
				}
				created := make([]models.Task, len(tasksToCreate))
				for i, task := range tasksToCreate {
					localTask := task
					s.Require().NoError(s.TaskRepo.Create(s.Context(), &localTask))
					created[i] = localTask
				}
				return created
			},
			expectedTaskCount: 1,
		},
		{
			name: "list all tasks for an instance",
			// Args set dynamically
			setupProjectAndInstances: func(s *test.Suite) (uint, uint, uint) { // Same setup as above
				proj := models.Project{Name: projectName + "-all", OwnerID: ownerID}
				s.Require().NoError(s.ProjectRepo.Create(s.Context(), &proj))
				inst1 := models.Instance{OwnerID: ownerID, ProjectID: proj.ID, ProviderID: "mock", Status: models.InstanceStatusReady}
				_, err := s.InstanceRepo.Create(s.Context(), &inst1)
				s.Require().NoError(err)
				inst2 := models.Instance{OwnerID: ownerID, ProjectID: proj.ID, ProviderID: "mock", Status: models.InstanceStatusReady}
				_, err = s.InstanceRepo.Create(s.Context(), &inst2)
				s.Require().NoError(err)
				return proj.ID, inst1.ID, inst2.ID
			},
			setupTasks: func(s *test.Suite, projectID, instanceID1, instanceID2 uint) []models.Task { // Same tasks as above
				tasksToCreate := []models.Task{
					{ProjectID: projectID, OwnerID: ownerID, InstanceID: instanceID1, Action: models.TaskActionCreateInstances, Status: models.TaskStatusCompleted, CreatedAt: time.Now().Add(-5 * time.Minute)},
					{ProjectID: projectID, OwnerID: ownerID, InstanceID: instanceID1, Action: models.TaskActionTerminateInstances, Status: models.TaskStatusPending, CreatedAt: time.Now().Add(-4 * time.Minute)},
					{ProjectID: projectID, OwnerID: ownerID, InstanceID: instanceID2, Action: models.TaskActionCreateInstances, Status: models.TaskStatusRunning, CreatedAt: time.Now().Add(-3 * time.Minute)},
				}
				created := make([]models.Task, len(tasksToCreate))
				for i, task := range tasksToCreate {
					localTask := task
					s.Require().NoError(s.TaskRepo.Create(s.Context(), &localTask))
					created[i] = localTask
				}
				return created
			},
			expectedTaskCount: 2,
		},
		{
			name: "no tasks found for instance",
			// Args set dynamically
			setupProjectAndInstances: func(s *test.Suite) (uint, uint, uint) {
				proj := models.Project{Name: projectName + "-none", OwnerID: ownerID}
				s.Require().NoError(s.ProjectRepo.Create(s.Context(), &proj))
				inst1 := models.Instance{OwnerID: ownerID, ProjectID: proj.ID, ProviderID: "mock", Status: models.InstanceStatusReady}
				_, err := s.InstanceRepo.Create(s.Context(), &inst1)
				s.Require().NoError(err)
				return proj.ID, inst1.ID, 0 // No instance2 needed
			},
			setupTasks:     nil,                                                // No tasks created for this instance
			noTasksMessage: fmt.Sprintf("No tasks found for instance ID %%d."), // %d will be replaced with instanceID1
		},
		{
			name:          "missing instance-id flag",
			args:          []string{"tasks", "list-by-instance", "-o", fmt.Sprintf("%d", ownerID)},
			expectedError: `required flag(s) "instance-id" not set`,
		},
		{
			name:          "invalid instance-id value",
			args:          []string{"tasks", "list-by-instance", "--instance-id", "abc", "-o", fmt.Sprintf("%d", ownerID)},
			expectedError: "invalid argument \"abc\" for \"-I, --instance-id\" flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			var projectID, instanceID1, instanceID2 uint
			if tt.setupProjectAndInstances != nil {
				projectID, instanceID1, instanceID2 = tt.setupProjectAndInstances(suite)
			}

			var createdTasks []models.Task
			if tt.setupTasks != nil {
				createdTasks = tt.setupTasks(suite, projectID, instanceID1, instanceID2)
			}

			currentArgs := make([]string, len(tt.args))
			copy(currentArgs, tt.args)

			// Dynamically construct args if not fully provided (for success cases)
			if len(currentArgs) == 0 {
				baseArgs := []string{"tasks", "list-by-instance", "--instance-id", fmt.Sprintf("%d", instanceID1), "-o", fmt.Sprintf("%d", ownerID)}
				if tt.name == "successful list for instance with specific action" {
					baseArgs = append(baseArgs, "--action", string(models.TaskActionTerminateInstances))
				}
				// Add other flags like limit/offset if testing pagination specifically
				currentArgs = baseArgs
			}

			originalClient := apiClient
			apiClient = suite.APIClient
			defer func() { apiClient = originalClient }()

			buf := new(bytes.Buffer)
			originalStdout := os.Stdout
			rPipe, wPipe, _ := os.Pipe()
			os.Stdout = wPipe

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = io.Copy(buf, rPipe)
			}()

			cmd := setupTasksCommandTest(t)
			cmd.SetArgs(currentArgs)
			err := cmd.Execute()

			_ = wPipe.Close()
			os.Stdout = originalStdout
			wg.Wait()
			_ = rPipe.Close()

			outputStr := buf.String()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.noTasksMessage != "" {
					assert.Contains(t, outputStr, fmt.Sprintf(tt.noTasksMessage, instanceID1))
				} else {
					var resultOutput taskListOutput
					err = json.Unmarshal([]byte(outputStr), &resultOutput)
					require.NoError(t, err, "Failed to unmarshal JSON output: %s", outputStr)
					assert.Len(t, resultOutput.Tasks, tt.expectedTaskCount)

					// Further assertions on content if needed, e.g., checking specific task IDs or actions
					if tt.name == "successful list for instance with specific action" {
						require.Len(t, resultOutput.Tasks, 1)
						assert.Equal(t, string(models.TaskActionTerminateInstances), resultOutput.Tasks[0].Action)
						// Find the created task that matches this for ID check
						var expectedTaskID uint
						for _, ct := range createdTasks {
							if ct.InstanceID == instanceID1 && ct.Action == models.TaskActionTerminateInstances {
								expectedTaskID = ct.ID
								break
							}
						}
						assert.NotZero(t, expectedTaskID, "Expected task for assertion not found in setup data")
						assert.Equal(t, expectedTaskID, resultOutput.Tasks[0].ID)
					}
				}
			}
		})
	}
}
