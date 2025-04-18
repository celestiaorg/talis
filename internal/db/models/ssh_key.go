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
	Name              string     `json:"name" gorm:"index:idx_sshkey_owner_provider_name,unique"`                 // User-defined name for the key
	PublicKey         string     `json:"public_key" gorm:"type:text;not null"`                                    // The actual public key content (e.g., "ssh-ed25519 AAA...")
	FingerprintSHA256 string     `json:"fingerprint_sha256" gorm:"index;not null"`                                // SHA256 fingerprint (e.g., "SHA256:xyz...")
	FingerprintMD5    string     `json:"fingerprint_md5" gorm:"index"`                                            // Optional: MD5 fingerprint (legacy)
	ProviderID        ProviderID `json:"provider_id" gorm:"index:idx_sshkey_owner_provider_name,unique;not null"` // e.g., "do", "aws"
	ProviderKeyID     string     `json:"provider_key_id" gorm:"index;not null"`                                   // The ID assigned by the cloud provider
	OwnerID           uint       `json:"owner_id" gorm:"index:idx_sshkey_owner_provider_name,unique;not null"`    // User ID (0 could represent Talis Server key)
}

// Validate checks if the SSHKey model is valid before saving.
func (k *SSHKey) Validate() error {
	if k.Name == "" {
		return fmt.Errorf("SSH key name cannot be empty")
	}
	if k.PublicKey == "" {
		return fmt.Errorf("SSH public key content cannot be empty")
	}
	if k.FingerprintSHA256 == "" {
		// We could potentially calculate this if missing, but for now require it
		return fmt.Errorf("SSH key SHA256 fingerprint cannot be empty")
	}
	if k.ProviderID == "" || !k.ProviderID.IsValid() {
		return fmt.Errorf("invalid or missing provider ID")
	}
	if k.ProviderKeyID == "" {
		return fmt.Errorf("provider key ID cannot be empty")
	}
	// OwnerID 0 is reserved for the Talis Server Key
	return nil
}

// BeforeSave GORM hook to run validation.
func (k *SSHKey) BeforeSave(tx *gorm.DB) error {
	return k.Validate()
}
