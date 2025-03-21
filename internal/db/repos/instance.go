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
	err := r.db.WithContext(ctx).Where("name IN (?)", names).Find(&instances).Error
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
	return r.db.WithContext(ctx).Where(&models.Instance{Name: name}).Update(models.InstancePublicIPField, ip).Error
}

// UpdateStatus updates the status of an instance
func (r *InstanceRepository) UpdateStatus(ctx context.Context, ID uint, status models.InstanceStatus) error {
	return r.db.WithContext(ctx).
		Where(&models.Instance{Model: gorm.Model{ID: ID}}).
		Update("status", status).Error
}

// UpdateStatusByName updates the status of an instance by its name
func (r *InstanceRepository) UpdateStatusByName(ctx context.Context, name string, status models.InstanceStatus) error {
	return r.db.WithContext(ctx).
		Where(&models.Instance{Name: name}).
		Update("status", status).Error
}

// List retrieves a paginated list of instances
func (r *InstanceRepository) List(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Limit(opts.Limit).Offset(opts.Offset).
		Order(models.InstanceCreatedAtField + " DESC").
		Find(&instances).Error
	return instances, err
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
	err := r.db.WithContext(ctx).Where(&models.Instance{JobID: jobID}).Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}
	return instances, nil
}

// GetByJobIDOrdered retrieves all instances for a given job ID, ordered by creation date (oldest first)
func (r *InstanceRepository) GetByJobIDOrdered(ctx context.Context, jobID uint) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).
		Where(&models.Instance{JobID: jobID}).
		Order(models.InstanceCreatedAtField + " ASC"). // ASC order to get oldest first
		Find(&instances).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}
	return instances, nil
}

// Terminate updates the status of an instance to terminated
func (r *InstanceRepository) Terminate(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&models.Instance{}).
		Where(&models.Instance{Model: gorm.Model{ID: id}}).
		Update(models.InstanceStatusField, models.InstanceStatusTerminated).Error
}
