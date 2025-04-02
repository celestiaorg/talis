package infrastructure

import (
	"fmt"

	"github.com/celestiaorg/talis/internal/types"
)

// ValidateVolume checks if the volume configuration is valid
func ValidateVolume(v *types.VolumeConfig, instanceRegion string) error {
	if v.SizeGB < 15 {
		return fmt.Errorf("volume size must be at least 15 GB, got %d", v.SizeGB)
	}
	if v.SizeGB > 30720 { // 30 TB in GB
		return fmt.Errorf("volume size must be at most 30 TB, got %d GB", v.SizeGB)
	}
	if v.Region != instanceRegion {
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
