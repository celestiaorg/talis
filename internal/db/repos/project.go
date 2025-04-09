// Package repos provides database repository implementations
package repos

import (
	"context"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// ProjectRepository handles database operations for projects
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository creates a new instance of ProjectRepository
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

// Create creates a new project in the database
func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// Get retrieves a project by ID from the database
func (r *ProjectRepository) Get(ctx context.Context, id uint) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByName retrieves a project by name from the database
func (r *ProjectRepository) GetByName(ctx context.Context, ownerID uint, name string) (*models.Project, error) {
	var project models.Project
	query := r.db.WithContext(ctx).Where(models.Project{
		OwnerID: ownerID,
		Name:    name,
	})
	if err := query.First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// List retrieves all projects from the database with pagination
func (r *ProjectRepository) List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.WithContext(ctx).Where(models.Project{OwnerID: ownerID}).
		Limit(opts.Limit).Offset(opts.Offset).Find(&projects).Error
	return projects, err
}

// Delete deletes a project by name from the database
func (r *ProjectRepository) Delete(ctx context.Context, ownerID uint, name string) error {
	return r.db.WithContext(ctx).Where(models.Project{
		OwnerID: ownerID,
		Name:    name,
	}).Delete(&models.Project{}).Error
}

// ListInstances retrieves all instances for a specific project from the database
func (r *ProjectRepository) ListInstances(ctx context.Context, projectID uint, opts *models.ListOptions) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).Limit(opts.Limit).Offset(opts.Offset).
		Where(models.Instance{ProjectID: projectID}).Find(&instances).Error
	return instances, err
}
