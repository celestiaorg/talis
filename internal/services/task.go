package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
)

// Task handles task-related operations
type Task struct {
	repo           *repos.TaskRepository
	projectService *Project
}

// NewTaskService creates a new instance of TaskService
func NewTaskService(repo *repos.TaskRepository, projectService *Project) *Task {
	return &Task{
		repo:           repo,
		projectService: projectService,
	}
}

// Create creates a new task
func (s *Task) Create(ctx context.Context, task *models.Task) error {
	return s.repo.Create(ctx, task)
}

// CreateBatch creates a batch of tasks
func (s *Task) CreateBatch(ctx context.Context, tasks []*models.Task) error {
	return s.repo.CreateBatch(ctx, tasks)
}

// Get retrieves a task by ID
func (s *Task) Get(ctx context.Context, ownerID uint, taskID uint) (*models.Task, error) {
	return s.repo.GetByID(ctx, ownerID, taskID)
}

// ListByProject retrieves all tasks for a specific project with pagination
func (s *Task) ListByProject(ctx context.Context, ownerID uint, projectName string, opts *models.ListOptions) ([]models.Task, error) {
	project, err := s.projectService.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByProject(ctx, ownerID, project.ID, opts)
}

// UpdateStatus updates the status of a task
func (s *Task) UpdateStatus(ctx context.Context, ownerID uint, taskID uint, status models.TaskStatus) error {
	return s.repo.UpdateStatus(ctx, ownerID, taskID, status)
}

// Update updates an existing task.
func (s *Task) Update(ctx context.Context, ownerID uint, task *models.Task) error {
	return s.repo.Update(ctx, ownerID, task)
}

// UpdateFailed updates a task as failed
func (s *Task) UpdateFailed(ctx context.Context, task *models.Task, errMsg, logMsg string) error {
	task.Status = models.TaskStatusFailed
	task.Error += fmt.Sprintf("\n%s", errMsg)
	task.Logs += fmt.Sprintf("\n%s", logMsg)
	return s.repo.Update(ctx, task.OwnerID, task)
}

// AddLogs appends logs to a task
func (s *Task) AddLogs(ctx context.Context, ownerID uint, taskID uint, logs string) error {
	task, err := s.Get(ctx, ownerID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Append new logs with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	updatedLogs := fmt.Sprintf("%s\n[%s] %s", task.Logs, timestamp, logs)

	task.Logs = updatedLogs
	return s.repo.Update(ctx, ownerID, task)
}

// SetResult updates a task with results data
func (s *Task) SetResult(ctx context.Context, ownerID uint, taskID uint, result json.RawMessage) error {
	task, err := s.Get(ctx, ownerID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Result = result
	return s.repo.Update(ctx, ownerID, task)
}

// SetError updates a task with error information
func (s *Task) SetError(ctx context.Context, ownerID uint, taskID uint, errMsg string) error {
	task, err := s.Get(ctx, ownerID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Error = errMsg
	task.Status = models.TaskStatusFailed

	if err := s.repo.Update(ctx, ownerID, task); err != nil {
		return err
	}

	// Send webhook notification if configured
	if task.WebhookURL != "" && !task.WebhookSent {
		if err := task.SendWebhook(); err != nil {
			// Log webhook error but don't fail the operation
			return nil
		}

		// Mark webhook as sent
		task.WebhookSent = true
		return s.repo.Update(ctx, ownerID, task)
	}

	return nil
}

// CompleteTask marks a task as completed and sends webhook if configured
func (s *Task) CompleteTask(ctx context.Context, ownerID uint, taskID uint, result json.RawMessage) error {
	task, err := s.Get(ctx, ownerID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Status = models.TaskStatusCompleted
	if result != nil {
		task.Result = result
	}

	if err := s.repo.Update(ctx, ownerID, task); err != nil {
		return err
	}

	// Send webhook notification if configured
	if task.WebhookURL != "" && !task.WebhookSent {
		if err := task.SendWebhook(); err != nil {
			// Log webhook error but don't fail the operation
			return nil
		}

		// Mark webhook as sent
		task.WebhookSent = true
		return s.repo.Update(ctx, ownerID, task)
	}

	return nil
}

// AcquireTaskLock attempts to lock a task for processing
// Returns true if the lock was acquired, false otherwise
func (s *Task) AcquireTaskLock(ctx context.Context, taskID uint) (bool, error) {
	return s.repo.AcquireTaskLock(ctx, taskID)
}

// ReleaseTaskLock releases a task lock
func (s *Task) ReleaseTaskLock(ctx context.Context, taskID uint) error {
	return s.repo.ReleaseTaskLock(ctx, taskID)
}

// RecoverStaleTasks finds tasks that were in progress when the system crashed
// and resets them to pending status with incremented attempts
func (s *Task) RecoverStaleTasks(ctx context.Context) (int64, error) {
	count, err := s.repo.RecoverStaleTasks(ctx)
	if err != nil {
		return 0, err
	}

	if count > 0 {
		logger.Infof("Recovered %d stale tasks that were in progress during system restart", count)
	}

	return count, nil
}

// GetSchedulableTasks retrieves tasks ready for the worker to process.
func (s *Task) GetSchedulableTasks(ctx context.Context, priority models.TaskPriority, limit int) ([]models.Task, error) {
	return s.repo.GetSchedulableTasks(ctx, priority, limit)
}

// ListTasksByInstanceID retrieves all tasks for a specific instance, with an optional action filter.
func (s *Task) ListTasksByInstanceID(ctx context.Context, ownerID uint, instanceID uint, actionFilter models.TaskAction, opts *models.ListOptions) ([]models.Task, error) {
	// Basic validation can be added here if needed, beyond what the repository does.
	// For example, checking if the instance itself exists or if the owner has access to the instance.
	// For now, we rely on the repository's ownerID and instanceID checks.
	return s.repo.ListByInstanceID(ctx, ownerID, instanceID, actionFilter, opts)
}
