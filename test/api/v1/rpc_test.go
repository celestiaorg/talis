package api_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/test"
)

var defaultProjectCreateParams = handlers.ProjectCreateParams{
	Name:        "test-project",
	Description: "A test project",
	Config:      `{"resources": {"cpu": 2, "memory": "4GB"}, "settings": {"env": "test"}}`,
	OwnerID:     models.AdminID,
}

var defaultTaskGetParams = handlers.TaskGetParams{
	TaskID:  1,
	OwnerID: models.AdminID,
}

type RPCError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	} `json:"error"`
	Success bool `json:"success"`
}

func TestProjectRPCMethods(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// Create a project using the API client
	project, err := suite.APIClient.CreateProject(suite.Context(), defaultProjectCreateParams)
	require.NoError(t, err)
	require.Equal(t, defaultProjectCreateParams.Name, project.Name)
	require.Equal(t, defaultProjectCreateParams.Description, project.Description)
	require.Equal(t, defaultProjectCreateParams.Config, project.Config)

	// Get project using RPC
	getParams := handlers.ProjectGetParams{
		Name:    defaultProjectCreateParams.Name,
		OwnerID: defaultProjectCreateParams.OwnerID,
	}
	retrievedProject, err := suite.APIClient.GetProject(suite.Context(), getParams)
	require.NoError(t, err)
	// Don't check ID since it's auto-incremented by the DB
	require.Equal(t, project.Name, retrievedProject.Name)
	require.Equal(t, project.Description, retrievedProject.Description)
	require.Equal(t, project.Config, retrievedProject.Config)

	// List projects using RPC
	listParams := handlers.ProjectListParams{Page: 1, OwnerID: defaultProjectCreateParams.OwnerID}
	listResponse, err := suite.APIClient.ListProjects(suite.Context(), listParams)
	require.NoError(t, err)
	require.NotEmpty(t, listResponse, "ListProjects should return projects")
	require.Equal(t, defaultProjectCreateParams.Name, listResponse[0].Name, "Project name mismatch in list")

	// Create another project
	secondProjectParams := handlers.ProjectCreateParams{
		Name:        "second-project",
		Description: "Another test project",
		Config:      defaultProjectCreateParams.Config,
		OwnerID:     defaultProjectCreateParams.OwnerID,
	}
	secondProject, err := suite.APIClient.CreateProject(suite.Context(), secondProjectParams)
	require.NoError(t, err)
	require.Equal(t, secondProjectParams.Name, secondProject.Name)

	// List projects again to verify we get both
	listParams = handlers.ProjectListParams{Page: 1, OwnerID: defaultProjectCreateParams.OwnerID}
	listResponse, err = suite.APIClient.ListProjects(suite.Context(), listParams)
	require.NoError(t, err)
	require.Equal(t, 2, len(listResponse))

	// Delete a project using RPC
	deleteParams := handlers.ProjectDeleteParams{Name: secondProject.Name, OwnerID: defaultProjectCreateParams.OwnerID}
	err = suite.APIClient.DeleteProject(suite.Context(), deleteParams)
	require.NoError(t, err)

	// List projects again to verify the delete worked
	listParams = handlers.ProjectListParams{Page: 1, OwnerID: defaultProjectCreateParams.OwnerID}
	listResponse, err = suite.APIClient.ListProjects(suite.Context(), listParams)
	require.NoError(t, err)
	require.Equal(t, 1, len(listResponse))
	// Don't check ID since it's auto-incremented by the DB
	require.Equal(t, defaultProjectCreateParams.Name, listResponse[0].Name)
}

