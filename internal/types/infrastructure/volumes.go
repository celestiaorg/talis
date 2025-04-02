package infrastructure

import (
	"fmt"
)

// VolumeConfig represents the configuration for a volume
type VolumeConfig struct {
	Name       string `json:"name"`        // Name of the volume
	SizeGB     int    `json:"size_gb"`     // Size in gigabytes
	Region     string `json:"region"`      // Region where to create the volume
	FileSystem string `json:"filesystem"`  // File system type (optional)
	MountPoint string `json:"mount_point"` // Where to mount the volume
}

// Validate checks if the volume configuration is valid
func (v *VolumeConfig) Validate(instanceRegion string) error {
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
