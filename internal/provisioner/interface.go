// Package provisioner provides system configuration and provisioning functionality
package provisioner

import (
	"context"

	"github.com/celestiaorg/talis/internal/types"
)

// Provisioner defines the interface for system configuration
type Provisioner interface {
	// Configure configures the system for the given instances
	Configure(ctx context.Context, instances []types.InstanceInfo) error

	// CreateInventory creates an inventory file for the provisioner
	CreateInventory(instances map[string]string, keyPath string) error

	// RunPlaybook runs the provisioning playbook
	RunPlaybook(inventoryPath string) error
}

// Config defines the interface for provisioner configuration
type Config interface {
	// Validate validates the configuration
	Validate() error
}
