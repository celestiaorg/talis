package infrastructure

import (
	"fmt"

	"github.com/celestiaorg/talis/internal/types"
)

// ValidateVolume checks if the volume configuration is valid
func ValidateVolume(v *types.VolumeConfig, instanceRegion string) error {
	if v.SizeGB < 1 {
		return fmt.Errorf("volume size must be at least 1 GB, got %d", v.SizeGB)
	}
	// Allow empty region (will be set to instance region) or validate it matches
	if v.Region != "" && v.Region != instanceRegion {
		return fmt.Errorf("volume region %s must match instance region %s", v.Region, instanceRegion)
	}
	return nil
}

// Provisioner defines the interface for system configuration
type Provisioner interface {
	// ConfigureHost configures a single host
	ConfigureHost(host string, sshKeyPath string) error

	// ConfigureHosts configures multiple hosts in parallel
	ConfigureHosts(hosts []string, sshKeyPath string) error

	// CreateInventory creates an Ansible inventory file
	CreateInventory(instances map[string]string, keyPath string) error

	// RunAnsiblePlaybook runs the Ansible playbook
	RunAnsiblePlaybook(inventoryName string) error
}
