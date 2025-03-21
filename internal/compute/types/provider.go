package types

// InstanceConfig represents the configuration for creating an instance
type InstanceConfig struct {
	Region            string   // Region where to create the instance
	Size              string   // Size/type of the instance
	Image             string   // OS image to use
	SSHKeyID          string   // SSH key name to use
	Tags              []string // Tags to apply to the instance
	NumberOfInstances int      // Number of instances to create
	CustomName        string   // Optional custom name for this specific instance
}

// InstanceInfo represents information about a created instance
type InstanceInfo struct {
	ID       string // Provider-specific instance ID
	Name     string // Instance name
	PublicIP string // Public IP address
	Provider string // Provider name (e.g., "digitalocean")
	Region   string // Region where instance was created
	Size     string // Instance size/type
}
