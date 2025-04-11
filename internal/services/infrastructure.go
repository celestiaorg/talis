// Package services provides business logic implementation for the API
package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/events"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// Infrastructure represents the infrastructure management system.
// It coordinates the creation and deletion of cloud resources across different providers.
type Infrastructure struct {
	name        string                  // Name of the infrastructure
	projectName string                  // Name of the project
	instances   []types.InstanceRequest // Instance configuration
	provider    compute.Provider        // Compute provider implementation
	jobID       string                  // Job ID
	jobName     string                  // Job name
	action      string                  // Action to perform (create/delete)
}

// NewInfrastructure creates a new infrastructure instance with the specified configuration.
// It initializes the appropriate cloud provider based on the request.
func NewInfrastructure(req *types.InstancesRequest) (*Infrastructure, error) {
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
		jobID:       jobID,
		jobName:     req.JobName,
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
		logger.Info("🚀 Creating infrastructure...")
		instances := make([]types.InstanceInfo, 0)
		for _, instance := range i.instances {
			// Use instance name if provided, otherwise use base name
			instanceName := instance.Name
			if instanceName == "" {
				instanceName = i.name
			}

			info, err := i.provider.CreateInstance(context.Background(), instanceName, types.InstanceConfig{
				Region:            instance.Region,
				OwnerID:           instance.OwnerID,
				Size:              instance.Size,
				Image:             instance.Image,
				SSHKeyID:          instance.SSHKeyName,
				Tags:              instance.Tags,
				NumberOfInstances: instance.NumberOfInstances,
				CustomName:        instance.Name,
				Volumes:           instance.Volumes,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create instances in region %s: %w", instance.Region, err)
			}
			// Convert types.InstanceInfo to our InstanceInfo and add to result
			for _, instanceInfo := range info {
				instances = append(instances, types.InstanceInfo{
					Name:          instanceInfo.Name,
					PublicIP:      instanceInfo.PublicIP,
					Provider:      instanceInfo.Provider,
					Tags:          instanceInfo.Tags,
					Region:        instanceInfo.Region,
					Size:          instanceInfo.Size,
					Volumes:       instanceInfo.Volumes,
					VolumeDetails: instanceInfo.VolumeDetails,
				})
			}
		}

		// Emit event for created instances
		events.Publish(events.Event{
			Type:      events.EventInstancesCreated,
			JobID:     i.jobID,
			JobName:   i.jobName,
			OwnerID:   i.instances[0].OwnerID,
			Instances: instances,
			Requests:  i.instances,
		})

		result = instances

	case "delete":
		logger.Info("🗑️ Deleting infrastructure...")
		var wg sync.WaitGroup
		deletedInstancesChan := make(chan string, 100)
		errorsChan := make(chan error, 100)

		for _, instance := range i.instances {
			// If instance has a custom name, use that directly
			if instance.Name != "" {
				wg.Add(1)
				go func(name string, region string) {
					defer wg.Done()
					logger.Infof("🗑️ Deleting %s droplet: %s in region %s", instance.Provider, name, region)

					if err := i.provider.DeleteInstance(context.Background(), name, region); err != nil {
						if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
							logger.Warnf("⚠️ Warning: Instance %s in region %s was already deleted", name, region)
							return
						}
						errorsChan <- fmt.Errorf("failed to delete instance %s in region %s: %w", name, region, err)
						return
					}
					deletedInstancesChan <- name
					logger.Debugf("✅ Successfully deleted instance: %s", name)
				}(instance.Name, instance.Region)
				continue
			}

			// If no custom name, try base name and indexed names
			if i.name != "" {
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
						logger.Debugf("✅ Successfully deleted instance: %s", name)
						return
					}
				}(i.name, instance.Region)

				// Try indexed names
				for j := 0; j < instance.NumberOfInstances; j++ {
					instanceName := fmt.Sprintf("%s-%d", i.name, j)
					wg.Add(1)
					go func(name string, region string) {
						defer wg.Done()
						logger.Infof("🗑️ Deleting %s droplet: %s in region %s", instance.Provider, name, region)

						if err := i.provider.DeleteInstance(context.Background(), name, region); err != nil {
							if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
								logger.Warnf("⚠️ Warning: Instance %s in region %s was already deleted", name, region)
								return
							}
							errorsChan <- fmt.Errorf("failed to delete instance %s in region %s: %w", name, region, err)
							return
						}
						deletedInstancesChan <- name
						logger.Debugf("✅ Successfully deleted instance: %s", name)
					}(instanceName, instance.Region)
				}
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
			logger.Errorf("errors during deletion: %v", errors)
			return nil, fmt.Errorf("errors during deletion: %v", errors)
		}

		logger.Infof("✅ Successfully deleted %d instances", len(deletedInstances))
		result = map[string]interface{}{
			"status":  "deleted",
			"deleted": deletedInstances,
			"count":   len(deletedInstances),
		}

		// Emit event for deleted instances
		events.Publish(events.Event{
			Type:    events.EventInstancesDeleted,
			JobID:   i.jobID,
			JobName: i.jobName,
		})

	default:
		logger.Errorf("unsupported action: %s", i.action)
		return nil, fmt.Errorf("unsupported action: %s", i.action)
	}

	return result, err
}
