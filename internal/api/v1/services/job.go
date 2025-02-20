package services

import (
	"context"
	"fmt"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
)

type JobService struct {
	repo *repos.JobRepository
}

func NewJobService(repo *repos.JobRepository) *JobService {
	return &JobService{repo: repo}
}

func (s *JobService) ListJobs(ctx context.Context, status models.JobStatus, ownerID uint, opts *models.ListOptions) ([]models.Job, error) {
	return s.repo.List(ctx, status, ownerID, opts)
}

func (s *JobService) CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	if job.Name == "" {
		job.Name = fmt.Sprintf("job-%s", time.Now().Format("20060102-150405"))
	}

	return job, s.repo.Create(ctx, job)
}

func (s *JobService) GetJobStatus(ctx context.Context, ownerID uint, id uint) (models.JobStatus, error) {
	j, err := s.repo.GetByID(ctx, ownerID, id)
	if err != nil {
		return models.JobStatusUnknown, err
	}
	return j.Status, nil
}

func (s *JobService) UpdateJobStatus(ctx context.Context, id uint, status models.JobStatus, result interface{}, errMsg string) error {
	return s.repo.UpdateStatus(ctx, id, status, result, errMsg)
}
