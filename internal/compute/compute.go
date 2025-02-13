// Package compute provides a ComputeProvider interface and implementations for different cloud providers
package compute

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ComputeProvider defines the interface for cloud providers
type ComputeProvider interface {
	ConfigureProvider(stack auto.Stack) error
	CreateInstance(ctx *pulumi.Context, name string, config InstanceConfig) (pulumi.Resource, error)
	GetNixOSConfig() string
	ValidateCredentials() error
	GetEnvironmentVars() map[string]string
}

// InstanceInfo represents information about a created instance
type InstanceInfo struct {
	PublicIP pulumi.StringOutput // Public IP address of the instance
}

// InstanceConfig represents the configuration for a created instance
type InstanceConfig struct {
	Region   string   // Region of the instance
	Size     string   // Size of the instance
	Image    string   // Image to use for the instance
	SSHKeyID string   // SSH key ID to use for the instance
	Tags     []string // Tags to apply to the instance
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
