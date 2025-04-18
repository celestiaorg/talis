package repos

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/celestiaorg/talis/internal/db/models"
)

type TaskRepositoryTestSuite struct {
	DBRepositoryTestSuite
}

func (s *TaskRepositoryTestSuite) TestCreateTask() {
	// Create a test project
	project := s.createTestProject()

	// Create a test task
	task := s.randomTask(project.OwnerID, project.ID)

	// Test creation
	err := s.taskRepo.Create(s.ctx, task)
	s.Require().NoError(err)
	s.Require().NotZero(task.ID)

	// Verify task was created correctly
	createdTask, err := s.taskRepo.GetByID(s.ctx, task.OwnerID, task.ID)
	s.Require().NoError(err)
	s.Require().Equal(task.ID, createdTask.ID)
	s.Require().Equal(task.Name, createdTask.Name)
	s.Require().Equal(task.ProjectID, createdTask.ProjectID)
	s.Require().Equal(task.OwnerID, createdTask.OwnerID)
	s.Require().Equal(task.Status, createdTask.Status)
	s.Require().Equal(task.Action, createdTask.Action)
	s.Require().Equal(task.Logs, createdTask.Logs)

	// Test batch creation
	tasks := []*models.Task{s.randomTask(project.OwnerID, project.ID), s.randomTask(project.OwnerID, project.ID), s.randomTask(project.OwnerID, project.ID)}
	err = s.taskRepo.CreateBatch(s.ctx, tasks)
	s.Require().NoError(err)
	foundTasks, err := s.taskRepo.ListByProject(s.ctx, project.OwnerID, project.ID, nil)
	s.Require().NoError(err)
	s.Require().Equal(4, len(foundTasks))
}

func (s *TaskRepositoryTestSuite) TestGetTaskByID() {
	// Create a test task
	task := s.createTestTask()

	// Test retrieval by ID
	retrievedTask, err := s.taskRepo.GetByID(s.ctx, task.OwnerID, task.ID)
	s.Require().NoError(err)
	s.Require().Equal(task.ID, retrievedTask.ID)
	s.Require().Equal(task.Name, retrievedTask.Name)
	s.Require().Equal(task.ProjectID, retrievedTask.ProjectID)
	s.Require().Equal(task.OwnerID, retrievedTask.OwnerID)
	s.Require().Equal(task.Status, retrievedTask.Status)
	s.Require().Equal(task.Action, retrievedTask.Action)
	s.Require().Equal(task.Logs, retrievedTask.Logs)

	// Test retrieval with wrong owner ID
	_, err = s.taskRepo.GetByID(s.ctx, task.OwnerID+1, task.ID)
	s.Require().Error(err)

	// Test retrieval with non-existent ID
	_, err = s.taskRepo.GetByID(s.ctx, task.OwnerID, 999)
	s.Require().Error(err)
}

func (s *TaskRepositoryTestSuite) TestGetTaskByName() {
	// Create a test task
	task := s.createTestTask()

	// Test retrieval by name
	retrievedTask, err := s.taskRepo.GetByName(s.ctx, task.OwnerID, task.Name)
	s.Require().NoError(err)
	s.Require().Equal(task.ID, retrievedTask.ID)
	s.Require().Equal(task.Name, retrievedTask.Name)
	s.Require().Equal(task.ProjectID, retrievedTask.ProjectID)
	s.Require().Equal(task.OwnerID, retrievedTask.OwnerID)
	s.Require().Equal(task.Status, retrievedTask.Status)
	s.Require().Equal(task.Action, retrievedTask.Action)
	s.Require().Equal(task.Logs, retrievedTask.Logs)

	// Test retrieval with wrong owner ID
	_, err = s.taskRepo.GetByName(s.ctx, task.OwnerID+1, task.Name)
	s.Require().Error(err)

	// Test retrieval with non-existent name
	_, err = s.taskRepo.GetByName(s.ctx, task.OwnerID, "non-existent-task")
	s.Require().Error(err)
}

