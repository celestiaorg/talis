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

	// Create a project directly in the database
	project := models.Project{
		OwnerID:     models.AdminID,
		Name:        defaultProjectCreateParams.Name,
		Description: defaultProjectCreateParams.Description,
		Config:      defaultProjectCreateParams.Config,
	}
	err := suite.ProjectRepo.Create(suite.Context(), &project)
	require.NoError(t, err)

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
	secondProject := models.Project{
		OwnerID:     models.AdminID,
		Name:        "second-project",
		Description: "Another test project",
		Config:      defaultProjectCreateParams.Config,
	}
	err = suite.ProjectRepo.Create(suite.Context(), &secondProject)
	require.NoError(t, err)

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
	project := models.Project{
		OwnerID:     models.AdminID,
		Name:        defaultProjectCreateParams.Name,
		Description: defaultProjectCreateParams.Description,
		Config:      defaultProjectCreateParams.Config,
	}
	err := suite.ProjectRepo.Create(suite.Context(), &project)
	require.NoError(t, err)

	// Create a task directly in the DB since we don't have a Create RPC method
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
		Status:      "running",
	}
	err = suite.APIClient.UpdateTaskStatus(suite.Context(), updateParams)
	require.NoError(t, err)

	// Get the updated task to verify the status change
	retrievedTask, err = suite.APIClient.GetTask(suite.Context(), getParams)
	require.NoError(t, err)
	require.Equal(t, models.TaskStatusRunning, retrievedTask.Status)

	// Abort a task using RPC
	abortParams := handlers.TaskAbortParams{
		ProjectName: project.Name,
		TaskName:    task.Name,
	}
	err = suite.APIClient.AbortTask(suite.Context(), abortParams)
	require.NoError(t, err)

	// Verify the task is now aborted
	retrievedTask, err = suite.APIClient.GetTask(suite.Context(), getParams)
	require.NoError(t, err)
	require.Equal(t, "aborted", retrievedTask.Status.String())
}
