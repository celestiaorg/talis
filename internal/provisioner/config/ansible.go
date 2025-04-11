// Package config provides configuration types for provisioners
package config

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
	// For now, we'll accept any configuration
	// In the future, we can add validation logic here
	return nil
}
