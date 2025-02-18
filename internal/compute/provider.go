package compute

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ComputeProvider defines the interface for cloud providers
type ComputeProvider interface {
	ConfigureProvider(stack auto.Stack) error
	CreateInstance(ctx *pulumi.Context, name string, config InstanceConfig) (InstanceInfo, error)
	ValidateCredentials() error
	GetEnvironmentVars() map[string]string
}

// InstanceConfig represents the configuration for a compute instance
type InstanceConfig struct {
	Region   string
	Size     string
	Image    string
	UserData string
	SSHKeyID string
	Tags     []string
}

// InstanceInfo represents information about a created instance
type InstanceInfo struct {
	PublicIP pulumi.StringOutput
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
		return &DigitalOceanProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// NewProvisioner creates a new system provisioner
func NewProvisioner(jobID string) Provisioner {
	return NewAnsibleConfigurator(jobID)
}
