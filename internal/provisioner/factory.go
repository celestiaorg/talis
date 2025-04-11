// Package provisioner provides system configuration and provisioning functionality
package provisioner

import (
	"fmt"

	"github.com/celestiaorg/talis/internal/provisioner/config"
)

// ProvisionerType represents the type of provisioner
type ProvisionerType string

const (
	// ProvisionerAnsible represents the Ansible provisioner
	ProvisionerAnsible ProvisionerType = "ansible"
)

// NewProvisioner creates a new provisioner of the specified type
func NewProvisioner(typ ProvisionerType, cfg Config) (Provisioner, error) {
	switch typ {
	case ProvisionerAnsible:
		ansibleConfig, ok := cfg.(*config.AnsibleConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Ansible provisioner: expected *config.AnsibleConfig")
		}
		return NewAnsibleProvisioner(ansibleConfig)
	default:
		return nil, fmt.Errorf("unknown provisioner type: %s", typ)
	}
}
