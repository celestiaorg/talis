package types

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

// VolumeDetails represents detailed information about a created volume
type VolumeDetails struct {
	ID         string `json:"id"`          // Volume ID
	Name       string `json:"name"`        // Volume name
	Region     string `json:"region"`      // Region where volume was created
	SizeGB     int    `json:"size_gb"`     // Size in gigabytes
	MountPoint string `json:"mount_point"` // Where the volume is mounted
}

// ValidateVolume checks if the volume configuration is valid
func ValidateVolume(v *VolumeConfig, instanceRegion string) error {
	// Allow empty region (will be set to instance region) or validate it matches
	if v.Region != "" && v.Region != instanceRegion {
		return fmt.Errorf("volume region %s does not match instance region %s", v.Region, instanceRegion)
	}
	return nil
}
