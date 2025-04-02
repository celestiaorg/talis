package types

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
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
	Provider string   `json:"provider"`          // Provider name (e.g., "digitalocean")
	Region   string   `json:"region"`            // Region where instance was created
	Size     string   `json:"size"`              // Instance size/type
	Volumes  []string `json:"volumes,omitempty"` // List of attached volume IDs
}

// ComputeProvider defines the interface for cloud providers
type ComputeProvider interface {
	// ValidateCredentials validates the provider credentials
	ValidateCredentials() error

	// GetEnvironmentVars returns the environment variables needed for the provider
	GetEnvironmentVars() map[string]string

	// ConfigureProvider configures the provider with the given stack
	ConfigureProvider(stack interface{}) error

	// CreateInstance creates a new instance
	CreateInstance(ctx context.Context, name string, config InstanceConfig) ([]InstanceInfo, error)

	// DeleteInstance deletes an instance
	DeleteInstance(ctx context.Context, name string, region string) error
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

// DOClient is the interface for DigitalOcean client operations
type DOClient interface {
	Droplets() DropletService
	Keys() KeyService
	Storage() StorageService
}

// DropletService is the interface for droplet operations
type DropletService interface {
	Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	Delete(ctx context.Context, id int) (*godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

// KeyService is the interface for SSH key operations
type KeyService interface {
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// StorageService is the interface for volume operations
type StorageService interface {
	CreateVolume(ctx context.Context, request *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error)
	DeleteVolume(ctx context.Context, id string) (*godo.Response, error)
	ListVolumes(ctx context.Context, opt *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error)
	GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error)
	GetVolumeAction(ctx context.Context, volumeID string, actionID int) (*godo.Action, *godo.Response, error)
	AttachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error)
	DetachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error)
}
