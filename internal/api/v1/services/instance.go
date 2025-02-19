package services

import (
	"context"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
)

type InstanceService struct {
	repo *repos.InstanceRepository
}

func NewInstanceService(repo *repos.InstanceRepository) *InstanceService {
	return &InstanceService{repo: repo}
}

func (s *InstanceService) ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	return s.repo.List(ctx, opts)
}

func (s *InstanceService) CreateInstance(ctx context.Context, instance *models.Instance) error {
	return s.repo.Create(ctx, instance)
}
