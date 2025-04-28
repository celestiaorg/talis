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

// setupProjectCommands creates a new cobra command with project subcommands for testing
func setupProjectCommands() *cobra.Command {
	// Create a new root command for testing
	cmd := &cobra.Command{
		Use:   "talis",
		Short: "Talis CLI tool",
	}

	// Add the owner-id flag
	cmd.PersistentFlags().StringP(flagOwnerID, "o", "", "Owner ID for resources")

	// Add the projects command and its subcommands
	projectsCmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects",
	}
	cmd.AddCommand(projectsCmd)

	// Add create command
	createCmd := createProjectCmd
	createCmd.ResetFlags()
	createCmd.Flags().StringP("name", "n", "", "Project name")
	createCmd.Flags().StringP("description", "d", "", "Project description")
	createCmd.Flags().StringP("config", "c", "", "Project configuration")
	_ = createCmd.MarkFlagRequired("name")
	projectsCmd.AddCommand(createCmd)

	// Add get command
	getCmd := getProjectCmd
	getCmd.ResetFlags()
	getCmd.Flags().StringP("name", "n", "", "Project name")
	_ = getCmd.MarkFlagRequired("name")
	projectsCmd.AddCommand(getCmd)

	// Add list command
	listCmd := listProjectsCmd
	listCmd.ResetFlags()
	listCmd.Flags().IntP("page", "p", 1, "Page number for pagination")
	projectsCmd.AddCommand(listCmd)

	// Add delete command
	deleteCmd := deleteProjectCmd
	deleteCmd.ResetFlags()
	deleteCmd.Flags().StringP("name", "n", "", "Project name")
	_ = deleteCmd.MarkFlagRequired("name")
	projectsCmd.AddCommand(deleteCmd)

	// Add list-instances command
	listInstancesCmd := listProjectInstancesCmd
	listInstancesCmd.ResetFlags()
	listInstancesCmd.Flags().StringP("name", "n", "", "Project name")
	listInstancesCmd.Flags().IntP("page", "p", 1, "Page number for pagination")
	_ = listInstancesCmd.MarkFlagRequired("name")
	projectsCmd.AddCommand(listInstancesCmd)

	return cmd
}

func TestCreateProjectCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful create",
			args: []string{"projects", "create", "--name", "test-project", "--description", "Test project description", "-o", fmt.Sprintf("%d", models.AdminID)},
			expectedOutput: `{
  "name": ""
}`,
		},
		{
			name:          "missing project name",
			args:          []string{"projects", "create"},
			expectedError: "required flag(s) \"name\" not set",
		},
		{
			name:          "missing owner-id",
			args:          []string{"projects", "create", "--name", "test-project", "--description", "Test project description"},
			expectedError: `required flag(s) "owner-id" not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

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
			cmd := setupProjectCommands()
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
				// Check if the API returned a JSON response
				if buf.Len() > 0 {
					// Make sure JSON is valid
					var actual interface{}
					err = json.Unmarshal(buf.Bytes(), &actual)
					assert.NoError(t, err, "Response is not valid JSON: %s", buf.String())

					// Verify project was actually created in the database
					if tt.name == "successful create" {
						project, err := suite.ProjectRepo.GetByName(suite.Context(), models.AdminID, "test-project")
						assert.NoError(t, err)
						assert.Equal(t, "test-project", project.Name)
						assert.Equal(t, "Test project description", project.Description)
					}
				}
			}
		})
	}
}

func TestGetProjectCmd(t *testing.T) {
	createdAt := time.Now()

	tests := []struct {
		name           string
		args           []string
		mockProject    models.Project
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful get",
			args: []string{"projects", "get", "--name", "test-project", "-o", fmt.Sprintf("%d", models.AdminID)},
			mockProject: models.Project{
				Name:        "test-project",
				Description: "Test project description",
				OwnerID:     models.AdminID,
				CreatedAt:   createdAt,
			},
			expectedOutput: `{
  "name": "test-project",
  "description": "Test project description"
}`,
		},
		{
			name:          "missing project name",
			args:          []string{"projects", "get"},
			expectedError: "required flag(s) \"name\" not set",
		},
		{
			name:          "missing owner-id",
			args:          []string{"projects", "get", "--name", "test-project"},
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
			cmd := setupProjectCommands()
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

			// If no error was expected, fail the test if any error occurred.
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check output format
			if tt.expectedOutput != "" && buf.Len() > 0 {
				var response map[string]interface{}
				err = json.Unmarshal(buf.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, buf.String())
				}
				require.Equal(t, tt.mockProject.Name, response["name"].(string))
				if desc, ok := response["description"]; ok {
					require.Equal(t, tt.mockProject.Description, desc.(string))
				}
			}
		})
	}
}

func TestListProjectsCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockProjects   []models.Project
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful list",
			args: []string{"projects", "list", "-o", fmt.Sprintf("%d", models.AdminID)},
			mockProjects: []models.Project{
				{
					Name:        "project1",
					Description: "Description 1",
					OwnerID:     models.AdminID,
				},
				{
					Name:        "project2",
					Description: "Description 2",
					OwnerID:     models.AdminID,
				},
			},
			expectedOutput: `{
  "projects": []
}`,
		},
		{
			name: "successful list with pagination",
			args: []string{"projects", "list", "--page", "2", "-o", fmt.Sprintf("%d", models.AdminID)},
			mockProjects: []models.Project{
				{
					Name:        "project3",
					Description: "Description 3",
					OwnerID:     models.AdminID,
				},
			},
			expectedOutput: `{
  "projects": []
}`,
		},
		{
			name:          "invalid page value",
			args:          []string{"projects", "list", "--page", "invalid"},
			expectedError: "invalid argument \"invalid\" for \"-p, --page\" flag: strconv.ParseInt: parsing \"invalid\": invalid syntax",
		},
		{
			name:          "missing owner-id",
			args:          []string{"projects", "list"},
			expectedError: `required flag(s) "owner-id" not set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new test suite
			suite := test.NewSuite(t)
			defer suite.Cleanup()

			// Set up mock response in the database
			if tt.mockProjects != nil {
				for _, project := range tt.mockProjects {
					result := suite.DB.Create(&project)
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
			cmd := setupProjectCommands()
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

			// If no error was expected, fail the test if any error occurred.
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check output format
			if tt.expectedOutput != "" && buf.Len() > 0 {
				var response map[string]interface{}
				err = json.Unmarshal(buf.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, buf.String())
				}
				// Verify the output has a projects field that's an array
				projects, ok := response["projects"]
				require.True(t, ok, "Response doesn't contain a 'projects' field")
				_, ok = projects.([]interface{})
				require.True(t, ok, "The 'projects' field is not an array")
			}
		})
	}
}

func TestDeleteProjectCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockProject    models.Project
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful delete",
			args: []string{"projects", "delete", "--name", "test-project", "-o", fmt.Sprintf("%d", models.AdminID)},
			mockProject: models.Project{
				Name:        "test-project",
				Description: "Test project description",
			},
			expectedOutput: "Project 'test-project' deleted successfully",
		},
		{
			name:          "missing project name",
			args:          []string{"projects", "delete"},
			expectedError: "required flag(s) \"name\" not set",
		},
		{
			name:          "missing owner-id",
			args:          []string{"projects", "delete", "--name", "test-project"},
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
			cmd := setupProjectCommands()
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

			// If no error was expected, fail the test if any error occurred.
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check output format
			if tt.expectedOutput != "" && buf.Len() > 0 {
				assert.Contains(t, buf.String(), tt.expectedOutput)
			}
		})
	}
}

func TestListProjectInstancesCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		mockProject    models.Project
		mockInstances  []models.Instance
		expectedOutput string
		expectedError  string
	}{
		{
			name: "successful list instances",
			args: []string{"projects", "instances", "--name", "test-project", "-o", fmt.Sprintf("%d", models.AdminID)},
			mockProject: models.Project{
				Name:    "test-project",
				OwnerID: models.AdminID,
			},
			mockInstances: []models.Instance{
				{
					Name:     "instance1",
					Status:   models.InstanceStatusReady,
					PublicIP: "192.168.1.1",
					Region:   "us-east-1",
					Size:     "small",
					OwnerID:  models.AdminID,
				},
				{
					Name:    "instance2",
					Status:  models.InstanceStatusPending,
					Region:  "us-west-1",
					Size:    "medium",
					OwnerID: models.AdminID,
				},
			},
			expectedOutput: `{
  "instances": []
}`,
		},
		{
			name:          "missing project name",
			args:          []string{"projects", "instances"},
			expectedError: "required flag(s) \"name\" not set",
		},
		{
			name:          "missing owner-id",
			args:          []string{"projects", "instances", "--name", "test-project"},
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

				if tt.mockInstances != nil {
					for i := range tt.mockInstances {
						tt.mockInstances[i].ProjectID = tt.mockProject.ID
						result := suite.DB.Create(&tt.mockInstances[i])
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
			cmd := setupProjectCommands()
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

			// If no error was expected, fail the test if any error occurred.
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check output format
			if tt.expectedOutput != "" && buf.Len() > 0 {
				// Just check if the output is valid JSON with the expected structure
				var response map[string]interface{}
				err = json.Unmarshal(buf.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, buf.String())
				}

				// Verify the output has an instances field that's an array
				instances, ok := response["instances"]
				require.True(t, ok, "Response doesn't contain an 'instances' field")
				_, ok = instances.([]interface{})
				require.True(t, ok, "The 'instances' field is not an array")
			}
		})
	}
}

func TestProjectsCommand(t *testing.T) {
	cmd := projectsCmd

	// Test that the projects command has the expected subcommands
	subCmds := cmd.Commands()
	assert.Equal(t, 5, len(subCmds), "Expected 5 subcommands")

	// Verify the subcommand names
	var subCmdNames []string
	for _, c := range subCmds {
		subCmdNames = append(subCmdNames, c.Name())
	}

	// Expect create, list, delete, instances subcommands
	assert.Contains(t, subCmdNames, "create")
	assert.Contains(t, subCmdNames, "list")
	assert.Contains(t, subCmdNames, "delete")
	assert.Contains(t, subCmdNames, "instances")

	// Verify flags for create command
	createCmd := findCommand(subCmds, "create")
	assert.NotNil(t, createCmd)
	assert.True(t, createCmd.Flags().HasFlags())
	nameFlag, _ := createCmd.Flags().GetString("name")
	assert.Equal(t, "", nameFlag)
	descFlag, _ := createCmd.Flags().GetString("description")
	assert.Equal(t, "", descFlag)

	// Verify flags for list command
	listCmd := findCommand(subCmds, "list")
	assert.NotNil(t, listCmd)
	assert.True(t, listCmd.Flags().HasFlags())
	pageFlag, _ := listCmd.Flags().GetInt("page")
	assert.Equal(t, 1, pageFlag)

	// Verify flags for delete command
	deleteCmd := findCommand(subCmds, "delete")
	assert.NotNil(t, deleteCmd)
	assert.True(t, deleteCmd.Flags().HasFlags())

	// Verify flags for instances command
	instancesCmd := findCommand(subCmds, "instances")
	assert.NotNil(t, instancesCmd)
	assert.True(t, instancesCmd.Flags().HasFlags())
}

// Helper function to find a command by name
func findCommand(cmds []*cobra.Command, name string) *cobra.Command {
	for _, c := range cmds {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
