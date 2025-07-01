package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/types"
)

// TestSetup sets up an in-memory database and repositories for testing
type TestSetup struct {
	DB              *gorm.DB
	InstanceRepo    *repos.InstanceRepository
	TaskRepo        *repos.TaskRepository
	ProjectRepo     *repos.ProjectRepository
	InstanceService *Instance
	TaskService     *Task
	ProjectService  *Project
	ctx             context.Context
}

// NewTestSetup creates a new test setup with in-memory database
func NewTestSetup(t *testing.T) *TestSetup {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "Failed to create in-memory database")

	// Run migrations
	err = db.AutoMigrate(
		&models.Instance{},
		&models.User{},
		&models.Project{},
		&models.Task{},
	)
	assert.NoError(t, err, "Failed to run migrations")

	// Create real repositories
	instanceRepo := repos.NewInstanceRepository(db)
	taskRepo := repos.NewTaskRepository(db)
	projectRepo := repos.NewProjectRepository(db)

	// Create real services
	projectService := NewProjectService(projectRepo)
	taskService := NewTaskService(taskRepo, projectService)
	instanceService := NewInstanceService(instanceRepo, taskService, projectService)

	return &TestSetup{
		DB:              db,
		InstanceRepo:    instanceRepo,
		TaskRepo:        taskRepo,
		ProjectRepo:     projectRepo,
		InstanceService: instanceService,
		TaskService:     taskService,
		ProjectService:  projectService,
		ctx:             context.Background(),
	}
}

// CleanUp cleans up resources after test
func (ts *TestSetup) CleanUp() {
	sqlDB, err := ts.DB.DB()
	if err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
}

func TestInstanceService_CreateInstance_SetsTaskInstanceID(t *testing.T) {
	// Create test setup with real in-memory database
	ts := NewTestSetup(t)
	defer ts.CleanUp()

	// Save original env var
	originalSSHKeyName := os.Getenv(constants.EnvTalisSSHKeyName)
	defer func() {
		err := os.Setenv(constants.EnvTalisSSHKeyName, originalSSHKeyName)
		if err != nil {
			t.Logf("Failed to restore %s: %v", constants.EnvTalisSSHKeyName, err)
		}
	}()

	// Set SSH key name env var for tests
	err := os.Setenv(constants.EnvTalisSSHKeyName, "test-key")
	assert.NoError(t, err)

	// Test data
	ownerID := uint(1)
	projectName := "test-project"
	projectID := uint(10)

	// Create a test project
	project := &models.Project{
		OwnerID: ownerID,
		Name:    projectName,
		Model:   gorm.Model{ID: projectID},
	}
	err = ts.ProjectRepo.Create(ts.ctx, project)
	assert.NoError(t, err)

	instanceReqs := []types.InstanceRequest{
		{
			OwnerID: ownerID, ProjectName: projectName, Provider: models.ProviderDO,
			Region: "nyc1", Size: "s-1vcpu-1gb", Image: "ubuntu-20-04-x64",
			NumberOfInstances: 1, Action: "create",
			Volumes: []types.VolumeConfig{{Name: "vol1", SizeGB: 10, MountPoint: "/mnt/vol1"}},
		},
	}

	// Create instances
	actualCreatedInstances, err := ts.InstanceService.CreateInstance(ts.ctx, instanceReqs)
	assert.NoError(t, err)
	assert.NotNil(t, actualCreatedInstances)
	assert.Len(t, actualCreatedInstances, 1)
	assert.NotZero(t, actualCreatedInstances[0].ID)

	// Verify instance was created with correct data
	expectedInstanceID := actualCreatedInstances[0].ID
	instance, err := ts.InstanceRepo.Get(ts.ctx, ownerID, expectedInstanceID)
	assert.NoError(t, err)
	assert.Equal(t, ownerID, instance.OwnerID)
	assert.Equal(t, projectID, instance.ProjectID)
	assert.Equal(t, models.InstanceStatusPending, instance.Status)

	// Get all tasks for this instance
	tasks, err := ts.TaskRepo.ListByInstanceID(ts.ctx, ownerID, expectedInstanceID, "", nil)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)

	// Verify task has correct instance ID
	assert.Equal(t, expectedInstanceID, tasks[0].InstanceID)

	// Verify task payload contains instance ID
	var taskPayload types.InstanceRequest
	err = json.Unmarshal(tasks[0].Payload, &taskPayload)
	assert.NoError(t, err)

	// Add debug output
	fmt.Printf("DEBUG TEST: Task InstanceID in DB: %d\n", tasks[0].InstanceID)
	fmt.Printf("DEBUG TEST: Task Payload InstanceID: %d\n", taskPayload.InstanceID)
	fmt.Printf("DEBUG TEST: Expected InstanceID: %d\n", expectedInstanceID)

	assert.Equal(t, expectedInstanceID, taskPayload.InstanceID)
}

func TestInstanceService_Terminate_SetsTaskInstanceID(t *testing.T) {
	// Create test setup with real in-memory database
	ts := NewTestSetup(t)
	defer ts.CleanUp()

	// Test data
	ownerID := uint(1)
	projectName := "test-project-term"
	projectID := uint(11)

	// Create a test project
	project := &models.Project{
		OwnerID: ownerID,
		Name:    projectName,
		Model:   gorm.Model{ID: projectID},
	}
	err := ts.ProjectRepo.Create(ts.ctx, project)
	assert.NoError(t, err)

	// Create test instances
	instance1 := &models.Instance{
		OwnerID:   ownerID,
		ProjectID: projectID,
		Status:    models.InstanceStatusReady,
	}
	instance2 := &models.Instance{
		OwnerID:   ownerID,
		ProjectID: projectID,
		Status:    models.InstanceStatusReady,
	}

	instances, err := ts.InstanceRepo.CreateBatch(ts.ctx, []*models.Instance{instance1, instance2})
	assert.NoError(t, err)
	assert.Len(t, instances, 2)

	instanceIDs := []uint{instances[0].ID, instances[1].ID}

	// Terminate instances
	err = ts.InstanceService.Terminate(ts.ctx, ownerID, projectName, instanceIDs)
	assert.NoError(t, err)

	// Verify termination tasks were created
	tasks, err := ts.TaskRepo.ListByProject(ts.ctx, ownerID, projectID, nil)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)

	// Verify tasks have correct properties
	for _, task := range tasks {
		assert.Equal(t, ownerID, task.OwnerID)
		assert.Equal(t, projectID, task.ProjectID)
		assert.Equal(t, models.TaskStatusPending, task.Status)
		assert.Equal(t, models.TaskActionTerminateInstances, task.Action)

		// Verify instance ID is in instanceIDs
		assert.Contains(t, instanceIDs, task.InstanceID)

		var payload types.DeleteInstanceRequest
		err = json.Unmarshal(task.Payload, &payload)
		assert.NoError(t, err)
		assert.Equal(t, task.InstanceID, payload.InstanceID)
	}
}
