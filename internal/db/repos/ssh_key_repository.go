package repos

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/celestiaorg/talis/internal/db/models"
)

// SSHKeyRepository provides methods for interacting with SSH keys in the database
type SSHKeyRepository struct {
	db *gorm.DB
}

// NewSSHKeyRepository creates a new SSHKeyRepository
func NewSSHKeyRepository(db *gorm.DB) *SSHKeyRepository {
	return &SSHKeyRepository{db: db}
}

// Create creates a new SSH key in the database
func (r *SSHKeyRepository) Create(ctx context.Context, key *models.SSHKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// GetByNames retrieves multiple SSH keys by their names for a specific owner.
// Returns only the keys that are found, does not return an error if some names are not found.
func (r *SSHKeyRepository) GetByNames(ctx context.Context, ownerID uint, names []string) ([]*models.SSHKey, error) {
	if len(names) == 0 {
		return []*models.SSHKey{}, nil // Return empty slice if no names provided
	}

	nameInterfaces := make([]interface{}, len(names))
	for i, name := range names {
		nameInterfaces[i] = name
	}

	var keys []*models.SSHKey
	err := r.db.WithContext(ctx).
		Where(&models.SSHKey{
			OwnerID: ownerID,
		}).
		Where(clause.IN{Column: models.SSHKeyNameColumn, Values: nameInterfaces}).
		Find(&keys).Error

	if err != nil {
		return nil, fmt.Errorf("database error fetching keys by name: %w", err)
	}
	return keys, nil
}

// Delete removes an SSH key by name and owner ID
func (r *SSHKeyRepository) Delete(ctx context.Context, ownerID uint, name string) error {
	result := r.db.WithContext(ctx).
		Where(&models.SSHKey{
			Name:    name,
			OwnerID: ownerID,
		}).
		Delete(&models.SSHKey{})

	if result.Error != nil {
		return fmt.Errorf("database error deleting key: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("SSH key '%s' not found for owner %d", name, ownerID)
	}

	return nil
}

// List retrieves all SSH keys for a specific owner
func (r *SSHKeyRepository) List(ctx context.Context, ownerID uint) ([]*models.SSHKey, error) {
	var keys []*models.SSHKey
	err := r.db.WithContext(ctx).
		Where(&models.SSHKey{
			OwnerID: ownerID,
		}).
		Find(&keys).Error

	if err != nil {
		return nil, fmt.Errorf("database error listing keys: %w", err)
	}
	return keys, nil
}
