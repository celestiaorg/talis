package types

// VolumeConfig represents the configuration for a volume
type VolumeConfig struct {
	Name       string `json:"name"`        // Name of the volume
	SizeGB     int    `json:"size_gb"`     // Size in gigabytes
	Region     string `json:"region"`      // Region where to create the volume
	FileSystem string `json:"filesystem"`  // File system type (optional)
	MountPoint string `json:"mount_point"` // Where to mount the volume
}

// InstanceConfig represents the configuration for creating an instance
type InstanceConfig struct {
	Region            string         `json:"region"`                // Region where to create the instance
	Size              string         `json:"size"`                  // Size/type of the instance
	Image             string         `json:"image"`                 // OS image to use
	SSHKeyID          string         `json:"ssh_key_id"`            // SSH key name to use
	Tags              []string       `json:"tags,omitempty"`        // Tags to apply to the instance
	NumberOfInstances int            `json:"number_of_instances"`   // Number of instances to create
	CustomName        string         `json:"custom_name,omitempty"` // Optional custom name for this specific instance
	Volumes           []VolumeConfig `json:"volumes,omitempty"`     // Volumes to attach to the instance
}

// InstanceInfo represents information about a created instance
type InstanceInfo struct {
	ID       string   `json:"id"`                // Provider-specific instance ID
	Name     string   `json:"name"`              // Instance name
	PublicIP string   `json:"public_ip"`         // Public IP address
	Provider string   `json:"provider"`          // Provider name (e.g., "do", "aws", etc)
	Region   string   `json:"region"`            // Region where instance was created
	Size     string   `json:"size"`              // Instance size/type
	Volumes  []string `json:"volumes,omitempty"` // List of attached volume IDs
}
