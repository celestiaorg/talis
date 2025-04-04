package repos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// JobRepository provides access to job-related database operations
type JobRepository struct {
	db *gorm.DB
}

// NewJobRepository creates a new job repository instance
func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Create creates a new job in the database
func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	if err := models.ValidateOwnerID(job.OwnerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}
	return r.db.WithContext(ctx).Create(job).Error
}

// Update updates an existing job in the database
func (r *JobRepository) Update(ctx context.Context, job *models.Job) error {
	if err := models.ValidateOwnerID(job.OwnerID); err != nil {
		return fmt.Errorf("invalid owner_id: %w", err)
	}
	return r.db.WithContext(ctx).Save(job).Error
}

// UpdateStatus updates the status of a job in the database
func (r *JobRepository) UpdateStatus(ctx context.Context, ID uint, status models.JobStatus, result interface{}, errMsg string) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	return r.db.WithContext(ctx).Model(&models.Job{}).
		Where(&models.Job{Model: gorm.Model{ID: ID}}).
		Updates(map[string]interface{}{
			"status": status,
			"result": resultJSON,
			"error":  errMsg,
		}).Error
}

// GetByID retrieves a job by its ID
// if the ownerID is 0, it will return the job regardless of the owner
func (r *JobRepository) GetByID(ctx context.Context, OwnerID, ID uint) (*models.Job, error) {
	if err := models.ValidateOwnerID(OwnerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	var job models.Job
	qry := &models.Job{Model: gorm.Model{ID: ID}}
	if OwnerID != models.AdminID {
		qry.OwnerID = OwnerID
	}
	err := r.db.WithContext(ctx).Where(qry).First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return &job, nil
}

// GetByName retrieves a job by its name
// if the ownerID is 0, it will return the job regardless of the owner
func (r *JobRepository) GetByName(ctx context.Context, OwnerID uint, name string) (*models.Job, error) {
	if err := models.ValidateOwnerID(OwnerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	var job models.Job
	qry := &models.Job{Name: name}
	if OwnerID != models.AdminID {
		qry.OwnerID = OwnerID
	}
	err := r.db.WithContext(ctx).Where(qry).First(&job).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return &job, nil
}

// List returns a list of jobs
// if the ownerID is 0, it will return the jobs regardless of the owner
// if the status is unknown, it will return all jobs regardless of their status
func (r *JobRepository) List(ctx context.Context, status models.JobStatus, OwnerID uint, opts *models.ListOptions) ([]models.Job, error) {
	if err := models.ValidateOwnerID(OwnerID); err != nil {
		return nil, fmt.Errorf("invalid owner_id: %w", err)
	}
	var jobs []models.Job
	qry := &models.Job{}

	// If status is unknown, we don't need to filter by status
	if status != models.JobStatusUnknown {
		qry.Status = status
	}

	// Zero is an option when admin is fetching a job
	if OwnerID != models.AdminID {
		qry.OwnerID = OwnerID
	}

	db := r.db.WithContext(ctx)
	if !opts.IncludeDeleted {
		db = db.Unscoped().Where("deleted_at IS NULL")
	}

	err := db.Model(&models.Job{}).
		Where(qry).
		Limit(opts.Limit).Offset(opts.Offset).
		Order(models.JobCreatedAtField + " DESC").
		Find(&jobs).Error
	return jobs, err
}

// Count returns the number of jobs
// if the ownerID is 0, it will return the number of jobs regardless of the owner
// if the status is unknown, it will return the number of jobs regardless of their status
func (r *JobRepository) Count(ctx context.Context, status models.JobStatus, OwnerID uint) (int64, error) {
	if err := models.ValidateOwnerID(OwnerID); err != nil {
		return 0, fmt.Errorf("invalid owner_id: %w", err)
	}
	var count int64
	qry := &models.Job{}

	// If status is unknown, we don't need to filter by status
	if status != models.JobStatusUnknown {
		qry.Status = status
	}

	// Zero is an option when admin is fetching a job
	if OwnerID != models.AdminID {
		qry.OwnerID = OwnerID
	}
	err := r.db.WithContext(ctx).Model(&models.Job{}).Where(qry).Count(&count).Error
	return count, err
}

// Query executes a raw SQL query against the jobs table
func (r *JobRepository) Query(_ context.Context, query string, args ...interface{}) ([]models.Job, error) {
	var jobs []models.Job
	result := r.db.Raw(query, args...).Scan(&jobs)
	return jobs, result.Error
}

// GetByProjectName retrieves a job by its project name
func (r *JobRepository) GetByProjectName(ctx context.Context, projectName string) (*models.Job, error) {
	var job models.Job
	result := r.db.WithContext(ctx).Where(&models.Job{ProjectName: projectName}).First(&job)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Silently return nil when no record is found
		}
		return nil, fmt.Errorf("failed to get job by project name: %w", result.Error)
	}
	return &job, nil
}
