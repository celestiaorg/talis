package repos

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

// AcquireTaskLock attempts to lock a task for processing.
// Returns true if the lock was acquired, false otherwise.
func (r *TaskRepository) AcquireTaskLock(ctx context.Context, taskID uint) (bool, error) {
	now := time.Now()
	lockExpiry := now.Add(models.TaskLockTimeout)

	// Create a task model with ID for the where clause
	taskModel := &models.Task{Model: gorm.Model{ID: taskID}}

	// Attempt to acquire the lock using an atomic update
	result := r.db.WithContext(ctx).Model(taskModel).
		Where(
			clause.Or(
				clause.Eq{Column: models.TaskLockedAtField, Value: nil},
				clause.Lt{Column: models.TaskLockExpiryField, Value: now},
			),
		).
		Updates(map[string]interface{}{
			models.TaskLockedAtField:   now,
			models.TaskLockExpiryField: lockExpiry,
			models.TaskStatusField:     models.TaskStatusRunning,
		})

	if result.Error != nil {
		return false, fmt.Errorf("failed to acquire task lock: %w", result.Error)
	}

	// If rows affected is 0, the lock was not acquired
	return result.RowsAffected > 0, nil
}

// ReleaseTaskLock releases a task lock
func (r *TaskRepository) ReleaseTaskLock(ctx context.Context, taskID uint) error {
	// Create a task model with ID for the where clause
	taskModel := &models.Task{Model: gorm.Model{ID: taskID}}

	// Update the task to release the lock
	result := r.db.WithContext(ctx).Model(taskModel).
		Updates(map[string]interface{}{
			models.TaskLockedAtField:   nil,
			models.TaskLockExpiryField: nil,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to release task lock: %w", result.Error)
	}

	return nil
}

// RecoverStaleTasks finds tasks that were in progress when the system crashed
// and resets them to pending status with incremented attempts
func (r *TaskRepository) RecoverStaleTasks(ctx context.Context) (int64, error) {
	now := time.Now()

	// Find tasks that are in running state with expired locks or no locks
	result := r.db.WithContext(ctx).Model(&models.Task{}).
		Where(&models.Task{Status: models.TaskStatusRunning}).
		Where(
			clause.Or(
				clause.Eq{Column: models.TaskLockedAtField, Value: nil},
				clause.Lt{Column: models.TaskLockExpiryField, Value: now},
			),
		).
		Updates(map[string]interface{}{
			models.TaskStatusField:     models.TaskStatusPending,
			models.TaskLockedAtField:   nil,
			models.TaskLockExpiryField: nil,
			models.TaskAttemptsField:   gorm.Expr(fmt.Sprintf("%s + 1", models.TaskAttemptsField)),
			models.TaskLogsField:       gorm.Expr(fmt.Sprintf("CONCAT(%s, '\n[RECOVERED] Task was in running state during system restart')", models.TaskLogsField)),
		})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to recover stale tasks: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// GetSchedulableTasks retrieves tasks that are ready for processing,
// ordered by priority (high first), error status (no error first), and then by creation date (oldest first).
// It fetches tasks with statuses other than Completed or Terminated.
func (r *TaskRepository) GetSchedulableTasks(ctx context.Context, priority models.TaskPriority, limit int) ([]models.Task, error) {
	var tasks []models.Task

	// Define statuses to include
	includeStatuses := []interface{}{
		models.TaskStatusPending,
		models.TaskStatusRunning,
	}

	// Build the query
	query := r.db.WithContext(ctx).Model(&models.Task{}).
		Where(models.Task{
			Priority: priority,
		}).
		Where(
			clause.IN{
				Column: models.TaskStatusField,
				Values: includeStatuses,
			},
			clause.Lt{Column: models.TaskAttemptsField, Value: maxAttempts},
			clause.Or(
				clause.Eq{Column: models.TaskLockedAtField, Value: nil},
				clause.Lt{Column: models.TaskLockExpiryField, Value: time.Now()},
			),
		).
		// Order by priority (lower number = higher priority), error presence (errors last), then by creation date (oldest first)
		Order(clause.OrderByColumn{Column: clause.Column{Name: models.TaskPriorityField}, Desc: false}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: models.TaskIDField}, Desc: false}) // faster than created_at

	// Apply limit
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(models.DefaultLimit)
	}

	// Execute the query
	err := query.Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query schedulable tasks: %w", err)
	}

	return tasks, nil
}

// IncrementAttempts atomically increments the attempts count for a task
func (r *TaskRepository) IncrementAttempts(ctx context.Context, taskID uint) error {
	err := r.db.WithContext(ctx).
		Model(&models.Task{Model: gorm.Model{ID: taskID}}).
		Update(
			models.TaskAttemptsField,
			gorm.Expr(fmt.Sprintf("%s + 1", models.TaskAttemptsField)),
		).Error

	if err != nil {
		return fmt.Errorf("failed to increment task attempts: %w", err)
	}

	return nil
}

// ListByInstanceID retrieves all tasks for a specific instance from the database with pagination and optional action filter.
func (r *TaskRepository) ListByInstanceID(ctx context.Context, ownerID uint, instanceID uint, actionFilter models.TaskAction, opts *models.ListOptions) ([]models.Task, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	if instanceID == 0 {
		return nil, fmt.Errorf("instanceID cannot be zero")
	}

	var tasks []models.Task
	query := r.db.WithContext(ctx).Where(models.Task{
		OwnerID:    ownerID,
		InstanceID: instanceID,
	})

	if actionFilter != "" {
		query = query.Where(&models.Task{Action: actionFilter})
	}

	if opts != nil {
		if opts.Limit > 0 {
			query = query.Limit(opts.Limit)
		}
		if opts.Offset > 0 {
			query = query.Offset(opts.Offset)
		}
	}

	err := query.Order(models.TaskCreatedAtField + " DESC").Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	return tasks, nil
}