func TestTaskRPCMethods(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// Create a project first
	project, err := suite.APIClient.CreateProject(suite.Context(), defaultProjectCreateParams)
	require.NoError(t, err)
	require.Equal(t, defaultProjectCreateParams.Name, project.Name)

	// Create a task directly using the repository for setup
	taskToCreate := models.Task{
		OwnerID:   models.AdminID,
		ProjectID: project.ID,
		Status:    models.TaskStatusCompleted,
		Action:    models.TaskActionCreateInstances,
	}
	err = suite.TaskRepo.Create(suite.Context(), &taskToCreate)
	require.NoError(t, err)

	// Get the task using RPC
	getParams := defaultTaskGetParams
	getParams.TaskID = taskToCreate.ID
	retrievedTask, err := suite.APIClient.GetTask(suite.Context(), getParams)
	require.NoError(t, err)
	require.Equal(t, taskToCreate.ID, retrievedTask.ID)
	require.Equal(t, taskToCreate.Status, retrievedTask.Status)

	// List tasks using RPC
	listParams := handlers.TaskListParams{ProjectName: project.Name, Page: 1, OwnerID: defaultProjectCreateParams.OwnerID}
	listResponse, err := suite.APIClient.ListTasks(suite.Context(), listParams)
	require.NoError(t, err)
	require.NotEmpty(t, listResponse, "ListTasks should return tasks")
	foundInList := false
	for _, listedTask := range listResponse {
		if listedTask.ID == taskToCreate.ID {
			foundInList = true
			require.Equal(t, taskToCreate.Status, listedTask.Status)
			break
		}
	}
	require.True(t, foundInList, "Created task not found in list by ID")

	// Create another task
	secondTaskToCreate := models.Task{
		OwnerID:   models.AdminID,
		ProjectID: project.ID,
		Status:    models.TaskStatusRunning,
		Action:    models.TaskActionTerminateInstances,
	}
	err = suite.TaskRepo.Create(suite.Context(), &secondTaskToCreate)
	require.NoError(t, err)

	// List tasks again to verify we get both
	listParams = handlers.TaskListParams{ProjectName: project.Name, Page: 1, OwnerID: defaultProjectCreateParams.OwnerID}
	listResponse, err = suite.APIClient.ListTasks(suite.Context(), listParams)
	require.NoError(t, err)
	foundTask1 := false
	foundTask2 := false
	for _, listedTask := range listResponse {
		if listedTask.ID == taskToCreate.ID {
			foundTask1 = true
		}
		if listedTask.ID == secondTaskToCreate.ID {
			foundTask2 = true
		}
	}
	require.True(t, foundTask1, "First created task not found in list")
	require.True(t, foundTask2, "Second created task not found in list")

	// Terminate the first task
	terminateParams := handlers.TaskTerminateParams{TaskID: taskToCreate.ID, OwnerID: models.AdminID}
	err = suite.APIClient.TerminateTask(suite.Context(), terminateParams)
	require.NoError(t, err)

	// Verify it's terminated
	getParams.TaskID = taskToCreate.ID
	terminatedTask, err := suite.APIClient.GetTask(suite.Context(), getParams)
	require.NoError(t, err)
	require.Equal(t, models.TaskStatusTerminated, terminatedTask.Status)
}

func TestClientUserMethods(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// Start with 0, then increment this variable whenever a user is successfully created
	// and decrement when a user is successfully deleted
	expectedUserCount := 0
	users, err := suite.APIClient.GetUsers(suite.Context(), handlers.UserGetParams{})
	require.NoError(t, err)
	require.NotNil(t, users)
	require.Empty(t, users.Users, "Expected no users in a fresh database")

	t.Run("CreateUser_Success", func(t *testing.T) {
		// Create first user
		newUser1, err := suite.APIClient.CreateUser(suite.Context(), defaultUser1)
		require.NoError(t, err)
		require.NotEmpty(t, newUser1.UserID, "User ID should not be empty")
		expectedUserCount++

		// Create second user
		newUser2, err := suite.APIClient.CreateUser(suite.Context(), defaultUser2)
		require.NoError(t, err)
		require.NotEmpty(t, newUser2.UserID, "User ID should not be empty")
		expectedUserCount++
	})

	t.Run("CreateUser_DuplicateUsername", func(t *testing.T) {
		// Try to create a user with the same username
		duplicateUser := defaultUser1
		_, err := suite.APIClient.CreateUser(suite.Context(), duplicateUser)
		require.Error(t, err, "Creating user with duplicate username should fail")
	})

	t.Run("GetUserByID_Success", func(t *testing.T) {
		// Create a user first
		newUser, err := suite.APIClient.CreateUser(suite.Context(), handlers.CreateUserParams{
			Username:     "testuser_getbyid",
			Email:        "getbyid@example.com",
			Role:         1,
			PublicSSHKey: "ssh-rsa TESTKEY",
		})
		require.NoError(t, err)
		expectedUserCount++

		// Get the user by ID
		resp, err := suite.APIClient.GetUserByID(suite.Context(), handlers.UserGetByIDParams{ID: newUser.UserID})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "testuser_getbyid", resp.Username)
		require.Equal(t, "getbyid@example.com", resp.Email)
	})

	t.Run("GetUserByUsername_Success", func(t *testing.T) {
		// Create a user first
		uniqueUsername := "unique_username_test"
		_, err := suite.APIClient.CreateUser(suite.Context(), handlers.CreateUserParams{
			Username:     uniqueUsername,
			Email:        "unique@example.com",
			Role:         1,
			PublicSSHKey: "ssh-rsa UNIQUEKEY",
		})
		require.NoError(t, err)
		expectedUserCount++

		// Get the user by username
		userResp, err := suite.APIClient.GetUsers(suite.Context(), handlers.UserGetParams{Username: uniqueUsername})
		require.NoError(t, err)
		require.NotNil(t, userResp.User)
		require.Equal(t, uniqueUsername, userResp.User.Username)
		require.Equal(t, "unique@example.com", userResp.User.Email)
	})

	t.Run("GetUserByUsername_NotFound", func(t *testing.T) {
		// Try to get a non-existent username
		_, err := suite.APIClient.GetUsers(suite.Context(), handlers.UserGetParams{Username: "nonexistent_user"})
		require.Error(t, err, "Getting non-existent username should return error")
	})

	t.Run("Get_All_Users", func(t *testing.T) {
		users, err := suite.APIClient.GetUsers(suite.Context(), handlers.UserGetParams{})
		require.NoError(t, err)
		require.Equal(t, expectedUserCount, len(users.Users))
	})

	t.Run("DeleteUser_Success", func(t *testing.T) {
		deletedUsername := "deleted_username_test"
		user, err := suite.APIClient.CreateUser(suite.Context(), handlers.CreateUserParams{
			Username:     deletedUsername,
			Email:        "deleted@example.com",
			Role:         1,
			PublicSSHKey: "ssh-rsa deletedKEY",
		})
		require.NoError(t, err)
		expectedUserCount++

		// Delete a existing user
		err = suite.APIClient.DeleteUser(suite.Context(), handlers.DeleteUserParams{ID: user.UserID})
		require.NoError(t, err)
		expectedUserCount--

		// Verify the user is actually deleted
		_, err = suite.APIClient.GetUserByID(suite.Context(), handlers.UserGetByIDParams{ID: user.UserID})
		require.Error(t, err, "User should no longer exist after deletion")

		// Delete an non existing user
		nonExistingUserID := uint(234)
		err = suite.APIClient.DeleteUser(suite.Context(), handlers.DeleteUserParams{ID: nonExistingUserID})
		require.Error(t, err)
	})
}

