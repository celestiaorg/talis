package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
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
func (s *Task) Create(ctx context.Context, ownerID uint, projectID uint, task *models.Task) error {
	task.ProjectID = projectID
	task.OwnerID = ownerID
	return s.repo.Create(ctx, task)
}

// GetByName retrieves a task by name
func (s *Task) GetByName(ctx context.Context, ownerID uint, taskName string) (*models.Task, error) {
	return s.repo.GetByName(ctx, ownerID, taskName)
}

// GetByID retrieves a task by ID
func (s *Task) GetByID(ctx context.Context, ownerID uint, taskID uint) (*models.Task, error) {
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

// UpdateStatusByName updates the status of a task by name
func (s *Task) UpdateStatusByName(ctx context.Context, ownerID uint, taskName string, status models.TaskStatus) error {
	task, err := s.repo.GetByName(ctx, ownerID, taskName)
	if err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, ownerID, task.ID, status)
}

// Update updates an existing task.
func (s *Task) Update(ctx context.Context, ownerID uint, task *models.Task) error {
	return s.repo.Update(ctx, ownerID, task)
}

// AddLogs appends logs to a task
func (s *Task) AddLogs(ctx context.Context, ownerID uint, taskID uint, logs string) error {
	task, err := s.GetByID(ctx, ownerID, taskID)
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
	task, err := s.GetByID(ctx, ownerID, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Result = result
	return s.repo.Update(ctx, ownerID, task)
}

// SetError updates a task with error information
func (s *Task) SetError(ctx context.Context, ownerID uint, taskID uint, errMsg string) error {
	task, err := s.GetByID(ctx, ownerID, taskID)
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
	task, err := s.GetByID(ctx, ownerID, taskID)
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
