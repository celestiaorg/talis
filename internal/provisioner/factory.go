// Package provisioner provides system configuration and provisioning functionality
package provisioner

import (
	"fmt"
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
		ansibleConfig, ok := cfg.(*AnsibleConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Ansible provisioner: expected *AnsibleConfig")
		}
		return NewAnsibleProvisioner(ansibleConfig)
	default:
		return nil, fmt.Errorf("unknown provisioner type: %s", typ)
	}
}
