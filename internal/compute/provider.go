package compute

import (
	"context"
	"fmt"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
)

// ComputeProvider is the interface for compute providers
type ComputeProvider interface {
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

// Provisioner is the interface for system configuration
type Provisioner interface {
	ConfigureHost(host string, sshKeyPath string) error
	ConfigureHosts(hosts []string, sshKeyPath string) error
	CreateInventory(instances map[string]string, keyPath string) error
	RunAnsiblePlaybook(inventoryName string) error
}

// NewComputeProvider creates a new compute provider based on the provider name
func NewComputeProvider(provider models.ProviderID) (types.ComputeProvider, error) {
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
