package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// Store handles job-related database operations
type JobRepository struct {
	db *gorm.DB
}

// NewStore creates a new job store
func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Create creates a new job in the database
func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *JobRepository) UpdateStatus(ctx context.Context, id string, status models.JobStatus, result interface{}, errMsg string) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}

	return r.db.WithContext(ctx).Model(&models.Job{}).
		Where(&models.Job{ID: id}).
		Updates(map[string]interface{}{
			"status":     status,
			"result":     resultJSON,
			"error":      errMsg,
			"updated_at": time.Now(),
		}).Error
}

// GetByID retrieves a job by its ID
func (r *JobRepository) GetByID(ctx context.Context, id string) (*models.Job, error) {
	var job models.Job
	err := r.db.WithContext(ctx).Where(&models.Job{ID: id}).First(&job).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %v", err)
	}
	return &job, nil
}

func (r *JobRepository) List(ctx context.Context, opts *models.ListOptions) ([]models.Job, error) {
	var jobs []models.Job
	query := r.db.WithContext(ctx).Model(&models.Job{})
	if opts.Status != "" {
		query = query.Where(&models.Job{Status: opts.Status})
	}
	err := query.
		Limit(opts.Limit).Offset(opts.Offset).
		Order(models.JobCreatedAtField + " DESC").
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Job{}).Count(&count).Error
	return count, err
}

func (r *JobRepository) Query(ctx context.Context, query string, args ...interface{}) ([]models.Job, error) {
	var jobs []models.Job
	err := r.db.Raw(query, args...).Scan(&jobs).Error
	return jobs, err
}
