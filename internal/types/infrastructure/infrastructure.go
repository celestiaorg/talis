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
