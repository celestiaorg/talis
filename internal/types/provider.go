package types

import "context"

// ComputeProvider is the interface for compute providers
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

// Provisioner is the interface for system configuration
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
	ComputeProvider
	Droplets() DropletService
	Keys() KeyService
	Storage() StorageService
}
