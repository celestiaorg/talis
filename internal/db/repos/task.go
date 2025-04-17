package repos

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
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

// CreateBatch creates a batch of tasks in the database
func (r *TaskRepository) CreateBatch(ctx context.Context, tasks []*models.Task) error {
	return r.db.WithContext(ctx).CreateInBatches(tasks, 100).Error
}

// GetByID retrieves a task by ID from the database
func (r *TaskRepository) GetByID(ctx context.Context, ownerID uint, id uint) (*models.Task, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	var task models.Task
	if err := r.db.WithContext(ctx).
		Where(models.Task{
			Model:   gorm.Model{ID: id},
			OwnerID: ownerID,
		}).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByName retrieves a task by name within a project from the database
func (r *TaskRepository) GetByName(ctx context.Context, ownerID uint, name string) (*models.Task, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	var task models.Task
	err := r.db.WithContext(ctx).Where(models.Task{
		OwnerID: ownerID,
		Name:    name,
	}).First(&task).Error
	return &task, err
}

// ListByProject retrieves all tasks for a specific project from the database with pagination
func (r *TaskRepository) ListByProject(ctx context.Context, ownerID uint, projectID uint, opts *models.ListOptions) ([]models.Task, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	var tasks []models.Task
	err := r.db.WithContext(ctx).Where(models.Task{
		OwnerID:   ownerID,
		ProjectID: projectID,
	}).Limit(opts.Limit).Offset(opts.Offset).Find(&tasks).Error
	return tasks, err
}

// UpdateStatus updates the status of a task in the database
func (r *TaskRepository) UpdateStatus(ctx context.Context, ownerID uint, id uint, status models.TaskStatus) error {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}
	return r.db.WithContext(ctx).Model(&models.Task{}).Where(models.Task{
		Model:   gorm.Model{ID: id},
		OwnerID: ownerID,
	}).Update(models.TaskStatusField, status).Error
}

// Update updates an existing task in the database.
func (r *TaskRepository) Update(ctx context.Context, ownerID uint, task *models.Task) error {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}
	return r.db.WithContext(ctx).Model(&models.Task{}).Where(models.Task{
		Model:   gorm.Model{ID: task.ID},
		OwnerID: ownerID,
	}).Updates(task).Error
}

// UpdateBatch updates a batch of tasks in the database, processing them in smaller chunks within transactions.
func (r *TaskRepository) UpdateBatch(ctx context.Context, tasks []*models.Task) error {
	batchSize := 10
	for i := 0; i < len(tasks); i += batchSize {
		end := i + batchSize
		if end > len(tasks) {
			end = len(tasks)
		}
		batch := tasks[i:end]

		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			for _, task := range batch {
				if err := models.ValidateOwnerID(task.OwnerID); err != nil {
					return fmt.Errorf("invalid owner_id: %w", err)
				}
				// Ensure task has an ID, otherwise update is meaningless
				if task.ID == 0 {
					return fmt.Errorf("task missing ID for update")
				}
				// Updates only non-zero fields of the task struct by default
				// Use Updates for partial updates, Save for full replacement
				if err := tx.Model(&models.Task{}).Where(models.Task{
					Model:   gorm.Model{ID: task.ID},
				}).Updates(task).Error; err != nil {
					return fmt.Errorf("failed to update task %d: %w", task.ID, err) // Rollback
				}
			}
			return nil // Commit
		})

		if err != nil {
			// Transaction failed, return the error
			return fmt.Errorf("batch update transaction failed: %w", err)
		}
	}

	return nil
}
