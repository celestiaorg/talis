package compute

import "github.com/celestiaorg/talis/internal/types"

// CalculateTotalDiskSize calculates the total disk size needed for a server
// by adding the base disk size from the package and any additional volumes
func CalculateTotalDiskSize(baseSize int, volumes []types.VolumeConfig) int {
	totalSize := baseSize

	for _, volume := range volumes {
		totalSize += volume.SizeGB
	}

	return totalSize
}

// ValidateVolumes checks if the volumes configuration is valid for VirtFusion
// Since VirtFusion doesn't support separate volumes, we validate that:
// 1. All volumes have valid mount points
// 2. Mount points don't conflict
func ValidateVolumes(volumes []types.VolumeConfig) error {
	mountPoints := make(map[string]bool)

	for _, volume := range volumes {
		if volume.MountPoint == "" {
			return types.NewValidationError("volume mount point cannot be empty")
		}

		if volume.SizeGB <= 0 {
			return types.NewValidationError("volume size must be positive")
		}

		if _, exists := mountPoints[volume.MountPoint]; exists {
			return types.NewValidationError("duplicate mount point: " + volume.MountPoint)
		}

		mountPoints[volume.MountPoint] = true
	}

	return nil
}
