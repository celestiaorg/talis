package repos

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// validateOwnerID ensures the ownerID is valid
func validateOwnerID(ownerID uint) error {
	if ownerID == 0 {
		return fmt.Errorf("owner_id cannot be 0")
	}
	return nil
}

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
	if err := validateOwnerID(instance.OwnerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}
	return r.db.WithContext(ctx).Create(instance).Error
}

// GetByID retrieves an instance by its ID
// if the ownerID is models.AdminID, it will return the instance regardless of ownership
func (r *InstanceRepository) GetByID(ctx context.Context, ownerID, JobID, ID uint) (*models.Instance, error) {
	if err := validateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	var instance models.Instance
	qry := &models.Instance{
		Model: gorm.Model{ID: ID},
		JobID: JobID,
	}
	if ownerID != models.AdminID {
		qry.OwnerID = ownerID
	}

	err := r.db.WithContext(ctx).
		Unscoped().
		Where(qry).
		First(&instance).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	return &instance, nil
}

// GetByNames retrieves instances by their names
func (r *InstanceRepository) GetByNames(ctx context.Context, ownerID uint, names []string) ([]models.Instance, error) {
	if err := validateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	var instances []models.Instance
	query := r.db.WithContext(ctx).
		Unscoped().
		Where("name IN (?)", names)
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	err := query.Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	return instances, nil
}

// Update updates an instance
func (r *InstanceRepository) Update(ctx context.Context, ownerID, ID uint, instance *models.Instance) error {
	if err := validateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}

	query := r.db.WithContext(ctx).
		Unscoped().
		Where(&models.Instance{Model: gorm.Model{ID: ID}})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	result := query.Updates(instance)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found or not owned by user")
	}
	return nil
}

// UpdateIPByName updates the public IP of an instance by its name
func (r *InstanceRepository) UpdateIPByName(ctx context.Context, ownerID uint, name string, ip string) error {
	if err := validateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}

	query := r.db.WithContext(ctx).
		Unscoped().
		Model(&models.Instance{}).
		Where(&models.Instance{Name: name})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	result := query.Update(models.InstancePublicIPField, ip)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found or not owned by user")
	}
	return nil
}

// UpdateStatus updates the status of an instance
func (r *InstanceRepository) UpdateStatus(ctx context.Context, ownerID, ID uint, status models.InstanceStatus) error {
	if err := validateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}

	query := r.db.WithContext(ctx).
		Unscoped().
		Model(&models.Instance{}).
		Where(&models.Instance{Model: gorm.Model{ID: ID}})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	result := query.Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found or not owned by user")
	}
	return nil
}

// UpdateStatusByName updates the status of an instance by its name
func (r *InstanceRepository) UpdateStatusByName(ctx context.Context, ownerID uint, name string, status models.InstanceStatus) error {
	if err := validateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}

	query := r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Where(&models.Instance{Name: name})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	result := query.Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found or not owned by user")
	}
	return nil
}

// List retrieves a paginated list of instances
func (r *InstanceRepository) List(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Instance, error) {
	if err := validateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	var instances []models.Instance
	query := r.db.WithContext(ctx)
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	if opts != nil && opts.IncludeDeleted {
		query = query.Unscoped()
	}

	err := query.Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}
	return instances, nil
}

// Count returns the total number of instances
func (r *InstanceRepository) Count(ctx context.Context, ownerID uint) (int64, error) {
	if err := validateOwnerID(ownerID); err != nil {
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
	if err := validateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	if ownerID != models.AdminID {
		return nil, fmt.Errorf("query method is restricted to admin users only")
	}

	var instances []models.Instance
	err := r.db.Raw(query, args...).Scan(&instances).Error
	return instances, err
}

// Get retrieves an instance by ID
func (r *InstanceRepository) Get(ctx context.Context, ownerID, id uint) (*models.Instance, error) {
	if err := validateOwnerID(ownerID); err != nil {
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

// GetByJobID retrieves all instances for a given job ID
func (r *InstanceRepository) GetByJobID(ctx context.Context, ownerID, jobID uint) ([]models.Instance, error) {
	if err := validateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	var instances []models.Instance
	query := r.db.WithContext(ctx).
		Unscoped().
		Where(&models.Instance{JobID: jobID})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	err := query.Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}
	return instances, nil
}

// GetByJobIDOrdered retrieves all instances for a given job ID, ordered by creation date (oldest first)
func (r *InstanceRepository) GetByJobIDOrdered(ctx context.Context, ownerID, jobID uint) ([]models.Instance, error) {
	if err := validateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	query := r.db.WithContext(ctx).
		Unscoped().
		Where(&models.Instance{JobID: jobID})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}
	query = query.Order(models.InstanceCreatedAtField + " ASC") // ASC order to get oldest first

	var instances []models.Instance
	err := query.Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}
	return instances, nil
}

// Terminate updates the status of an instance to terminated and performs a soft delete
func (r *InstanceRepository) Terminate(ctx context.Context, ownerID, id uint) error {
	if err := validateOwnerID(ownerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}

	query := r.db.WithContext(ctx).
		Unscoped().
		Model(&models.Instance{}).
		Where(&models.Instance{Model: gorm.Model{ID: id}})
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	// First update the status to terminated
	result := query.Update(models.InstanceStatusField, models.InstanceStatusTerminated)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found or not owned by user")
	}

	// Then perform the soft delete
	return r.db.WithContext(ctx).Delete(&models.Instance{}, id).Error
}

// GetByJobIDAndNames retrieves instances that belong to a specific job and match the given names
func (r *InstanceRepository) GetByJobIDAndNames(
	ctx context.Context,
	ownerID uint,
	jobID uint,
	names []string,
) ([]models.Instance, error) {
	if err := validateOwnerID(ownerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}

	query := r.db.WithContext(ctx).
		Unscoped().
		Where("job_id = ? AND name IN (?) AND deleted_at IS NULL", jobID, names)
	if ownerID != models.AdminID {
		query = query.Where(&models.Instance{OwnerID: ownerID})
	}

	var instances []models.Instance
	err := query.Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	return instances, nil
}
