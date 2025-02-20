package repos

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// Store handles job-related database operations
type InstanceRepository struct {
	db *gorm.DB
}

// NewStore creates a new job store
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
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	return &instance, nil
}

// Update updates an instance
func (r *InstanceRepository) Update(ctx context.Context, ID uint, instance *models.Instance) error {
	return r.db.WithContext(ctx).Where(&models.Instance{Model: gorm.Model{ID: ID}}).Updates(instance).Error
}

// UpdateStatus updates the status of an instance
func (r *InstanceRepository) UpdateStatus(ctx context.Context, ID uint, status models.InstanceStatus) error {
	return r.db.WithContext(ctx).
		Where(&models.Instance{Model: gorm.Model{ID: ID}}).
		Update("status", status).Error
}

func (r *InstanceRepository) List(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Limit(opts.Limit).Offset(opts.Offset).
		Order(models.InstanceCreatedAtField + " DESC").
		Find(&instances).Error
	return instances, err
}

func (r *InstanceRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Count(&count).Error
	return count, err
}

func (r *InstanceRepository) Query(ctx context.Context, query string, args ...interface{}) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.Raw(query, args...).Scan(&instances).Error
	return instances, err
}
