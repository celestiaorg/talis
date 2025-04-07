package compute

import (
	"context"
	"fmt"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
)

// Provider defines the interface for cloud providers
type Provider interface {
	// ValidateCredentials validates the provider credentials
	ValidateCredentials() error

	// GetEnvironmentVars returns the environment variables needed for the provider
	GetEnvironmentVars() map[string]string

	// ConfigureProvider configures the provider with the given stack
	ConfigureProvider(stack interface{}) error

	// CreateInstance creates a new instance
	CreateInstance(ctx context.Context, name string, config types.InstanceConfig) ([]types.InstanceInfo, error)

	// DeleteInstance deletes an instance
	DeleteInstance(ctx context.Context, name string, region string) error
}

// InstanceConfig represents the configuration for creating an instance
type InstanceConfig struct {
	Region            string               // Region where to create the instance
	Size              string               // Size/type of the instance
	Image             string               // OS image to use
	SSHKeyID          string               // SSH key name to use
	Tags              []string             // Tags to apply to the instance
	NumberOfInstances int                  // Number of instances to create
	CustomName        string               // Optional custom name for this specific instance
	Volumes           []types.VolumeConfig `json:"volumes,omitempty"` // Volumes to attach to the instance
}

// InstanceInfo represents information about a created instance
type InstanceInfo struct {
	ID            string                // Provider-specific instance ID
	Name          string                // Instance name
	PublicIP      string                // Public IP address
	Provider      models.ProviderID     // Provider name (e.g., "do")
	Region        string                // Region where instance was created
	Size          string                // Instance size/type
	Volumes       []string              `json:"volumes,omitempty"`        // List of attached volume IDs
	VolumeDetails []types.VolumeDetails `json:"volume_details,omitempty"` // Detailed information about attached volumes
}

// Provisioner is the interface for system configuration
type Provisioner interface {
	ConfigureHost(host string, sshKeyPath string) error
	ConfigureHosts(hosts []string, sshKeyPath string) error
	CreateInventory(instances map[string]string, keyPath string) error
	RunAnsiblePlaybook(inventoryName string) error
}

// NewComputeProvider creates a new compute provider based on the provider name
func NewComputeProvider(provider models.ProviderID) (Provider, error) {
	switch provider {
	case models.ProviderDO:
		return NewDigitalOceanProvider()
	case "do-mock", "digitalocean-mock":
		return NewMockDOClient(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// NewProvisioner creates a new system provisioner
func NewProvisioner(jobID string) types.Provisioner {
	return NewAnsibleConfigurator(jobID)
}
