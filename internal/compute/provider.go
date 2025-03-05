package compute

import (
	"context"
	"fmt"
)

// ComputeProvider defines the interface for cloud providers
type ComputeProvider interface {
	ConfigureProvider(stack interface{}) error
	CreateInstance(ctx context.Context, name string, config InstanceConfig) (InstanceInfo, error)
	DeleteInstance(ctx context.Context, name string) error
	ValidateCredentials() error
	GetEnvironmentVars() map[string]string
}

// InstanceConfig represents the configuration for a compute instance
type InstanceConfig struct {
	Region            string
	Size              string
	Image             string
	UserData          string
	SSHKeyID          string
	Tags              []string
	NumberOfInstances int
}

// InstanceInfo contains information about a created instance
type InstanceInfo struct {
	ID       string
	Name     string
	PublicIP string
	Provider string
	Region   string
	Size     string
}

// Provisioner is the interface for system configuration
type Provisioner interface {
	ConfigureHost(host string, sshKeyPath string) error
	ConfigureHosts(hosts []string, sshKeyPath string) error
	CreateInventory(instances map[string]string, keyPath string) error
	RunAnsiblePlaybook(inventoryName string) error
}

// NewComputeProvider creates a new compute provider based on the provider name
func NewComputeProvider(provider string) (ComputeProvider, error) {
	switch provider {
	case "digitalocean":
		return NewDigitalOceanProvider()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// NewProvisioner creates a new system provisioner
func NewProvisioner(jobID string) Provisioner {
	return NewAnsibleConfigurator(jobID)
}
