package services

import (
	"context"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
)

// TaskService handles task-related operations
type TaskService struct {
	repo        *repos.TaskRepository
	projectRepo *repos.ProjectRepository
}

// NewTaskService creates a new instance of TaskService
func NewTaskService(repo *repos.TaskRepository, projectRepo *repos.ProjectRepository) *TaskService {
	return &TaskService{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

// Create creates a new task
func (s *TaskService) Create(ctx context.Context, ownerID uint, projectName string, task *models.Task) error {
	project, err := s.projectRepo.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return err
	}
	task.ProjectID = project.ID
	return s.repo.Create(ctx, task)
}

// GetByName retrieves a task by name within a project
func (s *TaskService) GetByName(ctx context.Context, ownerID uint, projectName string, taskName string) (*models.Task, error) {
	project, err := s.projectRepo.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByName(ctx, ownerID, project.ID, taskName)
}

// ListByProject retrieves all tasks for a specific project with pagination
func (s *TaskService) ListByProject(ctx context.Context, ownerID uint, projectName string, opts *models.ListOptions) ([]models.Task, error) {
	project, err := s.projectRepo.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByProject(ctx, ownerID, project.ID, opts)
}

// UpdateStatus updates the status of a task
func (s *TaskService) UpdateStatus(ctx context.Context, ownerID uint, projectName string, taskName string, status string) error {
	project, err := s.projectRepo.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return err
	}
	task, err := s.repo.GetByName(ctx, ownerID, project.ID, taskName)
	if err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, ownerID, task.ID, status)
}
