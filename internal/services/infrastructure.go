// Package services provides business logic implementation for the API
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/google/uuid"
)

// Infrastructure handles the provisioning and management of compute instances
type Infrastructure struct {
	provider    types.Provider
	provisioner *compute.Provisioner
}

// NewInfrastructure creates a new infrastructure service
func NewInfrastructure(req *types.InstancesRequest) (*Infrastructure, error) {
	provider, err := compute.Provider()
	if err != nil {
		return nil, fmt.Errorf("failed to create compute provider: %w", err)
	}

	if err := provider.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("failed to validate provider credentials: %w", err)
	}

	jobID := uuid.New().String()
	provisioner := compute.NewProvisioner(jobID)

	return &Infrastructure{
		provider:    provider,
		provisioner: provisioner,
	}, nil
}

// Execute creates and provisions instances based on the provided requests
func (i *Infrastructure) Execute(ctx context.Context, requests []types.InstanceRequest) (interface{}, error) {
	var createdInstances []types.InstanceInfo

	for _, req := range requests {
		sshKeyPath := i.getAnsibleSSHKeyPath(req)

		config := types.InstanceConfig{
			Region:          req.Region,
			Size:            req.Size,
			Image:           req.Image,
			SSHKeys:         []string{sshKeyPath},
			HypervisorID:    req.HypervisorID,
			HypervisorGroup: req.HypervisorGroup,
		}

		instances, err := i.provider.CreateInstance(ctx, req.Name, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create instance %s: %w", req.Name, err)
		}

		if len(instances) == 0 {
			return nil, fmt.Errorf("no instances created for %s", req.Name)
		}

		createdInstances = append(createdInstances, instances...)

		if err := i.provisioner.Provision(instances[0], sshKeyPath); err != nil {
			return nil, fmt.Errorf("failed to provision instance %s: %w", req.Name, err)
		}
	}

	return createdInstances, nil
}

func (i *Infrastructure) getAnsibleSSHKeyPath(req types.InstanceRequest) string {
	// Check for custom SSH key path in request
	if req.SSHKeyPath != "" {
		return req.SSHKeyPath
	}

	// Check for SSH key path in environment
	if envPath := os.Getenv("TALIS_SSH_KEY_PATH"); envPath != "" {
		return envPath
	}

	// Default to ~/.ssh/id_rsa
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "id_rsa")
}

// RunProvisioning executes the provisioning process for the created instances
func (i *Infrastructure) RunProvisioning(instances []types.InstanceInfo) error {
	if i.provisioner == nil {
		return fmt.Errorf("provisioner not initialized")
	}

	if len(instances) == 0 {
		return fmt.Errorf("no instances provided for provisioning")
	}

	// Get SSH key path for Ansible
	sshKeyPath := i.getAnsibleSSHKeyPath(types.InstanceRequest{})

	// Run provisioning for each instance
	for _, instance := range instances {
		if err := i.provisioner.Provision(instance, sshKeyPath); err != nil {
			return fmt.Errorf("failed to provision instance %s: %w", instance.Name, err)
		}
	}

	return nil
}
