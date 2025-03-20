// Package infrastructure provides types and functionality for managing cloud infrastructure operations.
// It acts as an abstraction layer between the API and the actual cloud providers,
// handling both creation and deletion of cloud resources in a provider-agnostic way.
package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/compute/types"
)

// Infrastructure represents the infrastructure management system.
// It coordinates the creation and deletion of cloud resources across different providers
// and handles the provisioning of those resources using configuration management tools.
type Infrastructure struct {
	name        string                  // Name of the infrastructure
	projectName string                  // Name of the project
	instances   []InstanceRequest       // Instance configuration
	provider    compute.ComputeProvider // Compute provider implementation
	provisioner compute.Provisioner
	jobID       string
	action      string // Action to perform (create/delete)
}

// NewInfrastructure creates a new infrastructure instance with the specified configuration.
// It initializes the appropriate cloud provider and provisioner based on the request.
//
// Parameters:
//   - req: The job request containing the infrastructure configuration
//
// Returns:
//   - *Infrastructure: A configured infrastructure manager
//   - error: Any error that occurred during initialization
func NewInfrastructure(req *JobRequest) (*Infrastructure, error) {
	provider, err := compute.NewComputeProvider(req.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute provider: %w", err)
	}

	// Generate job ID using timestamp for unique identification
	timestamp := time.Now().Format("20060102-150405")
	jobID := fmt.Sprintf("job-%s", timestamp)

	return &Infrastructure{
		name:        req.InstanceName,
		projectName: req.ProjectName,
		instances:   req.Instances,
		provider:    provider,
		provisioner: compute.NewProvisioner(jobID),
		jobID:       jobID,
		action:      req.Action,
	}, nil
}

// Execute performs the requested infrastructure operation (create or delete).
// For creation, it spawns the requested number of instances with the specified configuration.
// For deletion, it removes the specified instances from the cloud provider.
//
// Returns:
//   - interface{}: The result of the operation
//   - For creation: []InstanceInfo containing details of created instances
//   - For deletion: map[string]interface{} with operation status and deleted instances
//   - error: Any error that occurred during execution
func (i *Infrastructure) Execute() (interface{}, error) {
	var result interface{}
	var err error

	switch i.action {
	case "create":
		fmt.Printf("🚀 Creating infrastructure...\n")
		instances := make([]InstanceInfo, 0)
		for _, instance := range i.instances {
			// Create all instances for this configuration at once
			instanceName := i.name
			info, err := i.provider.CreateInstance(context.Background(), instanceName, types.InstanceConfig{
				Region:            instance.Region,
				Size:              instance.Size,
				Image:             instance.Image,
				SSHKeyID:          instance.SSHKeyName,
				Tags:              instance.Tags,
				NumberOfInstances: instance.NumberOfInstances,
				CustomName:        instance.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create instances in region %s: %w", instance.Region, err)
			}
			// Convert compute.InstanceInfo to our InstanceInfo and add to result
			for _, instanceInfo := range info {
				instances = append(instances, InstanceInfo{
					Name:     instanceInfo.Name,
					IP:       instanceInfo.PublicIP,
					Provider: instanceInfo.Provider,
					Region:   instanceInfo.Region,
					Size:     instanceInfo.Size,
				})
			}
		}
		result = instances

	case "delete":
		fmt.Printf("🗑️ Deleting infrastructure...\n")
		var wg sync.WaitGroup
		deletedInstancesChan := make(chan string, 100)
		errorsChan := make(chan error, 100)

		for _, instance := range i.instances {
			// Try base name first
			wg.Add(1)
			go func(name string, region string) {
				defer wg.Done()
				if err := i.provider.DeleteInstance(context.Background(), name, region); err != nil {
					if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
						errorsChan <- fmt.Errorf("failed to delete instance %s in region %s: %w", name, region, err)
						return
					}
				} else {
					deletedInstancesChan <- name
					return
				}
			}(i.name, instance.Region)

			// Try indexed names
			for j := 0; j < instance.NumberOfInstances; j++ {
				instanceName := fmt.Sprintf("%s-%d", i.name, j)
				wg.Add(1)
				go func(name string, region string) {
					defer wg.Done()
					fmt.Printf("🗑️ Deleting %s droplet: %s in region %s\n", instance.Provider, name, region)

					if err := i.provider.DeleteInstance(context.Background(), name, region); err != nil {
						if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
							fmt.Printf("⚠️ Warning: Instance %s in region %s was already deleted\n", name, region)
							return
						}
						errorsChan <- fmt.Errorf("failed to delete instance %s in region %s: %w", name, region, err)
						return
					}
					deletedInstancesChan <- name
				}(instanceName, instance.Region)
			}
		}

		// Create a goroutine to close channels after WaitGroup is done
		go func() {
			wg.Wait()
			close(deletedInstancesChan)
			close(errorsChan)
		}()

		// Collect results
		var deletedInstances []string
		for name := range deletedInstancesChan {
			deletedInstances = append(deletedInstances, name)
		}

		// Check for errors
		var errors []error
		for err := range errorsChan {
			errors = append(errors, err)
		}

		if len(errors) > 0 {
			return nil, fmt.Errorf("errors during deletion: %v", errors)
		}

		result = map[string]interface{}{
			"status":  "deleted",
			"deleted": deletedInstances,
			"count":   len(deletedInstances),
		}

	default:
		return nil, fmt.Errorf("unsupported action: %s", i.action)
	}

	return result, err
}
