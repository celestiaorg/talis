// Package repos provides repository implementations for database operations
package repos

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

// Create creates a new instance in the database
func (r *InstanceRepository) Create(ctx context.Context, instance *models.Instance) (*models.Instance, error) {
	if instance == nil {
		return nil, fmt.Errorf("instance cannot be nil")
	}
	if err := models.ValidateOwnerID(instance.OwnerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	if err := r.db.WithContext(ctx).Create(instance).Error; err != nil {
		return nil, err
	}
	return instance, nil
}

// Update updates an instance by its ID. Only non-zero fields in the instance parameter will be updated.
// GORM's Updates method will:
// - Only update fields with non-zero values in the instance parameter
// - Ignore zero-value fields (null, 0, "", false)
// - Not update fields marked with gorm:"-" tag
// - Not update CreatedAt field
// - Automatically update UpdatedAt if it exists
func (r *InstanceRepository) Update(ctx context.Context, ownerID, id uint, instance *models.Instance) error {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Model(&models.Instance{}).Where(&models.Instance{Model: gorm.Model{ID: id}})
		if ownerID != models.AdminID {
			query = query.Where(&models.Instance{OwnerID: ownerID})
		}

		result := query.Updates(instance)
		if err := result.Error; err != nil {
			return fmt.Errorf("failed to update instance by ID: %w", err)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("instance not found or not owned by user")
		}
		return nil
	})
}

// List retrieves a paginated list of instances
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
func (r *InstanceRepository) List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Instance, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	var instances []models.Instance
	query := r.applyListOptions(r.db.WithContext(ctx), opts)
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	err := query.Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}
	return instances, nil
}

// Count returns the total number of instances
func (r *InstanceRepository) Count(ctx context.Context, ownerID uint) (int64, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return 0, fmt.Errorf("invalid owner_id: %w", err)
	}

	query := r.db.WithContext(ctx).
		Unscoped().
		Model(&models.Instance{})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

// Query executes a custom query against the instance table
// This method is admin-only for security reasons
func (r *InstanceRepository) Query(ctx context.Context, ownerID uint, query string, args ...interface{}) ([]models.Instance, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	var instances []models.Instance
	dbQuery := r.db.WithContext(ctx).Unscoped()
	if ownerID != models.AdminID {
		dbQuery = dbQuery.Where(&models.Instance{OwnerID: ownerID})
	}
	err := dbQuery.Where(query, args...).Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	return instances, nil
}

// Get retrieves an instance by ID
func (r *InstanceRepository) Get(ctx context.Context, ownerID, id uint) (*models.Instance, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	var instance models.Instance
	query := r.db.WithContext(ctx).
		Unscoped().
		Where(&models.Instance{Model: gorm.Model{ID: id}})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	if err := query.First(&instance).Error; err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	return &instance, nil
}

// GetByProjectIDAndInstanceIDs retrieves instances that belong to a specific project and match the given instance IDs.
func (r *InstanceRepository) GetByProjectIDAndInstanceIDs(
	ctx context.Context,
	ownerID uint,
	projectID uint,
	instanceIDs []uint,
) ([]models.Instance, error) {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	if len(instanceIDs) == 0 {
		return []models.Instance{}, nil
	}

	// Convert []uint to []interface{} for clause.IN
	idsAsInterfaces := make([]interface{}, len(instanceIDs))
	for i, id := range instanceIDs {
		idsAsInterfaces[i] = id
	}

	var instances []models.Instance
	query := r.db.WithContext(ctx).
		Where(&models.Instance{ProjectID: projectID}).
		Where(clause.IN{Column: models.InstanceIDField, Values: idsAsInterfaces})

	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	err := query.Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances by project and IDs: %w", err)
	}
	return instances, nil
}

// Terminate updates the status of an instance to terminated and performs a soft delete
func (r *InstanceRepository) Terminate(ctx context.Context, ownerID, id uint) error {
	if err := models.ValidateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}

	// To prevent race conditions, we use a transaction to wrap the update and delete
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First update the status to terminated
		query := tx.Model(&models.Instance{}).Where(&models.Instance{Model: gorm.Model{ID: id}})
		if ownerID != models.AdminID {
			query = query.Where(&models.Instance{OwnerID: ownerID})
		}
		result := query.Update(models.InstanceStatusField, models.InstanceStatusTerminated)
		if err := result.Error; err != nil {
			return fmt.Errorf("failed to update instance status: %w", err)
		}

		return nil
	})
}

// CreateBatch creates a batch of instances
func (r *InstanceRepository) CreateBatch(ctx context.Context, instances []*models.Instance) ([]*models.Instance, error) {
	if instances == nil {
		return nil, fmt.Errorf("instances cannot be nil")
	}

	for i, instance := range instances {
		if instance == nil {
			return nil, fmt.Errorf("instance at index %d cannot be nil", i)
		}
		if err := models.ValidateOwnerID(instance.OwnerID); err != nil {
			return nil, fmt.Errorf("invalid owner_id for instance at index %d: %w", i, err)
		}
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(instances, models.DBBatchSize).Error
	}); err != nil {
		return nil, err
	}
	return instances, nil
}
