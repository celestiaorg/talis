package api_test

import (
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
}

var defaultTaskGetParams = handlers.TaskGetParams{
	ProjectName: "test-project",
	TaskName:    "test-task",
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
		Name: defaultProjectCreateParams.Name,
	}
	retrievedProject, err := suite.APIClient.GetProject(suite.Context(), getParams)
	require.NoError(t, err)
	// Don't check ID since it's auto-incremented by the DB
	require.Equal(t, project.Name, retrievedProject.Name)
	require.Equal(t, project.Description, retrievedProject.Description)
	require.Equal(t, project.Config, retrievedProject.Config)

	// List projects using RPC
	listParams := handlers.ProjectListParams{
		Page: 1,
	}
	projects, err := suite.APIClient.ListProjects(suite.Context(), listParams)
	require.NoError(t, err)
	require.NotEmpty(t, projects)
	require.Equal(t, 1, len(projects))
	// Don't check ID since it's auto-incremented by the DB
	require.Equal(t, project.Name, projects[0].Name)

	// Create another project
	secondProjectParams := handlers.ProjectCreateParams{
		Name:        "second-project",
		Description: "Another test project",
		Config:      defaultProjectCreateParams.Config,
	}
	secondProject, err := suite.APIClient.CreateProject(suite.Context(), secondProjectParams)
	require.NoError(t, err)
	require.Equal(t, secondProjectParams.Name, secondProject.Name)

	// List projects again to verify we get both
	projects, err = suite.APIClient.ListProjects(suite.Context(), listParams)
	require.NoError(t, err)
	require.Equal(t, 2, len(projects))

	// Delete a project using RPC
	deleteParams := handlers.ProjectDeleteParams{
		Name: secondProject.Name,
	}
	err = suite.APIClient.DeleteProject(suite.Context(), deleteParams)
	require.NoError(t, err)

	// List projects again to verify the delete worked
	projects, err = suite.APIClient.ListProjects(suite.Context(), listParams)
	require.NoError(t, err)
	require.Equal(t, 1, len(projects))
	// Don't check ID since it's auto-incremented by the DB
	require.Equal(t, defaultProjectCreateParams.Name, projects[0].Name)
}

func TestTaskRPCMethods(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// Create a project first
	project, err := suite.APIClient.CreateProject(suite.Context(), defaultProjectCreateParams)
	require.NoError(t, err)
	require.Equal(t, defaultProjectCreateParams.Name, project.Name)

	// Since we don't have a CreateTask API client method, we'll need to use the repository directly
	// TODO: Add CreateTask API client method
	task := models.Task{
		OwnerID:   models.AdminID,
		ProjectID: project.ID,
		Name:      "test-task",
		Status:    models.TaskStatusPending,
	}
	err = suite.TaskRepo.Create(suite.Context(), &task)
	require.NoError(t, err)

	// Get the task using RPC
	getParams := defaultTaskGetParams
	retrievedTask, err := suite.APIClient.GetTask(suite.Context(), getParams)
	require.NoError(t, err)
	// Don't check ID since it's auto-incremented by the DB
	require.Equal(t, task.Name, retrievedTask.Name)
	// ProjectID is also database-dependent, so skip the comparison
	require.Equal(t, task.Status, retrievedTask.Status)

	// List tasks using RPC
	listParams := handlers.TaskListParams{
		ProjectName: project.Name,
		Page:        1,
	}
	tasks, err := suite.APIClient.ListTasks(suite.Context(), listParams)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)
	require.Equal(t, 1, len(tasks))
	// Don't check ID since it's auto-incremented by the DB
	require.Equal(t, task.Name, tasks[0].Name)

	// Create another task
	// TODO: Add CreateTask API client method
	secondTask := models.Task{
		OwnerID:   models.AdminID,
		ProjectID: project.ID,
		Name:      "second-task",
		Status:    models.TaskStatusPending,
	}
	err = suite.TaskRepo.Create(suite.Context(), &secondTask)
	require.NoError(t, err)

	// List tasks again to verify we get both
	tasks, err = suite.APIClient.ListTasks(suite.Context(), listParams)
	require.NoError(t, err)
	require.Equal(t, 2, len(tasks))

	// Update a task status using RPC
	updateParams := handlers.TaskUpdateStatusParams{
		ProjectName: project.Name,
		TaskName:    task.Name,
		Status:      models.TaskStatusRunning,
	}
	err = suite.APIClient.UpdateTaskStatus(suite.Context(), updateParams)
	require.NoError(t, err)

	// Get the updated task to verify the status change
	retrievedTask, err = suite.APIClient.GetTask(suite.Context(), getParams)
	require.NoError(t, err)
	require.Equal(t, models.TaskStatusRunning, retrievedTask.Status)

	// Abort a task using RPC
	terminateParams := handlers.TaskTerminateParams{
		ProjectName: project.Name,
		TaskName:    task.Name,
	}
	err = suite.APIClient.TerminateTask(suite.Context(), terminateParams)
	require.NoError(t, err)

	// Verify the task is now terminated
	retrievedTask, err = suite.APIClient.GetTask(suite.Context(), getParams)
	require.NoError(t, err)
	require.Equal(t, models.TaskStatusTerminated, retrievedTask.Status)
}
