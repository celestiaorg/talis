// Package services provides business logic implementation for the API
package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// Infrastructure represents the infrastructure management system.
// It coordinates the creation and deletion of cloud resources across different providers
// and handles the provisioning of those resources using configuration management tools.
type Infrastructure struct {
	provider    compute.Provider        // Compute provider implementation
	provisioner compute.Provisioner
	jobID       string
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
func NewInfrastructure(providerID models.ProviderID) (*Infrastructure, error) {
	provider, err := compute.NewComputeProvider(providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute provider: %w", err)
	}

	// Generate job ID using timestamp for unique identification
	timestamp := time.Now().Format("20060102-150405")
	jobID := fmt.Sprintf("job-%s", timestamp)

	return &Infrastructure{
		provider:    provider,
		provisioner: compute.NewProvisioner(jobID),
		jobID:       jobID,
	}, nil
}

// Create performs the requested infrastructure operation (create).
// For creation, it spawns the requested number of instances with the specified configuration.
//
// Returns:
//   - interface{}: The result of the operation
//   - For creation: []InstanceInfo containing details of created instances
//   - error: Any error that occurred during execution
func (i *Infrastructure) Create(instances []types.InstanceRequest) error {
		logger.Info("🚀 Creating infrastructure...")
		for _, instance := range instances {
			// Use instance name if provided, otherwise use base name
			instanceName := instance.Name
			if instanceName == "" {
				instanceName = i.name
			}

			info, err := i.provider.CreateInstance(context.Background(), instance)
			if err != nil {
				return fmt.Errorf("failed to create instances in region %s: %w", instance.Region, err)
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
					// NOTE: the provider CreateInstance doesn't do anything with the payload, so we need to pass it through here
					PayloadPath:    instance.PayloadPath,
					ExecutePayload: instance.ExecutePayload,
				})
			}
		}
		result = instances


	return result, err
}

// Delete performs the requested infrastructure operation (delete).
// For deletion, it removes the specified instances from the cloud provider.
//
// Returns:
//   - interface{}: The result of the operation
//   - For deletion: map[string]interface{} with operation status and deleted instances
//   - error: Any error that occurred during execution
func (i *Infrastructure) Delete() (interface{}, error) {
	var result interface{}
	var err error

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


	return result, err
}

// getAnsibleSSHKeyPath determines the appropriate SSH private key path for Ansible
// based on the instance requests, prioritizing custom paths, then key types,
// and falling back to the default RSA key path.
// It assumes the key configuration from the first request applies to the whole job.
func getAnsibleSSHKeyPath(instanceRequests []types.InstanceRequest) string {
	sshKeyPath := "$HOME/.ssh/id_rsa" // Default
	if len(instanceRequests) > 0 {
		firstReq := instanceRequests[0]
		if firstReq.SSHKeyPath != "" { // Priority 1: Custom path
			sshKeyPath = firstReq.SSHKeyPath
		} else if firstReq.SSHKeyType != "" { // Priority 2: Key type
			switch strings.ToLower(firstReq.SSHKeyType) {
			case "ed25519":
				sshKeyPath = "$HOME/.ssh/id_ed25519"
			case "ecdsa":
				sshKeyPath = "$HOME/.ssh/id_ecdsa"
				// Add other types if needed
			}
		}
	}
	// Expand environment variables like $HOME (Ansible handles this, but doing it here is safe)
	return os.ExpandEnv(sshKeyPath)
}

// RunProvisioning applies Ansible configuration to all instances
func (i *Infrastructure) RunProvisioning(instances []types.InstanceInfo) error {
	// Check if any instance requires provisioning based on the original request config
	needsProvisioning := false
	for _, reqInst := range i.instances { // i.instances holds InstanceRequest
		if reqInst.Provision {
			needsProvisioning = true
			break
		}
	}

	if !needsProvisioning {
		fmt.Println("⏭️ Skipping provisioning as requested")
		return nil
	}

	if len(instances) == 0 {
		logger.Warnf("No instances returned from provider for job %s, cannot provision.", i.jobID)
		return nil
	}

	fmt.Println("⚙️ Starting Ansible provisioning...")

	// --- Step 1: Ensure all hosts are ready for SSH connections ---
	// TODO: This currently uses a default/shared key path assumption. Improve if needed.
	sshKeyPath := getAnsibleSSHKeyPath(i.instances)
	hosts := make([]string, len(instances))
	for idx, inst := range instances {
		hosts[idx] = inst.PublicIP
	}
	if err := i.provisioner.ConfigureHosts(hosts, sshKeyPath); err != nil {
		// ConfigureHosts already logs details
		return fmt.Errorf("failed to ensure SSH readiness for all hosts: %w", err)
	}

	// --- Step 2: Create the inventory file using InstanceInfo ---
	// CreateInventory now handles extracting info and determining the key path internally
	if err := i.provisioner.CreateInventory(instances, sshKeyPath); err != nil {
		return fmt.Errorf("failed to create Ansible inventory: %w", err)
	}

	// --- Step 3: Run the Ansible playbook ---
	fmt.Println("📝 Running Ansible playbook...")
	if err := i.provisioner.RunAnsiblePlaybook(i.jobID); err != nil {
		return fmt.Errorf("failed to run Ansible playbook: %w", err)
	}

	fmt.Println("✅ Ansible playbook completed successfully")
	return nil
}
