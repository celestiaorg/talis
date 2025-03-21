package services

import (
	"context"
	"fmt"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
)

// JobService provides business logic for job operations
type JobService struct {
	repo *repos.JobRepository
}

// NewJobService creates a new job service instance
func NewJobService(repo *repos.JobRepository) *JobService {
	return &JobService{repo: repo}
}

// ListJobs retrieves a paginated list of jobs
func (s *JobService) ListJobs(ctx context.Context, status models.JobStatus, ownerID uint, opts *models.ListOptions) ([]models.Job, error) {
	return s.repo.List(ctx, status, ownerID, opts)
}

// CreateJob creates a new job
func (s *JobService) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	// TODO: this should be removed since job.Name is required. This code is never hit?
	if job.Name == "" {
		job.Name = fmt.Sprintf("job-%s", time.Now().Format("20060102-150405"))
	}

	return job, s.repo.Create(ctx, job)
}

// GetJobStatus retrieves the status of a job
func (s *JobService) GetJobStatus(ctx context.Context, ownerID uint, id uint) (models.JobStatus, error) {
	j, err := s.repo.GetByID(ctx, ownerID, id)
	if err != nil {
		return models.JobStatusUnknown, err
	}
	return j.Status, nil
}

// UpdateJobStatus updates the status of a job
func (s *JobService) UpdateJobStatus(ctx context.Context, id uint, status models.JobStatus, result interface{}, errMsg string) error {
	return s.repo.UpdateStatus(ctx, id, status, result, errMsg)
}

// GetByProjectName retrieves a job by its project name
func (s *JobService) GetByProjectName(ctx context.Context, projectName string) (*models.Job, error) {
	return s.repo.GetByProjectName(ctx, projectName)
}

// GetJob retrieves a job by its ID
func (s *JobService) GetJob(ctx context.Context, id uint) (*models.Job, error) {
	return s.repo.GetByID(ctx, 0, id) // Using 0 as ownerID since it's not being used yet
}
