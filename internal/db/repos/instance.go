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

// Update updates an instance
func (r *InstanceRepository) Update(ctx context.Context, ID uint, instance *models.Instance) error {
	return r.db.WithContext(ctx).Where(&models.Instance{Model: gorm.Model{ID: ID}}).Updates(instance).Error
}

// UpdateIPByName updates the public IP of an instance by its name
func (r *InstanceRepository) UpdateIPByName(ctx context.Context, name string, ip string) error {
	return r.db.WithContext(ctx).Model(&models.Instance{}).
		Where(&models.Instance{Name: name}).
		Update(models.InstancePublicIPField, ip).Error
}

// UpdateStatus updates the status of an instance
func (r *InstanceRepository) UpdateStatus(ctx context.Context, ID uint, status models.InstanceStatus) error {
	return r.db.WithContext(ctx).Model(&models.Instance{}).
		Where(&models.Instance{Model: gorm.Model{ID: ID}}).
		Update("status", status).Error
}

// UpdateStatusByName updates the status of an instance by its name
func (r *InstanceRepository) UpdateStatusByName(ctx context.Context, name string, status models.InstanceStatus) error {
	return r.db.WithContext(ctx).Model(&models.Instance{}).
		Where(&models.Instance{Name: name}).
		Update("status", status).Error
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

// Query executes a custom query against the instance table
func (r *InstanceRepository) Query(ctx context.Context, query string, args ...interface{}) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.Raw(query, args...).Scan(&instances).Error
	return instances, err
}

// Get retrieves an instance by ID
func (r *InstanceRepository) Get(ctx context.Context, id uint) (*models.Instance, error) {
	var instance models.Instance
	if err := r.db.First(&instance, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
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
	// First update the status to terminated
	if err := r.db.WithContext(ctx).Model(&models.Instance{}).
		Where(&models.Instance{Model: gorm.Model{ID: id}}).
		Update(models.InstanceStatusField, models.InstanceStatusTerminated).Error; err != nil {
		return err
	}

	// Then perform the soft delete
	return r.db.WithContext(ctx).Delete(&models.Instance{}, id).Error
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
