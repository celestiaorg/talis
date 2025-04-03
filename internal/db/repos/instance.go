// Package repos provides repository implementations for database operations
package repos

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// InstanceRepository provides access to instance-related database operations
type InstanceRepository struct {
	db *gorm.DB
}

// NewInstanceRepository creates a new instance repository instance
func NewInstanceRepository(db *gorm.DB) *InstanceRepository {
	return &InstanceRepository{db: db}
}

// Create creates a new job in the database
func (r *InstanceRepository) Create(ctx context.Context, instance *models.Instance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

// GetByID retrieves an instance by its ID
// if the jobID is 0, it will return the instance regardless of the job (Designed for admin)
func (r *InstanceRepository) GetByID(ctx context.Context, JobID, ID uint) (*models.Instance, error) {
	var instance models.Instance
	qry := &models.Instance{Model: gorm.Model{ID: ID}}
	if JobID != 0 {
		qry.JobID = JobID
	}
	err := r.db.WithContext(ctx).Where(qry).First(&instance).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return &instance, nil
}

// GetByNames retrieves instances by their names
// TODO: Need to add ownerID to the query for security in future
func (r *InstanceRepository) GetByNames(ctx context.Context, names []string) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).Unscoped().Where("name IN (?)", names).Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	return instances, nil
}

// UpdateByID updates an instance by its ID. Only non-zero fields in the instance parameter will be updated.
// GORM's Updates method will:
// - Only update fields with non-zero values in the instance parameter
// - Ignore zero-value fields (null, 0, "", false)
// - Not update fields marked with gorm:"-" tag
// - Not update CreatedAt field
// - Automatically update UpdatedAt if it exists
func (r *InstanceRepository) UpdateByID(ctx context.Context, id uint, instance *models.Instance) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Instance{}).
			Where(&models.Instance{Model: gorm.Model{ID: id}}).
			Updates(instance).Error; err != nil {
			return fmt.Errorf("failed to update instance by ID: %w", err)
		}
		return nil
	})
}

// UpdateByName updates an instance by its name. Only non-zero fields in the instance parameter will be updated.
// See UpdateByID method for details on how GORM handles field updates.
func (r *InstanceRepository) UpdateByName(ctx context.Context, name string, instance *models.Instance) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Instance{}).
			Where(&models.Instance{Name: name}).
			Updates(instance).Error; err != nil {
			return fmt.Errorf("failed to update instance by name: %w", err)
		}
		return nil
	})
}

// applyListOptions applies the list options to the given query
func (r *InstanceRepository) applyListOptions(query *gorm.DB, opts *models.ListOptions) *gorm.DB {
	if opts == nil {
		return query.Where("status != ?", models.InstanceStatusTerminated)
	}

	// Apply status filter if provided
	if opts.InstanceStatus != nil {
		if opts.StatusFilter == models.StatusFilterNotEqual {
			query = query.Where("status != ?", *opts.InstanceStatus)
		} else {
			query = query.Where("status = ?", *opts.InstanceStatus)
		}
	} else if !opts.IncludeDeleted {
		// By default, only show non-terminated instances if not including deleted
		query = query.Where("status != ?", models.InstanceStatusTerminated)
	}

	// Apply soft delete filter
	if opts.IncludeDeleted {
		query = query.Unscoped()
	}

	// Apply pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	return query
}

// List returns a list of instances based on the provided options
func (r *InstanceRepository) List(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	var instances []models.Instance
	query := r.applyListOptions(r.db.WithContext(ctx), opts)

	err := query.Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}
	return instances, nil
}

// Count returns the total number of instances
func (r *InstanceRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Count(&count).Error
	return count, err
}

// Query executes a raw SQL query against the jobs table
func (r *InstanceRepository) Query(_ context.Context, query string, args ...interface{}) ([]models.Instance, error) {
	var instances []models.Instance
	result := r.db.Raw(query, args...).Scan(&instances)
	return instances, result.Error
}

// Get retrieves a job by ID
func (r *InstanceRepository) Get(_ context.Context, id uint) (*models.Instance, error) {
	var instance models.Instance
	if err := r.db.First(&instance, id).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

// GetByJobID retrieves all instances for a given job ID
func (r *InstanceRepository) GetByJobID(ctx context.Context, jobID uint) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).Unscoped().Where(&models.Instance{JobID: jobID}).Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}
	return instances, nil
}

// GetByJobIDOrdered retrieves all instances for a given job ID, ordered by creation date (oldest first)
func (r *InstanceRepository) GetByJobIDOrdered(ctx context.Context, jobID uint) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).
		Unscoped().
		Where(&models.Instance{JobID: jobID}).
		Order(models.InstanceCreatedAtField + " ASC"). // ASC order to get oldest first
		Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}
	return instances, nil
}

// Terminate updates the status of an instance to terminated and performs a soft delete
func (r *InstanceRepository) Terminate(ctx context.Context, id uint) error {
	// To prevent race conditions, we use a transaction to wrap the update and delete
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First update the status to terminated
		if err := tx.Model(&models.Instance{}).
			Where(&models.Instance{Model: gorm.Model{ID: id}}).
			Update(models.InstanceStatusField, models.InstanceStatusTerminated).Error; err != nil {
			return fmt.Errorf("failed to update instance status: %w", err)
		}

		// Then perform the soft delete
		if err := tx.Delete(&models.Instance{}, id).Error; err != nil {
			return fmt.Errorf("failed to soft delete instance: %w", err)
		}

		return nil
	})
}

// GetByJobIDAndNames retrieves instances that belong to a specific job and match the given names
func (r *InstanceRepository) GetByJobIDAndNames(
	ctx context.Context,
	jobID uint,
	names []string,
) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).
		Where("job_id = ? AND name IN (?) AND deleted_at IS NULL", jobID, names).
		Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	return instances, nil
}
