package repos

import (
	"context"

	"github.com/celestiaorg/talis/internal/db/models"
	"gorm.io/gorm"
)

// TaskRepository handles database operations for tasks
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new instance of TaskRepository
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

// Create creates a new task in the database
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// Get retrieves a task by ID from the database
func (r *TaskRepository) Get(ctx context.Context, id uint) (*models.Task, error) {
	var task models.Task
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByName retrieves a task by name within a project from the database
func (r *TaskRepository) GetByName(ctx context.Context, ownerID uint, projectID uint, name string) (*models.Task, error) {
	var task models.Task
	if err := r.db.WithContext(ctx).Where("owner_id = ? AND project_id = ? AND name = ?", ownerID, projectID, name).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListByProject retrieves all tasks for a specific project from the database with pagination
func (r *TaskRepository) ListByProject(ctx context.Context, ownerID uint, projectID uint, opts *models.ListOptions) ([]models.Task, error) {
	var tasks []models.Task
	query := r.db.WithContext(ctx).Where("owner_id = ? AND project_id = ?", ownerID, projectID)

	if !opts.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	if err := query.Limit(opts.Limit).Offset(opts.Offset).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// Delete deletes a task by ID from the database
func (r *TaskRepository) Delete(ctx context.Context, ownerID uint, id uint) error {
	return r.db.WithContext(ctx).Where("owner_id = ? AND id = ?", ownerID, id).Delete(&models.Task{}).Error
}

// UpdateStatus updates the status of a task in the database
func (r *TaskRepository) UpdateStatus(ctx context.Context, ownerID uint, id uint, status string) error {
	return r.db.WithContext(ctx).Model(&models.Task{}).Where("owner_id = ? AND id = ?", ownerID, id).Update("status", status).Error
}
