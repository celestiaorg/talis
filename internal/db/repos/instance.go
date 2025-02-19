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

func (r *InstanceRepository) GetByID(ctx context.Context, ID string) (*models.Instance, error) {
	var instance models.Instance
	err := r.db.WithContext(ctx).Where(&models.Instance{ID: ID}).First(&instance).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	return &instance, nil
}

func (r *InstanceRepository) Update(ctx context.Context, ID string, instance *models.Instance) error {
	return r.db.WithContext(ctx).Where(&models.Instance{ID: ID}).Updates(instance).Error
}

func (r *InstanceRepository) Delete(ctx context.Context, ID string) error {
	return r.db.WithContext(ctx).
		Where(&models.Instance{ID: ID}).
		Update(models.InstanceDeletedField, true).Error
}

func (r *InstanceRepository) List(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.WithContext(ctx).
		Where(&models.Instance{Deleted: false}).
		Model(&models.Instance{}).
		Limit(opts.Limit).Offset(opts.Offset).
		Order(models.InstanceCreatedAtField + " DESC").
		Find(&instances).Error
	return instances, err
}

func (r *InstanceRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Where(&models.Instance{Deleted: false}).
		Model(&models.Instance{}).
		Count(&count).Error
	return count, err
}

func (r *InstanceRepository) Query(ctx context.Context, query string, args ...interface{}) ([]models.Instance, error) {
	var instances []models.Instance
	err := r.db.Raw(query, args...).Scan(&instances).Error
	return instances, err
}
