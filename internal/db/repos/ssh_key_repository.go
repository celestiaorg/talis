package repos

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/celestiaorg/talis/internal/db/models"
)

// SSHKeyRepository defines the interface for SSH key database operations.
// type SSHKeyRepository interface { // <-- Remove interface
// 	Create(ctx context.Context, key *models.SSHKey) error
// 	GetByNames(ctx context.Context, ownerID uint, providerID models.ProviderID, names []string) ([]*models.SSHKey, error)
// 	// Add List, Delete, GetByID etc. as needed later
// }

// SSHKeyRepository provides methods for interacting with SSHKey data in the database.
type SSHKeyRepository struct { // <-- Rename struct
	db *gorm.DB
}

// NewSSHKeyRepository creates a new SSH key repository.
func NewSSHKeyRepository(db *gorm.DB) *SSHKeyRepository { // <-- Update return type
	return &SSHKeyRepository{db: db} // <-- Update struct name
}

// Create adds a new SSH key record to the database.
func (r *SSHKeyRepository) Create(ctx context.Context, key *models.SSHKey) error { // <-- Update receiver
	if err := key.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	result := r.db.WithContext(ctx).Create(key)
	return result.Error
}

// GetTalisServerKey retrieves the specific SSH key designated as the Talis Server key for a given provider.
// func (r *sshKeyRepository) GetTalisServerKey(ctx context.Context, providerID models.ProviderID) (*models.SSHKey, error) {
// 	var key models.SSHKey
// 	result := r.db.WithContext(ctx).
// 		Where("is_talis_server_key = ? AND provider_id = ?", true, providerID).
// 		First(&key)
// 	if result.Error != nil {
// 		if result.Error == gorm.ErrRecordNotFound {
// 			return nil, fmt.Errorf("Talis server key not found for provider %s: %w", providerID, result.Error)
// 		}
// 		return nil, fmt.Errorf("database error fetching Talis server key: %w", result.Error)
// 	}
// 	return &key, nil
// }

// GetByNames retrieves multiple SSH keys by their names for a specific owner and provider.
// Returns only the keys that are found, does not return an error if some names are not found.
func (r *SSHKeyRepository) GetByNames(ctx context.Context, ownerID uint, providerID models.ProviderID, names []string) ([]*models.SSHKey, error) {
	if len(names) == 0 {
		return []*models.SSHKey{}, nil // Return empty slice if no names provided
	}

	nameInterfaces := make([]interface{}, len(names))
	for i, name := range names {
		nameInterfaces[i] = name
	}

	var keys []*models.SSHKey
	result := r.db.WithContext(ctx).
		Where(&models.SSHKey{
			OwnerID:    ownerID,
			ProviderID: providerID,
		}).
		Where(clause.IN{Column: models.SSHKeyNameColumn, Values: nameInterfaces}).
		Find(&keys)

	if result.Error != nil {
		return nil, fmt.Errorf("database error fetching keys by name: %w", result.Error)
	}
	return keys, nil
}

// GetByFingerprintSHA256 retrieves an SSH key by its SHA256 fingerprint for a specific owner and provider.
func (r *SSHKeyRepository) GetByFingerprintSHA256(ctx context.Context, ownerID uint, providerID models.ProviderID, fingerprintSHA256 string) (*models.SSHKey, error) {
	var key models.SSHKey
	result := r.db.WithContext(ctx).
		Where(&models.SSHKey{
			OwnerID:           ownerID,
			ProviderID:        providerID,
			FingerprintSHA256: fingerprintSHA256,
		}).
		First(&key)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Return gorm.ErrRecordNotFound directly so callers can check for it specifically.
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("database error fetching key by fingerprint: %w", result.Error)
	}
	return &key, nil
}
