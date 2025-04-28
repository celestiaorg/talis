package compute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/celestiaorg/talis/internal/types"
)

// Provisioner handles the provisioning of instances using Ansible
type Provisioner struct {
	jobID string
}

// NewProvisioner creates a new provisioner instance
func NewProvisioner(jobID string) *Provisioner {
	return &Provisioner{
		jobID: jobID,
	}
}

// Provision provisions an instance using Ansible
func (p *Provisioner) Provision(instance types.InstanceInfo, sshKeyPath string) error {
	// Skip if no payload to execute
	if !instance.ExecutePayload {
		return nil
	}

	// Validate payload path
	if instance.PayloadPath == "" {
		return fmt.Errorf("payload path is required for provisioning")
	}

	// Check if payload file exists
	if _, err := os.Stat(instance.PayloadPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("payload file does not exist: %s", instance.PayloadPath)
		}
		return fmt.Errorf("error accessing payload file: %w", err)
	}

	// Create working directory for this job
	workDir := filepath.Join(os.TempDir(), "talis", p.jobID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}

	// TODO: Implement actual Ansible provisioning
	// For now, just log that we would provision
	fmt.Printf("Would provision instance %s with payload %s\n", instance.Name, instance.PayloadPath)

	return nil
}
