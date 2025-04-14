// Package services provides business logic implementation for the API
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// Infrastructure represents the infrastructure management system.
// It coordinates the creation and deletion of cloud resources across different providers
// and handles the provisioning of those resources using configuration management tools.
type Infrastructure struct {
	name        string                  // Name of the infrastructure
	projectName string                  // Name of the project
	instances   []types.InstanceRequest // Instance configuration
	provider    compute.Provider        // Compute provider implementation
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

	default:
		logger.Errorf("unsupported action: %s", i.action)
		return nil, fmt.Errorf("unsupported action: %s", i.action)
	}

	return result, err
}

// RunProvisioning applies Ansible configuration to all instances
func (i *Infrastructure) RunProvisioning(instances []types.InstanceInfo) error {
	// Check if any instance requires provisioning
	needsProvisioning := false
	for _, inst := range i.instances {
		if inst.Provision {
			needsProvisioning = true
			break
		}
	}

	if !needsProvisioning {
		fmt.Println("⏭️ Skipping provisioning as requested")
		return nil
	}

	fmt.Println("⚙️ Starting Ansible provisioning...")

	// Create inventory file for all instances
	instanceMap := make(map[string]string)
	for _, instance := range instances {
		instanceMap[instance.Name] = instance.PublicIP
	}

	// Create inventory file with the user's SSH key
	sshKeyPath := os.ExpandEnv("$HOME/.ssh/id_rsa")
	if err := i.provisioner.CreateInventory(instanceMap, sshKeyPath); err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	// --- Prepare Extra Variables ---
	extraVars := make(map[string]interface{})
	hostPayloadVars := make(map[string]map[string]interface{}) // Per-host payload vars

	// Create a map for easy lookup of request by instance name
	requestMap := make(map[string]types.InstanceRequest)
	for _, req := range i.instances { // Use i.instances which holds the original requests
		// Handle potential naming variations if necessary (e.g., base name + index)
		if req.NumberOfInstances == 1 {
			requestMap[req.Name] = req
		} else {
			baseName := req.Name
			for idx := 0; idx < req.NumberOfInstances; idx++ {
				instanceName := fmt.Sprintf("%s-%d", baseName, idx)
				requestMap[instanceName] = req
			}
		}
	}

	// Loop through provisioned instances to prepare payload vars
	for _, pInstance := range instances {
		instanceVars := make(map[string]interface{})
		instanceVars["payload_present"] = false // Default

		// Find the corresponding request
		if req, ok := requestMap[pInstance.Name]; ok && req.PayloadPath != "" {
			logger.Debugf("Preparing payload vars for instance %s from path %s", pInstance.Name, req.PayloadPath)
			// Determine destination path
			destFilename := filepath.Base(req.PayloadPath)
			destPath := filepath.Join("/root", destFilename) // Use filepath.Join for safety

			// Set vars for this host
			instanceVars["payload_present"] = true
			instanceVars["payload_src_path"] = req.PayloadPath // Pass the source path
			instanceVars["payload_dest_path"] = destPath
			instanceVars["payload_execute"] = req.ExecutePayload
			logger.Debugf("Payload vars for %s: src=%s, dest=%s, execute=%t", pInstance.Name, req.PayloadPath, destPath, req.ExecutePayload)
		}
		// Add host-specific vars under the instance name key
		hostPayloadVars[pInstance.Name] = instanceVars
	}

	// Add the per-host payload vars to the main extraVars
	extraVars["host_payload_vars"] = hostPayloadVars

	// Add other necessary extra vars if needed (e.g., common settings)
	extraVars["target_hosts"] = "all"

	// Configure hosts in parallel
	errChan := make(chan error, len(instances))
	var wg sync.WaitGroup

	for _, instance := range instances {
		wg.Add(1)
		go func(inst types.InstanceInfo) {
			defer wg.Done()
			if err := i.provisionInstance(inst); err != nil {
				errChan <- fmt.Errorf("failed to provision %s: %w", inst.Name, err)
			}
		}(instance)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Run Ansible playbook with prepared extra variables
	fmt.Println("📝 Running Ansible playbook...")
	if err := i.provisioner.RunAnsiblePlaybook(i.jobID, extraVars); err != nil {
		return fmt.Errorf("failed to run Ansible playbook: %w", err)
	}

	fmt.Println("✅ Ansible playbook completed successfully")
	return nil
}

// provisionInstance configures a single instance with Ansible
func (i *Infrastructure) provisionInstance(instance types.InstanceInfo) error {
	fmt.Printf("🔧 Starting provisioning for %s (%s)...\n", instance.Name, instance.PublicIP)

	// Use the user's SSH key path
	sshKeyPath := os.ExpandEnv("$HOME/.ssh/id_rsa")
	if err := i.provisioner.ConfigureHost(instance.PublicIP, sshKeyPath); err != nil {
		return fmt.Errorf("failed to configure host: %w", err)
	}

	fmt.Printf("✅ Provisioning completed for %s\n", instance.Name)
	return nil
}