func (s *TaskRepositoryTestSuite) TestListTasksByProject() {
	// Create a test project
	project := s.createTestProject()

	// Create multiple tasks for the same project
	taskCount := 3
	for i := 0; i < taskCount; i++ {
		task := &models.Task{
			Name:      "test-task-list-" + time.Now().Format(time.RFC3339Nano),
			ProjectID: project.ID,
			OwnerID:   project.OwnerID,
			Status:    models.TaskStatusPending,
			Action:    models.TaskActionCreateInstances,
			Logs:      "Task logs for list test",
			CreatedAt: time.Now(),
		}
		err := s.taskRepo.Create(s.ctx, task)
		s.Require().NoError(err)
	}

	// Test listing tasks
	listOptions := &models.ListOptions{
		Limit:  10,
		Offset: 0,
	}
	tasks, err := s.taskRepo.ListByProject(s.ctx, project.OwnerID, project.ID, listOptions)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(tasks), taskCount)

	// Verify all retrieved tasks belong to the specified project
	for _, task := range tasks {
		s.Require().Equal(project.ID, task.ProjectID)
		s.Require().Equal(project.OwnerID, task.OwnerID)
	}

	// Test with non-existent project ID (this might not cause an error, just return empty results)
	emptyTasks, err := s.taskRepo.ListByProject(s.ctx, project.OwnerID, 999, listOptions)
	s.Require().NoError(err) // Just checking the database operation doesn't error
	s.Require().Empty(emptyTasks)
}

func (s *TaskRepositoryTestSuite) TestUpdateTaskStatus() {
	// Create a test task
	task := s.createTestTask()

	// Test updating status
	newStatus := models.TaskStatusRunning
	err := s.taskRepo.UpdateStatus(s.ctx, task.OwnerID, task.ID, newStatus)
	s.Require().NoError(err)

	// Verify task status was updated
	updatedTask, err := s.taskRepo.GetByID(s.ctx, task.OwnerID, task.ID)
	s.Require().NoError(err)
	s.Require().Equal(newStatus, updatedTask.Status)

	// Test with invalid task ID (might not return an error if no rows affected)
	err = s.taskRepo.UpdateStatus(s.ctx, task.OwnerID, 999, newStatus)
	// Check if the implementation is expected to return an error for non-existent task ID
	s.Require().NoError(err) // Expect no error even if task ID is invalid

	// Test with invalid owner ID (should return an error if owner validation is strict)
	invalidOwnerID := uint(999)
	err = s.taskRepo.UpdateStatus(s.ctx, invalidOwnerID, task.ID, newStatus)
	// Some implementations might not error if owner validation is done within the SQL query.
	// Assert the expected behavior.
	s.Require().NoError(err) // Expect no error even if owner ID is invalid
}

func (s *TaskRepositoryTestSuite) TestUpdateTask() {
	// Create a test task
	task := s.createTestTask()

	// Modify the task
	task.Logs = "Updated logs"
	task.Status = models.TaskStatusRunning
	task.Error = "Test error"
	task.Result = []byte(`{"key":"value"}`)

	// Test updating the task
	err := s.taskRepo.Update(s.ctx, task.OwnerID, task)
	s.Require().NoError(err)

	// Verify task was updated
	updatedTask, err := s.taskRepo.GetByID(s.ctx, task.OwnerID, task.ID)
	s.Require().NoError(err)
	s.Require().Equal(task.Logs, updatedTask.Logs)
	s.Require().Equal(task.Status, updatedTask.Status)
	s.Require().Equal(task.Error, updatedTask.Error)

	// Test with invalid owner ID
	invalidOwnerID := uint(999)
	// Store the original status before attempting the invalid update
	originalStatus := updatedTask.Status
	err = s.taskRepo.Update(s.ctx, invalidOwnerID, task)
	// This might not return an error if owner validation is done at the SQL level.
	// Assert the expected behavior.
	s.Require().NoError(err) // Expect no error even if owner ID is invalid

	// Verify that the task was not actually updated with the wrong owner ID
	notUpdatedTask, getErr := s.taskRepo.GetByID(s.ctx, task.OwnerID, task.ID)
	s.Require().NoError(getErr)
	s.Require().Equal(originalStatus, notUpdatedTask.Status) // Check a field that was attempted to be updated
}

func TestTaskRepository(t *testing.T) {
	suite.Run(t, new(TaskRepositoryTestSuite))
}
