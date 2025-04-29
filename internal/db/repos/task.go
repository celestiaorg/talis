package repos

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

const maxAttempts = 10

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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(tasks, models.DBBatchSize).Error
	})
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
	query := r.db.WithContext(ctx).Where(models.Task{
		OwnerID:   ownerID,
		ProjectID: projectID,
	})
	if opts != nil {
		query = query.Limit(opts.Limit).Offset(opts.Offset)
	}
	err := query.Find(&tasks).Error
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

// GetSchedulableTasks retrieves tasks that are ready for processing,
// ordered by error status (no error first) and then by creation date (oldest first).
// It fetches tasks with statuses other than Completed or Terminated.
// The excludedActions parameter allows filtering out specific task action types.
func (r *TaskRepository) GetSchedulableTasks(ctx context.Context, limit int, excludedActions ...models.TaskAction) ([]models.Task, error) {
	var tasks []models.Task

	// Define statuses to exclude
	excludedStatuses := []models.TaskStatus{
		models.TaskStatusCompleted,
		models.TaskStatusTerminated,
	}

	// Build the query
	query := r.db.WithContext(ctx).Model(&models.Task{}).Where(
		"status NOT IN ?", excludedStatuses,
	).Where("attempts < ?", maxAttempts)

	// Exclude specific task action types if provided
	if len(excludedActions) > 0 {
		query = query.Where("action NOT IN ?", excludedActions)
	}

	// Order by error presence (errors last), then by creation date (oldest first)
	// Use DB-specific syntax for CASE WHEN or similar logic if needed, assuming standard SQL here.
	// GORM automatically quotes column names.
	query = query.Order("CASE WHEN error = '' THEN 0 ELSE 1 END").Order("created_at ASC")

	// Apply limit
	defaultLimit := 100
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(defaultLimit)
	}

	// Execute the query
	err := query.Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query schedulable tasks: %w", err)
	}

	return tasks, nil
}