func TestProjectCreateErrors(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	t.Run("CreateProject_Success", func(t *testing.T) {
		// Create a project with valid parameters
		project, err := suite.APIClient.CreateProject(suite.Context(), defaultProjectCreateParams)
		require.NoError(t, err)
		require.Equal(t, defaultProjectCreateParams.Name, project.Name)
		require.Equal(t, defaultProjectCreateParams.Description, project.Description)
		require.Equal(t, defaultProjectCreateParams.Config, project.Config)
	})

	t.Run("CreateProject_DuplicateKey", func(t *testing.T) {
		// Try to create a project with the same name
		params := defaultProjectCreateParams
		_, err := suite.APIClient.CreateProject(suite.Context(), params)
		require.Error(t, err)

		// Parse the error response
		var rpcErr RPCError
		err = json.Unmarshal([]byte(err.Error()), &rpcErr)
		require.NoError(t, err)
		require.Equal(t, 400, rpcErr.Error.Code)
		require.Equal(t, "Project already exists", rpcErr.Error.Message)
		require.Contains(t, rpcErr.Error.Data, "UNIQUE constraint failed")
		require.False(t, rpcErr.Success)
	})

	t.Run("CreateProject_EmptyName", func(t *testing.T) {
		// Try to create a project with empty name
		invalidProject := defaultProjectCreateParams
		invalidProject.Name = ""
		_, err := suite.APIClient.CreateProject(suite.Context(), invalidProject)
		require.Error(t, err, "Creating project with empty name should fail")
		require.Contains(t, err.Error(), "project name is required", "Error message should indicate name is required")
	})

	t.Run("CreateProject_InvalidConfig", func(t *testing.T) {
		// Create a project with invalid JSON config
		invalidProject := defaultProjectCreateParams
		invalidProject.Name = "invalid-project"
		invalidProject.Config = `{"invalid": json}` // This should cause a JSON parsing error

		_, err := suite.APIClient.CreateProject(suite.Context(), invalidProject)
		require.Error(t, err, "Creating project with invalid config should fail")
		require.Contains(t, err.Error(), "invalid character", "Error message should indicate invalid JSON")
	})

	t.Run("CreateProject_InvalidOwnerID", func(t *testing.T) {
		// Try to create a project with non-existent owner ID
		invalidProject := defaultProjectCreateParams
		invalidProject.Name = "invalid-owner-project"
		invalidProject.OwnerID = 0 // Use 0 as invalid owner ID since it's not allowed

		_, err := suite.APIClient.CreateProject(suite.Context(), invalidProject)
		require.Error(t, err, "Creating project with invalid owner ID should fail")
		require.Contains(t, err.Error(), "owner_id is required", "Error message should indicate owner ID is required")
	})
}
