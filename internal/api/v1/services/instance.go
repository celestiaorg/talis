package services

import (
	"context"
	"fmt"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// InstanceService provides business logic for instance operations
type InstanceService struct {
	repo       *repos.InstanceRepository
	jobService *JobService
}

// NewInstanceService creates a new instance service instance
func NewInstanceService(repo *repos.InstanceRepository, jobService *JobService) *InstanceService {
	return &InstanceService{
		repo:       repo,
		jobService: jobService,
	}
}

// ListInstances retrieves a paginated list of instances
func (s *InstanceService) ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	return s.repo.List(ctx, opts)
}

// CreateInstance creates a new instance
func (s *InstanceService) CreateInstance(
	ctx context.Context,
	name,
	projectName,
	webhookURL string,
	instances []infrastructure.InstanceRequest,
) error {
	// Start async infrastructure creation
	return fmt.Errorf("not implemented")
}

// GetInstance retrieves an instance by ID
func (s *InstanceService) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, id)
}
