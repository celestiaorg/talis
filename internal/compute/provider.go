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
	CreateInstance(ctx context.Context, req *types.InstanceRequest) error

	// DeleteInstance deletes an instance
	DeleteInstance(ctx context.Context, providerInstanceID int) error
}

// Provisioner is the interface for system configuration
type Provisioner interface {
	// ConfigureHost configures a single host
	ConfigureHost(ctx context.Context, host string, sshKeyPath string) error

	// ConfigureHosts configures multiple hosts in parallel, ensuring SSH readiness
	ConfigureHosts(ctx context.Context, hosts []string, sshKeyPath string) error

	// CreateInventory creates an Ansible inventory file from instance info
	CreateInventory(instance *types.InstanceRequest, sshKeyPath string) (string, error)

	// RunAnsiblePlaybook runs the Ansible playbook
	RunAnsiblePlaybook(inventoryName string, tags []string) error
}

// NewComputeProvider creates a new compute provider based on the provider name
func NewComputeProvider(provider models.ProviderID) (Provider, error) {
	switch provider {
	case models.ProviderDO:
		return NewDigitalOceanProvider()
	case models.ProviderXimera:
		return NewXimeraProvider()
	case "do-mock", "digitalocean-mock":
		return mocks.NewMockDOClient(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// NewProvisioner creates a new system provisioner
func NewProvisioner(jobID string) Provisioner {
	return NewAnsibleConfigurator(jobID)
}
