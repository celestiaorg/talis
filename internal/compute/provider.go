package compute

import (
	"context"
	"fmt"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/test/mocks"
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

// Provisioner is the interface for system configuration
type Provisioner interface {
	// ConfigureHost configures a single host
	ConfigureHost(host string, sshKeyPath string) error

	// ConfigureHosts configures multiple hosts in parallel, ensuring SSH readiness
	ConfigureHosts(hosts []string, sshKeyPath string) error

	// CreateInventory creates an Ansible inventory file from instance info
	CreateInventory(instances []types.InstanceInfo, sshKeyPath string) error

	// RunAnsiblePlaybook runs the Ansible playbook
	RunAnsiblePlaybook(inventoryName string) error
}

// NewComputeProvider creates a new compute provider based on the provider name
func NewComputeProvider(provider models.ProviderID) (Provider, error) {
	switch provider {
	case models.ProviderDO:
		return NewDigitalOceanProvider()
	case "do-mock", "digitalocean-mock":
		return mocks.NewMockDOClient(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// IsValidProvider checks whether the given provider ID is supported.
func IsValidProvider(provider models.ProviderID) bool {
	if _, ok := validProviders[provider]; !ok {
		return false
	}
	return true
}

var validProviders = map[models.ProviderID]struct{}{
	"do-mock": {}, "digitalocean-mock": {}, "do": {},
}

// NewProvisioner creates a new system provisioner
func NewProvisioner(jobID string) Provisioner {
	return NewAnsibleConfigurator(jobID)
}
