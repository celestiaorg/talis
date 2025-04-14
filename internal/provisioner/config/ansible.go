// Package config provides configuration types for provisioners
package config

import "fmt"

const (
	// DefaultInventoryBasePath is the default base path for Ansible inventory files
	DefaultInventoryBasePath = "ansible/inventory"
)

// AnsibleConfig represents the configuration for Ansible provisioner
type AnsibleConfig struct {
	// JobID is the unique identifier for the current job
	JobID string
	// JobName is the name of the current job
	JobName string
	// OwnerID is the ID of the owner of the instances
	OwnerID uint
	// SSHKeyPath is the path to the SSH private key
	SSHKeyPath string
	// SSHUser is the user for SSH connections
	SSHUser string
	// PlaybookPath is the path to the Ansible playbook
	PlaybookPath string
	// InventoryBasePath is the base path for inventory files
	InventoryBasePath string
}

// Validate validates the Ansible configuration
func (c *AnsibleConfig) Validate() error {
	// Validate required fields
	if c.JobID == "" {
		return fmt.Errorf("JobID is required")
	}
	// TODO: Add validation for OwnerID
	// if c.OwnerID == 0 {
	// 	return fmt.Errorf("OwnerID is required")
	// }

	// Set default values for optional fields
	if c.InventoryBasePath == "" {
		c.InventoryBasePath = DefaultInventoryBasePath
	}

	return nil
}
