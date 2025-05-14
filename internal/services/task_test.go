package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

func TestTaskService_ListTasksByInstanceID(t *testing.T) {
	// Create test setup with in-memory database
	ts := NewTestSetup(t)
	defer ts.CleanUp()

	// Test data
	ownerID := uint(1)
	projectID := uint(10)
	projectName := "test-project"

	// Create a test project
	project := &models.Project{
		OwnerID: ownerID,
		Name:    projectName,
		Model:   gorm.Model{ID: projectID},
	}
	err := ts.ProjectRepo.Create(ts.ctx, project)
	assert.NoError(t, err)

	// Create a test instance
	instance := &models.Instance{
		OwnerID:   ownerID,
		ProjectID: projectID,
		Status:    models.InstanceStatusReady,
	}
	createdInstance, err := ts.InstanceRepo.Create(ts.ctx, instance)
	assert.NoError(t, err)
	instanceID := createdInstance.ID

	// Create two tasks for this instance with different actions
	task1 := &models.Task{
		OwnerID:    ownerID,
		ProjectID:  projectID,
		InstanceID: instanceID,
		Status:     models.TaskStatusPending,
		Action:     models.TaskActionCreateInstances,
	}
	err = ts.TaskRepo.Create(ts.ctx, task1)
	assert.NoError(t, err)

	task2 := &models.Task{
		OwnerID:    ownerID,
		ProjectID:  projectID,
		InstanceID: instanceID,
		Status:     models.TaskStatusPending,
		Action:     models.TaskActionTerminateInstances,
	}
	err = ts.TaskRepo.Create(ts.ctx, task2)
	assert.NoError(t, err)

	// Test case 1: List all tasks for instance (no filter)
	tasks, err := ts.TaskService.ListTasksByInstanceID(ts.ctx, ownerID, instanceID, "", nil)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2, "Should find all tasks for the instance")

	// Test case 2: List tasks with specific action filter
	tasks, err = ts.TaskService.ListTasksByInstanceID(ts.ctx, ownerID, instanceID, models.TaskActionCreateInstances, nil)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1, "Should find one task with CreateInstances action")
	assert.Equal(t, models.TaskActionCreateInstances, tasks[0].Action)

	// Test case 3: Using pagination options
	opts := &models.ListOptions{Limit: 1, Offset: 0}
	tasks, err = ts.TaskService.ListTasksByInstanceID(ts.ctx, ownerID, instanceID, "", opts)
	assert.NoError(t, err)
	assert.Len(t, tasks, 1, "Should respect pagination limit")
}

func TestTaskService_ListTasksByInstanceID_InstanceNotFound(t *testing.T) {
	// Create test setup with in-memory database
	ts := NewTestSetup(t)
	defer ts.CleanUp()

	// Test data
	ownerID := uint(1)
	nonExistentInstanceID := uint(9999)

	// List tasks for a non-existent instance
	tasks, err := ts.TaskService.ListTasksByInstanceID(ts.ctx, ownerID, nonExistentInstanceID, "", nil)
	assert.NoError(t, err) // This should not error, just return empty results
	assert.Len(t, tasks, 0, "Should return empty result for non-existent instance")
}
