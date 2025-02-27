package services

import (
	"context"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
)

// InstanceService provides business logic for instance operations
type InstanceService struct {
	repo *repos.InstanceRepository
}

// NewInstanceService creates a new instance service instance
func NewInstanceService(repo *repos.InstanceRepository) *InstanceService {
	return &InstanceService{repo: repo}
}

// ListInstances retrieves a paginated list of instances
func (s *InstanceService) ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	return s.repo.List(ctx, opts)
}

// CreateInstance creates a new instance
func (s *InstanceService) CreateInstance(ctx context.Context, instance *models.Instance) error {
	return s.repo.Create(ctx, instance)
}
