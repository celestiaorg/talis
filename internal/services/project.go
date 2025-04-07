package services

import (
	"context"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
)

// ProjectService handles project-related operations
type ProjectService struct {
	repo *repos.ProjectRepository
}

// NewProjectService creates a new instance of ProjectService
func NewProjectService(repo *repos.ProjectRepository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

// Create creates a new project
func (s *ProjectService) Create(ctx context.Context, project *models.Project) error {
	return s.repo.Create(ctx, project)
}

// GetByName retrieves a project by name
func (s *ProjectService) GetByName(ctx context.Context, ownerID uint, name string) (*models.Project, error) {
	return s.repo.GetByName(ctx, ownerID, name)
}

// List retrieves all projects with pagination
func (s *ProjectService) List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Project, error) {
	return s.repo.List(ctx, ownerID, opts)
}

// Delete deletes a project by name
func (s *ProjectService) Delete(ctx context.Context, ownerID uint, name string) error {
	return s.repo.Delete(ctx, ownerID, name)
}

// ListInstances retrieves all instances for a specific project
func (s *ProjectService) ListInstances(ctx context.Context, ownerID uint, projectName string) ([]models.Instance, error) {
	project, err := s.repo.GetByName(ctx, ownerID, projectName)
	if err != nil {
		return nil, err
	}
	return s.repo.ListInstances(ctx, project.ID)
}
