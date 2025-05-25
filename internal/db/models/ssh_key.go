package models

import (
	"fmt"

	"gorm.io/gorm"
)

// SSHKey field name constants for database queries
const (
	SSHKeyNameColumn = "name"
)

// SSHKey represents an SSH public key registered with a provider and managed by Talis.
type SSHKey struct {
	gorm.Model
	Name      string `json:"name" gorm:"index:idx_sshkey_owner_provider_name,unique"`              // User-defined name for the key
	PublicKey string `json:"public_key" gorm:"type:text;not null"`                                 // The actual public key content (e.g., "ssh-ed25519 AAA...")
	OwnerID   uint   `json:"owner_id" gorm:"index:idx_sshkey_owner_provider_name,unique;not null"` // User ID (0 could represent Talis Server key)
}

// Validate checks if the SSHKey model is valid before saving.
func (k *SSHKey) Validate() error {
	if k.Name == "" {
		return fmt.Errorf("SSH key name cannot be empty")
	}
	if k.PublicKey == "" {
		return fmt.Errorf("SSH public key content cannot be empty")
	}
	// OwnerID 0 is reserved for the Talis Server Key
	return nil
}

// BeforeSave GORM hook to run validation.
func (k *SSHKey) BeforeSave(_ *gorm.DB) error {
	return k.Validate()
}

// BeforeCreate GORM hook to run validation.
func (k *SSHKey) BeforeCreate(_ *gorm.DB) error {
	return k.Validate()
}
