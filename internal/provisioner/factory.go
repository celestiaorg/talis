// Package provisioner provides system configuration and provisioning functionality
package provisioner

import (
	"fmt"

	"github.com/celestiaorg/talis/internal/provisioner/config"
)

// Type represents the type of provisioner
type Type string

const (
	// ProvisionerAnsible represents the Ansible provisioner
	ProvisionerAnsible Type = "ansible"
)

// NewProvisioner creates a new provisioner of the specified type
func NewProvisioner(typ Type, cfg Config) (Provisioner, error) {
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
