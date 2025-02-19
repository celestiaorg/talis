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

func (s *JobService) ListJobs(ctx context.Context, opts *models.ListOptions) ([]models.Job, error) {
	return s.repo.List(ctx, opts)
}

func (s *JobService) CreateJob(ctx context.Context, ID, webhookURL string) (*models.Job, error) {
	if ID == "" {
		ID = fmt.Sprintf("job-%s", time.Now().Format("20060102-150405"))
	}

	j := &models.Job{
		ID:         ID,
		Status:     models.JobStatusPending,
		CreatedAt:  time.Now(),
		WebhookURL: webhookURL,
	}
	return j, s.repo.Create(ctx, j)
}

func (s *JobService) GetJobStatus(ctx context.Context, id string) (models.JobStatus, error) {
	j, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	return j.Status, nil
}

func (s *JobService) UpdateJobStatus(ctx context.Context, ID string, status models.JobStatus, result interface{}, errMsg string) error {
	return s.repo.UpdateStatus(ctx, ID, status, result, errMsg)
}
