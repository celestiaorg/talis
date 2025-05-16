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

	const instanceID = uint(1)

	// Create a test task
	task := s.randomTask(project.OwnerID, project.ID, instanceID)

	// Test creation
	err := s.taskRepo.Create(s.ctx, task)
	s.Require().NoError(err)
	s.Require().NotZero(task.ID)

	// Verify task was created correctly
	createdTask, err := s.taskRepo.GetByID(s.ctx, task.OwnerID, task.ID)
	s.Require().NoError(err)
	s.Require().Equal(task.ID, createdTask.ID)
	s.Require().Equal(task.ProjectID, createdTask.ProjectID)
	s.Require().Equal(task.OwnerID, createdTask.OwnerID)
	s.Require().Equal(task.Status, createdTask.Status)
	s.Require().Equal(task.Action, createdTask.Action)
	s.Require().Equal(task.Logs, createdTask.Logs)

	// Test batch creation
	tasks := []*models.Task{
		s.randomTask(project.OwnerID, project.ID, instanceID+1),
		s.randomTask(project.OwnerID, project.ID, instanceID+2),
		s.randomTask(project.OwnerID, project.ID, instanceID+3),
	}
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

func (s *TaskRepositoryTestSuite) TestListTasksByProject() {
	// Create a test project
	project := s.createTestProject()

	// Create multiple tasks for the same project
	taskCount := 3
	for i := 0; i < taskCount; i++ {
		task := &models.Task{
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

func (s *TaskRepositoryTestSuite) TestGetSchedulableTasks() {
	project := s.createTestProject() // Use a common project for these tasks
	ownerID := project.OwnerID
	now := time.Now()

	// Seed tasks with different statuses, error presence, and creation times
	tasksToCreate := []models.Task{
		// Should be excluded
		{ProjectID: project.ID, OwnerID: ownerID, Status: models.TaskStatusCompleted, CreatedAt: now.Add(-10 * time.Minute), Action: models.TaskActionCreateInstances, Priority: models.TaskPriorityHigh},
		{ProjectID: project.ID, OwnerID: ownerID, Status: models.TaskStatusTerminated, CreatedAt: now.Add(-9 * time.Minute), Action: models.TaskActionCreateInstances, Priority: models.TaskPriorityHigh},
		// Should be included - No Error, ordered by CreatedAt ASC
		{ProjectID: project.ID, OwnerID: ownerID, Status: models.TaskStatusPending, CreatedAt: now.Add(-8 * time.Minute), Action: models.TaskActionCreateInstances, Priority: models.TaskPriorityHigh}, // Expected 1st
		{ProjectID: project.ID, OwnerID: ownerID, Status: models.TaskStatusRunning, CreatedAt: now.Add(-7 * time.Minute), Action: models.TaskActionCreateInstances, Priority: models.TaskPriorityHigh}, // Expected 2nd
		{ProjectID: project.ID, OwnerID: ownerID, Status: models.TaskStatusPending, CreatedAt: now.Add(-6 * time.Minute), Action: models.TaskActionCreateInstances, Priority: models.TaskPriorityHigh}, // Expected 3rd
		// Should be included - With Error, ordered by CreatedAt ASC
		{ProjectID: project.ID, OwnerID: ownerID, Status: models.TaskStatusFailed, Error: "Some error", CreatedAt: now.Add(-5 * time.Minute), Action: models.TaskActionCreateInstances, Priority: models.TaskPriorityHigh},    // Expected 4th
		{ProjectID: project.ID, OwnerID: ownerID, Status: models.TaskStatusFailed, Error: "Another error", CreatedAt: now.Add(-4 * time.Minute), Action: models.TaskActionCreateInstances, Priority: models.TaskPriorityHigh}, // Expected 5th (if limit allows)
		// Should be excluded - Exceeds maxAttempts
		{ProjectID: project.ID, OwnerID: ownerID, Status: models.TaskStatusFailed, Error: "Too many attempts", Attempts: maxAttempts, CreatedAt: now.Add(-3 * time.Minute), Action: models.TaskActionCreateInstances, Priority: models.TaskPriorityHigh},
	}

	createdTasksWithIDs := make([]models.Task, 0, len(tasksToCreate))
	for _, task := range tasksToCreate {
		// Use a new variable in the loop to avoid capturing the loop variable's address
		newTask := task
		err := s.taskRepo.Create(s.ctx, &newTask)
		s.Require().NoError(err)
		createdTasksWithIDs = append(createdTasksWithIDs, newTask)
	}

	// --- Test Case 1: Limit = 4 ---
	limit := 4
	schedulableTasks, err := s.taskRepo.GetSchedulableTasks(s.ctx, models.TaskPriorityHigh, limit)
	s.Require().NoError(err)
	s.Require().Len(schedulableTasks, 4, "Expected 4 schedulable tasks for limit 4")
	if len(schedulableTasks) > 1 {
		s.Assert().True(schedulableTasks[0].CreatedAt.Before(schedulableTasks[1].CreatedAt) || schedulableTasks[0].CreatedAt.Equal(schedulableTasks[1].CreatedAt), "Tasks should be ordered by creation time if error status is same")
	}

	// --- Test Case 2: Limit = 2 (Testing limit and no-error ordering) ---
	limit = 2
	schedulableTasks, err = s.taskRepo.GetSchedulableTasks(s.ctx, models.TaskPriorityHigh, limit)
	s.Require().NoError(err)
	s.Require().Len(schedulableTasks, 2, "Expected 2 schedulable tasks for limit 2")
	if len(schedulableTasks) == 2 {
		s.Assert().Empty(schedulableTasks[0].Error, "First task should have no error")
		s.Assert().Empty(schedulableTasks[1].Error, "Second task should have no error")
		s.Assert().True(schedulableTasks[0].CreatedAt.Before(schedulableTasks[1].CreatedAt) || schedulableTasks[0].CreatedAt.Equal(schedulableTasks[1].CreatedAt))
	}

	// --- Test Case 3: Limit = 10 (Testing retrieval of all eligible tasks) ---
	limit = 10 // Higher than eligible tasks (5 are eligible: 3 no error, 2 with error)
	schedulableTasks, err = s.taskRepo.GetSchedulableTasks(s.ctx, models.TaskPriorityHigh, limit)
	s.Require().NoError(err)
	s.Require().Len(schedulableTasks, 5, "Expected all 5 eligible schedulable tasks")
	// Check that tasks with errors come after tasks without errors
	foundErrorTask := false
	for _, task := range schedulableTasks {
		if task.Error != "" {
			foundErrorTask = true
		} else {
			s.Assert().False(foundErrorTask, "Tasks without errors should appear before tasks with errors")
		}
	}

	// --- Test Case 4: Verify task with maxAttempts is excluded ---
	// This is implicitly tested if count for limit 10 is 5 (as task-max-attempts is not one of them).
	foundMaxAttemptsTask := false
	for _, task := range schedulableTasks {
		if task.Attempts >= maxAttempts { // Check actual name if needed, but this check is more direct
			// Find the original task with maxAttempts to compare ID if necessary
			for _, originalTask := range createdTasksWithIDs {
				if originalTask.Attempts >= maxAttempts && task.ID == originalTask.ID {
					foundMaxAttemptsTask = true
					break
				}
			}
		}
	}
	s.Assert().False(foundMaxAttemptsTask, "Task with max attempts should be excluded")
}

func (s *TaskRepositoryTestSuite) TestListByInstanceID() {
	// 1. Setup: Create a project and an instance for context
	ownerID := s.randomOwnerID()
	project := s.createTestProjectForOwner(ownerID)

	instance1 := s.createTestInstanceForOwner(ownerID)
	instance1.ProjectID = project.ID // Associate instance with project
	s.Require().NoError(s.instanceRepo.Update(s.ctx, ownerID, instance1.ID, instance1), "Failed to update instance1 with projectID")

	instance2 := s.createTestInstanceForOwner(ownerID)
	instance2.ProjectID = project.ID
	s.Require().NoError(s.instanceRepo.Update(s.ctx, ownerID, instance2.ID, instance2), "Failed to update instance2 with projectID")

	// 2. Create tasks: some for instance1, some for instance2, some with different actions
	// Tasks for instance1
	task1I1Create := s.createTestTaskForProject(ownerID, project.ID, instance1.ID)
	task1I1Create.Action = models.TaskActionCreateInstances
	s.Require().NoError(s.taskRepo.Update(s.ctx, ownerID, task1I1Create))

	time.Sleep(10 * time.Millisecond) // Ensure CreatedAt is different for ordering
	task2I1Terminate := s.createTestTaskForProject(ownerID, project.ID, instance1.ID)
	task2I1Terminate.Action = models.TaskActionTerminateInstances
	s.Require().NoError(s.taskRepo.Update(s.ctx, ownerID, task2I1Terminate))

	time.Sleep(10 * time.Millisecond)
	task3I1Create := s.createTestTaskForProject(ownerID, project.ID, instance1.ID)
	task3I1Create.Action = models.TaskActionCreateInstances
	s.Require().NoError(s.taskRepo.Update(s.ctx, ownerID, task3I1Create))

	// Task for instance2
	task1I2Create := s.createTestTaskForProject(ownerID, project.ID, instance2.ID)
	task1I2Create.Action = models.TaskActionCreateInstances
	s.Require().NoError(s.taskRepo.Update(s.ctx, ownerID, task1I2Create))

	// --- Test Case 1: List all tasks for instance1 ---
	tasksI1, err := s.taskRepo.ListByInstanceID(s.ctx, ownerID, instance1.ID, "", nil)
	s.Require().NoError(err)
	s.Require().Len(tasksI1, 3, "Should find 3 tasks for instance1")
	// Check order (default is created_at DESC)
	s.Assert().Equal(task3I1Create.ID, tasksI1[0].ID)
	s.Assert().Equal(task2I1Terminate.ID, tasksI1[1].ID)
	s.Assert().Equal(task1I1Create.ID, tasksI1[2].ID)

	// --- Test Case 2: List tasks for instance1 with action filter "create_instances" ---
	tasksI1CreateAction, err := s.taskRepo.ListByInstanceID(s.ctx, ownerID, instance1.ID, models.TaskActionCreateInstances, nil)
	s.Require().NoError(err)
	s.Require().Len(tasksI1CreateAction, 2, "Should find 2 create_instances tasks for instance1")
	s.Assert().Contains([]uint{task3I1Create.ID, task1I1Create.ID}, tasksI1CreateAction[0].ID)
	s.Assert().Contains([]uint{task3I1Create.ID, task1I1Create.ID}, tasksI1CreateAction[1].ID)

	// --- Test Case 3: List tasks for instance1 with action filter "terminate_instances" ---
	tasksI1TerminateAction, err := s.taskRepo.ListByInstanceID(s.ctx, ownerID, instance1.ID, models.TaskActionTerminateInstances, nil)
	s.Require().NoError(err)
	s.Require().Len(tasksI1TerminateAction, 1, "Should find 1 terminate_instances task for instance1")
	s.Assert().Equal(task2I1Terminate.ID, tasksI1TerminateAction[0].ID)

	// --- Test Case 4: Pagination - Limit 1, Offset 1 for instance1 ---
	listOpts := &models.ListOptions{Limit: 1, Offset: 1}
	tasksI1Paginated, err := s.taskRepo.ListByInstanceID(s.ctx, ownerID, instance1.ID, "", listOpts)
	s.Require().NoError(err)
	s.Require().Len(tasksI1Paginated, 1, "Should find 1 task with limit 1, offset 1")
	s.Assert().Equal(task2I1Terminate.ID, tasksI1Paginated[0].ID) // Second most recent

	// --- Test Case 5: No tasks for a non-existent instance ID ---
	noTasks, err := s.taskRepo.ListByInstanceID(s.ctx, ownerID, 9999, "", nil)
	s.Require().NoError(err)
	s.Require().Empty(noTasks, "Should find no tasks for a non-existent instance ID")

	// --- Test Case 6: Wrong owner ID ---
	wrongOwnerTasks, err := s.taskRepo.ListByInstanceID(s.ctx, ownerID+123, instance1.ID, "", nil)
	// Expect an error due to ValidateOwnerID, or empty if query just returns no results.
	// Current ValidateOwnerID is a soft check, so query will proceed and return empty.
	s.Require().NoError(err) // If ValidateOwnerID becomes strict, this should be s.Require().Error(err)
	s.Require().Empty(wrongOwnerTasks)

	// --- Test Case 7: instanceID is zero (should error out based on repo logic) ---
	_, err = s.taskRepo.ListByInstanceID(s.ctx, ownerID, 0, "", nil)
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "instanceID cannot be zero")
}

func TestTaskRepository(t *testing.T) {
	suite.Run(t, new(TaskRepositoryTestSuite))
}
