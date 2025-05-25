package services

import (
	"context"
	"fmt"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
)

// SSHKeyService provides logic for managing SSH keys in the database.
type SSHKeyService struct {
	repo *repos.SSHKeyRepository
}

// NewSSHKeyService creates a new SSH key service.
func NewSSHKeyService(repo *repos.SSHKeyRepository) *SSHKeyService {
	return &SSHKeyService{
		repo: repo,
	}
}

// GetPublicKeysByNames looks up multiple user SSH keys by name for a specific owner
// and returns a slice of their public key contents.
// It logs warnings for names not found but does not return an error for partial matches.
func (s *SSHKeyService) GetPublicKeysByNames(ctx context.Context, ownerID uint, names []string) ([]string, error) {
	if len(names) == 0 {
		return []string{}, nil
	}

	keys, err := s.repo.GetByNames(ctx, ownerID, names)
	if err != nil {
		return nil, fmt.Errorf("database error fetching user keys by name: %w", err)
	}

	publicKeys := make([]string, 0, len(keys))
	foundNames := make(map[string]bool)
	for _, key := range keys {
		publicKeys = append(publicKeys, key.PublicKey)
		foundNames[key.Name] = true
	}

	// Log warnings for names not found in the database cache
	for _, name := range names {
		if !foundNames[name] {
			logger.Warnf("SSH key '%s' for owner %d not found in Talis database.", name, ownerID)
		}
	}

	return publicKeys, nil
}

// CreateSSHKey creates a new SSH key in the database
func (s *SSHKeyService) CreateSSHKey(ctx context.Context, key *models.SSHKey) error {
	return s.repo.Create(ctx, key)
}

// ListKeys retrieves all SSH keys for a specific owner
func (s *SSHKeyService) ListKeys(ctx context.Context, ownerID uint) ([]*models.SSHKey, error) {
	return s.repo.List(ctx, ownerID)
}

// DeleteKey deletes an SSH key by name for a specific owner
func (s *SSHKeyService) DeleteKey(ctx context.Context, ownerID uint, name string) error {
	return s.repo.Delete(ctx, ownerID, name)
}
