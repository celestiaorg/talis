package services

import (
	"context"
	"fmt"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
	// "github.com/celestiaorg/talis/internal/sshutil" // Needed for EnsureTalisServerKey later
)

// SSHKeyService provides logic for managing SSH keys in the database.
type SSHKeyService struct {
	repo *repos.SSHKeyRepository
	// providerMap map[models.ProviderID]compute.Provider // Needed later for provider interaction
}

// NewSSHKeyService creates a new SSH key service.
func NewSSHKeyService(repo *repos.SSHKeyRepository /* Add providerMap later */) *SSHKeyService {
	return &SSHKeyService{
		repo: repo,
		// providerMap: providerMap,
	}
}

// EnsureTalisServerKey checks if the Talis server key exists in the DB for supported providers
// and potentially registers it with the provider API and DB if missing.
// This is a placeholder - full implementation requires provider interaction.
// func (s *SSHKeyService) EnsureTalisServerKey(ctx context.Context, privateKeyPEM string) error { // Removed
// 	logger.Info("Checking Talis Server SSH key registration (placeholder)...")
// 	if privateKeyPEM == "" {
// 		logger.Warn("TALIS_SERVER_SSH_PRIVATE_KEY is not set. Cannot ensure server key registration.")
// 		// Depending on policy, this might be a fatal error.
// 		return nil // Allow startup for now
// 	}
//
// 	// TODO: Implement actual logic:
// 	// 1. Derive public key & fingerprint from privateKeyPEM using sshutil.
// 	// 2. Iterate through supported providers in providerMap.
// 	// 3. For each provider:
// 	//    a. Check if a key with IsTalisServerKey=true exists in DB (s.repo.GetTalisServerKey).
// 	//    b. If not found in DB:
// 	//       i. Check if key exists on provider API using fingerprint.
// 	//       ii. If not on provider, register it via provider API.
// 	//       iii. Get ProviderKeyID (from registration or lookup).
// 	//       iv. Create DB entry using s.repo.Create with IsTalisServerKey=true, OwnerID=0, etc.
// 	//    c. If found in DB, optionally verify it still exists on provider API.
//
// 	logger.Info("Talis Server SSH key check complete (placeholder). Ensure keys are manually registered with providers and DB for now.")
// 	return nil
// }

// GetTalisServerProviderKeyID retrieves the cached ProviderKeyID for the Talis server key for a specific provider.
// func (s *SSHKeyService) GetTalisServerProviderKeyID(ctx context.Context, providerID models.ProviderID) (string, error) { // Removed
// 	key, err := s.repo.GetTalisServerKey(ctx, providerID)
// 	if err != nil {
// 		// Wrap error for clarity
// 		return "", fmt.Errorf("failed to get Talis server key ID from DB for provider %s: %w", providerID, err)
// 	}
// 	return key.ProviderKeyID, nil
// }

// GetProviderKeyIDsByNames looks up multiple user SSH keys by name for a specific owner and provider
// and returns a slice of their ProviderKeyIDs.
// It logs warnings for names not found but does not return an error for partial matches.
func (s *SSHKeyService) GetProviderKeyIDsByNames(ctx context.Context, ownerID uint, providerID models.ProviderID, names []string) ([]string, error) {
	if len(names) == 0 {
		return []string{}, nil
	}

	keys, err := s.repo.GetByNames(ctx, ownerID, providerID, names)
	if err != nil {
		return nil, fmt.Errorf("database error fetching user keys by name for provider %s: %w", providerID, err)
	}

	foundIDs := make([]string, 0, len(keys))
	foundNames := make(map[string]bool)
	for _, key := range keys {
		foundIDs = append(foundIDs, key.ProviderKeyID)
		foundNames[key.Name] = true
	}

	// Log warnings for names not found in the database cache
	for _, name := range names {
		if !foundNames[name] {
			logger.Warnf("SSH key '%s' for owner %d and provider %s not found in Talis database cache.", name, ownerID, providerID)
		}
	}

	return foundIDs, nil
}
