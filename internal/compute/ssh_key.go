package compute

import (
	"context"
	"fmt"

	"github.com/celestiaorg/talis/internal/db/repos"
)

// SSHKeyManager handles SSH key operations for VirtFusion servers
type SSHKeyManager struct {
	userRepo *repos.UserRepository
}

// NewSSHKeyManager creates a new SSHKeyManager instance
func NewSSHKeyManager(userRepo *repos.UserRepository) *SSHKeyManager {
	return &SSHKeyManager{
		userRepo: userRepo,
	}
}

// GetSSHKey retrieves the SSH key content for a given key name and user ID
func (m *SSHKeyManager) GetSSHKey(ctx context.Context, userID uint, keyName string) (string, error) {
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user.PublicSSHKey == "" {
		return "", fmt.Errorf("user has no SSH key configured")
	}

	// In our current implementation, we only support one SSH key per user
	// The keyName parameter is reserved for future use when we support multiple keys
	return user.PublicSSHKey, nil
}

// GetSSHKeys retrieves all SSH keys for a given user ID
func (m *SSHKeyManager) GetSSHKeys(ctx context.Context, userID uint) ([]string, error) {
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.PublicSSHKey == "" {
		return []string{}, nil
	}

	// Currently, we only support one SSH key per user
	return []string{user.PublicSSHKey}, nil
}

// ValidateSSHKey checks if the provided SSH key name exists for the user
func (m *SSHKeyManager) ValidateSSHKey(ctx context.Context, userID uint, keyName string) error {
	user, err := m.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.PublicSSHKey == "" {
		return fmt.Errorf("user has no SSH key configured")
	}

	// Currently, we only support one SSH key per user, so any non-empty keyName is valid
	// as long as the user has a key configured
	return nil
}
