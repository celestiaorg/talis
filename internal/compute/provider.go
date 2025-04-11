package compute

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/provisioner"
	"github.com/celestiaorg/talis/internal/provisioner/config"
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
	// Configure configures the system for the given instances
	Configure(ctx context.Context, instances []types.InstanceInfo) error

	// CreateInventory creates an inventory file for the provisioner
	CreateInventory(instances map[string]string, keyPath string) error

	// RunPlaybook runs the provisioning playbook
	RunPlaybook(inventoryPath string) error
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

// NewProvisioner creates a new system provisioner
func NewProvisioner(jobID string) Provisioner {
	// Extract job name from job ID (format: "job-20250411-134559")
	jobName := strings.TrimPrefix(jobID, "job-")

	cfg := &config.AnsibleConfig{
		JobID:             jobID,
		JobName:           jobName,
		SSHUser:           provisioner.DefaultSSHUser,
		SSHKeyPath:        os.ExpandEnv(provisioner.DefaultSSHKeyPath),
		PlaybookPath:      provisioner.PathToPlaybook,
		InventoryBasePath: provisioner.InventoryBasePath,
	}

	p, err := provisioner.NewAnsibleProvisioner(cfg)
	if err != nil {
		// This should never happen with our default config
		panic(fmt.Sprintf("failed to create provisioner: %v", err))
	}
	return p
}

// NewAnsibleProvisioner creates a new Ansible provisioner
func NewAnsibleProvisioner(jobID string) (Provisioner, error) {
	// Extract job name from job ID (format: "job-20250411-134559")
	jobName := strings.TrimPrefix(jobID, "job-")

	cfg := &config.AnsibleConfig{
		JobID:             jobID,
		JobName:           jobName,
		SSHUser:           provisioner.DefaultSSHUser,
		SSHKeyPath:        os.ExpandEnv(provisioner.DefaultSSHKeyPath),
		PlaybookPath:      provisioner.PathToPlaybook,
		InventoryBasePath: provisioner.InventoryBasePath,
	}

	return provisioner.NewAnsibleProvisioner(cfg)
}
