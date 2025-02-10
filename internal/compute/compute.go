package compute

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// InstanceInfo stores information about the created instance (e.g. IP)
type InstanceInfo struct {
	PublicIP pulumi.StringOutput
}

// ComputeProvider is the interface for creating instances across different providers
type ComputeProvider interface {
	CreateInstance(ctx *pulumi.Context, name string, config InstanceConfig) (InstanceInfo, error)
	ConfigureProvider(stack auto.Stack) error
}

type InstanceConfig struct {
	Region   string
	Size     string
	Image    string
	UserData string
	SSHKeyID string
	Tags     []string
}

// NewComputeProvider returns the correct implementation based on providerName
func NewComputeProvider(providerName string) (ComputeProvider, error) {
	switch providerName {
	case "digitalocean":
		return &DigitalOceanProvider{}, nil
	case "aws":
		// return &AWSProvider{}, nil  // Para futura implementación
		return nil, fmt.Errorf("aws provider not implemented yet")
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}
