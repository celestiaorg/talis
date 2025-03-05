package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/celestiaorg/talis/internal/compute"
)

// Infrastructure represents the infrastructure management
type Infrastructure struct {
	name        string                  // Name of the infrastructure
	projectName string                  // Name of the project
	instances   []InstanceRequest       // Instance configuration
	provider    compute.ComputeProvider // Compute provider implementation
	provisioner compute.Provisioner
	jobID       string
	action      string // Action to perform (create/delete)
}

// NewInfrastructure creates a new infrastructure instance
func NewInfrastructure(req *JobRequest) (*Infrastructure, error) {
	provider, err := compute.NewComputeProvider(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute provider: %w", err)
	}

	// Generate job ID using timestamp
	timestamp := time.Now().Format("20060102-150405")
	jobID := fmt.Sprintf("job-%s", timestamp)

	return &Infrastructure{
		name:        req.Name,
		projectName: req.ProjectName,
		instances:   req.Instances,
		provider:    provider,
		provisioner: compute.NewProvisioner(jobID),
		jobID:       jobID,
		action:      req.Action,
	}, nil
}

// Execute executes the infrastructure operation
func (i *Infrastructure) Execute() (interface{}, error) {
	fmt.Printf("🚀 Creating infrastructure...\n")

	var result interface{}
	var err error

	switch i.action {
	case "create":
		instances := make([]InstanceInfo, 0)
		for _, instance := range i.instances {
			for j := 0; j < instance.NumberOfInstances; j++ {
				// Create instance name with index if multiple instances
				instanceName := i.name
				if instance.NumberOfInstances > 1 {
					instanceName = fmt.Sprintf("%s-%d", i.name, j)
				}

				info, err := i.provider.CreateInstance(context.Background(), instanceName, compute.InstanceConfig{
					Region:   instance.Region,
					Size:     instance.Size,
					Image:    instance.Image,
					SSHKeyID: instance.SSHKeyName,
					Tags:     instance.Tags,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to create instance %s: %w", instanceName, err)
				}
				instances = append(instances, InstanceInfo{
					Name: instanceName,
					IP:   info.PublicIP,
				})
			}
		}
		result = instances

	case "delete":
		fmt.Printf("🗑️ Deleting infrastructure...\n")
		for _, instance := range i.instances {
			// Use the instance name directly for deletion
			instanceName := i.name
			fmt.Printf("🗑️ Deleting %s droplet: %s\n", instance.Provider, instanceName)

			if err := i.provider.DeleteInstance(context.Background(), instanceName); err != nil {
				return nil, fmt.Errorf("failed to delete instance %s: %w", instanceName, err)
			}
		}
		result = map[string]string{
			"status": "deleted",
		}

	default:
		return nil, fmt.Errorf("unsupported action: %s", i.action)
	}

	return result, err
}

// handleCreate handles the creation of infrastructure
func (i *Infrastructure) handleCreate() (interface{}, error) {
	fmt.Println("🚀 Creating infrastructure...")

	// Create instances
	instances := make([]InstanceInfo, 0)
	for _, inst := range i.instances {
		for idx := 0; idx < inst.NumberOfInstances; idx++ {
			name := fmt.Sprintf("%s-%d", i.name, idx)
			info, err := i.provider.CreateInstance(context.Background(), name, compute.InstanceConfig{
				Region:   inst.Region,
				Size:     inst.Size,
				Image:    inst.Image,
				SSHKeyID: inst.SSHKeyName,
				Tags:     inst.Tags,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create instance %s: %w", name, err)
			}
			instances = append(instances, InstanceInfo{
				Name: name,
				IP:   info.PublicIP,
			})
		}
	}

	// Give instances a moment to fully initialize
	fmt.Println("⏳ Waiting for instances to be fully ready...")
	time.Sleep(30 * time.Second)

	// Run Nix provisioning if enabled
	if err := i.RunProvisioning(instances); err != nil {
		return nil, fmt.Errorf("failed to provision instances: %w", err)
	}

	return instances, nil
}

// handleDelete handles the deletion of infrastructure
func (i *Infrastructure) handleDelete() (interface{}, error) {
	fmt.Println("🗑️ Deleting infrastructure...")

	// Delete instances
	for _, inst := range i.instances {
		for idx := 0; idx < inst.NumberOfInstances; idx++ {
			name := fmt.Sprintf("%s-%d", i.name, idx)
			if err := i.provider.DeleteInstance(context.Background(), name); err != nil {
				return nil, fmt.Errorf("failed to delete instance %s: %w", name, err)
			}
		}
	}

	return map[string]string{"status": "deleted"}, nil
}

// updateJobStatus updates the job status (this es un placeholder - implementar con una DB real)
//
//nolint:unused // Will be used in future implementation
func (i *Infrastructure) updateJobStatus(jobID, status string, result interface{}) {
	// TODO: Store in database and trigger webhook if configured
	fmt.Printf("Job %s status updated to %s: %v\n", jobID, status, result)
}
