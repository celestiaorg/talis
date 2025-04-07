package infrastructure

import (
	"context"

	"github.com/celestiaorg/talis/internal/compute"
)

// ComputeProvider defines the interface for cloud providers
type ComputeProvider interface {
	// ValidateCredentials validates the provider credentials
	ValidateCredentials() error

	// GetEnvironmentVars returns the environment variables needed for the provider
	GetEnvironmentVars() map[string]string

	// ConfigureProvider configures the provider with the given stack
	ConfigureProvider(stack interface{}) error

	// CreateInstance creates a new instance
	CreateInstance(ctx context.Context, name string, config compute.InstanceConfig) ([]compute.InstanceInfo, error)

	// DeleteInstance deletes an instance
	DeleteInstance(ctx context.Context, name string, region string) error
}
