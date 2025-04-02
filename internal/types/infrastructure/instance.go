package infrastructure

import (
	"context"

	"github.com/celestiaorg/talis/internal/types"
)

// InstanceConfig represents the configuration for creating an instance
type InstanceConfig struct {
	Region            string               `json:"region"`                // Region where to create the instance
	Size              string               `json:"size"`                  // Size/type of the instance
	Image             string               `json:"image"`                 // OS image to use
	SSHKeyID          string               `json:"ssh_key_id"`            // SSH key name to use
	Tags              []string             `json:"tags,omitempty"`        // Tags to apply to the instance
	NumberOfInstances int                  `json:"number_of_instances"`   // Number of instances to create
	CustomName        string               `json:"custom_name,omitempty"` // Optional custom name for this specific instance
	Volumes           []types.VolumeConfig `json:"volumes,omitempty"`     // Volumes to attach to the instance
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
