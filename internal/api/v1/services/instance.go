package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// Instance provides business logic for instance operations
type Instance struct {
	repo       *repos.InstanceRepository
	jobService *Job
}

// NewInstanceService creates a new instance service instance
func NewInstanceService(repo *repos.InstanceRepository, jobService *Job) *Instance {
	return &Instance{
		repo:       repo,
		jobService: jobService,
	}
}

// ListInstances retrieves a paginated list of instances
func (s *Instance) ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	return s.repo.List(ctx, opts)
}

// CreateInstance creates a new instance
func (s *Instance) CreateInstance(ctx context.Context, ownerID uint, jobName string, instances []infrastructure.InstanceRequest) error {
	job, err := s.jobService.jobRepo.GetByName(ctx, ownerID, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	for _, i := range instances {
		baseName := i.Name
		if baseName == "" {
			baseName = fmt.Sprintf("instance-%s", uuid.New().String())
		}

		// Create multiple instances if requested
		numInstances := i.NumberOfInstances
		if numInstances < 1 {
			numInstances = 1
		}

		for idx := 0; idx < numInstances; idx++ {
			instanceName := baseName
			if numInstances > 1 {
				instanceName = fmt.Sprintf("%s-%d", baseName, idx)
			}

			instance := &models.Instance{
				Name:       instanceName,
				JobID:      job.ID,
				ProviderID: i.Provider,
				Status:     models.InstanceStatusPending,
				Region:     i.Region,
				Size:       i.Size,
			}

			if err := s.repo.Create(ctx, instance); err != nil {
				return fmt.Errorf("failed to create instance: %w", err)
			}
		}
	}

	// Start provisioning in background
	go s.provisionInstances(ctx, job.ID, instances)

	return nil
}

// GetInstance retrieves an instance by ID
func (s *Instance) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, id)
}

// provisionInstances provisions the job asynchronously
func (s *Instance) provisionInstances(ctx context.Context, jobID uint, instances []infrastructure.InstanceRequest) {
	// TODO: do something with the logs as now it makes the server logs messy
	go func() {
		fmt.Println("🚀 Starting async infrastructure creation...")

		// TODO: need to update the instance status to provisioning

		// Create a JobRequest for provisioning
		infra, err := infrastructure.NewInfrastructure(&infrastructure.InstancesRequest{
			JobName:     "example-job", // Replace with actual job name
			Instances:   instances,
			ProjectName: "example-project", // TODO: replace it with the actual project name in another PR
			Action:      "create",
			Provider:    instances[0].Provider, // Use the provider directly from the instance
		})
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure: %v\n", err)
			return
		}

		result, err := infra.Execute()
		if err != nil {
			// Check if error is due to resource not found
			if strings.Contains(err.Error(), "404") &&
				strings.Contains(err.Error(), "could not be found") {
				fmt.Printf("⚠️ Warning: Some old resources were not found (already deleted)\n")
			} else {
				fmt.Printf("❌ Failed to execute infrastructure: %v\n", err)
				return
			}
		}

		// Update instance information in database
		pInstances, ok := result.([]infrastructure.InstanceInfo)
		if !ok {
			fmt.Printf("❌ Invalid result type: %T\n", result)
			return
		}

		fmt.Printf("📝 Created instances: %+v\n", pInstances)

		// Update instance information in database
		for _, instance := range pInstances {
			// Update IP and status
			if err := s.repo.UpdateIPByName(ctx, instance.Name, instance.IP); err != nil {
				fmt.Printf("❌ Failed to update instance %s IP: %v\n", instance.Name, err)
				continue
			}
			fmt.Printf("✅ Updated instance %s with IP %s\n", instance.Name, instance.IP)

			if err := s.repo.UpdateStatusByName(ctx, instance.Name, models.InstanceStatusReady); err != nil {
				fmt.Printf("❌ Failed to update instance %s status: %v\n", instance.Name, err)
				continue
			}
			fmt.Printf("✅ Updated instance %s status to ready\n", instance.Name)
		}

		// Start Ansible provisioning if requested
		if instances[0].Provision {
			if err := infra.RunProvisioning(pInstances); err != nil {
				fmt.Printf("❌ Failed to run provisioning: %v\n", err)
				return
			}
		}

		fmt.Printf("✅ Infrastructure creation completed for job ID %d\n", jobID)
	}()
}

// GetInstancesByJobID retrieves all instances for a specific job
func (s *Instance) GetInstancesByJobID(ctx context.Context, jobID uint) ([]models.Instance, error) {
	fmt.Printf("📥 Getting instances for job ID %d from database...\n", jobID)

	// Get instances for the specific job
	instances, err := s.repo.GetByJobID(ctx, jobID)
	if err != nil {
		fmt.Printf("❌ Error getting instances for job %d: %v\n", jobID, err)
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}

	fmt.Printf("✅ Retrieved %d instances for job %d from database\n", len(instances), jobID)
	return instances, nil
}

// Terminate handles the termination of instances for a given job name and instance names.
func (s *Instance) Terminate(ctx context.Context, ownerID uint, jobName string, instanceNames []string) error {
	// First verify the job exists and belongs to the owner
	job, err := s.jobService.jobRepo.GetByName(ctx, ownerID, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	// Get instances that belong to this job and match the provided names
	instances, err := s.repo.GetByJobIDAndNames(ctx, job.ID, instanceNames)
	if err != nil {
		return fmt.Errorf("failed to get instances: %w", err)
	}

	// Verify we found all requested instances
	if len(instances) == 0 {
		logger.Infof("No active instances found with the specified names for job '%s', request is a no-op", jobName)
		return nil
	}
	if len(instances) != len(instanceNames) {
		// Some instances were not found, log which ones
		foundNames := make(map[string]bool)
		for _, instance := range instances {
			foundNames[instance.Name] = true
		}
		var missingNames []string
		for _, name := range instanceNames {
			if !foundNames[name] {
				missingNames = append(missingNames, name)
			}
		}
		logger.Infof("Some instances were not found or are already deleted for job '%s': %v", jobName, missingNames)
	}

	s.terminate(ctx, job.Name, instances)
	return nil
}

// handleInfrastructureDeletion handles the infrastructure deletion process
func (s *Instance) terminate(ctx context.Context, jobName string, instances []models.Instance) {
	go func() {
		if len(instances) == 0 {
			fmt.Printf("❌ No instances found to terminate\n")
			return
		}

		// Prepare deletion result
		deletionResult := map[string]interface{}{
			"status":  "deleting",
			"deleted": []string{},
		}

		// Try to delete each selected instance
		// TODO: Consider async deletion in multiple goroutines
		for _, instance := range instances {
			fmt.Printf("🗑️ Attempting to delete instance: %s\n", instance.Name)

			// Create a new infrastructure request for each instance
			instanceInfraReq := &infrastructure.InstancesRequest{
				JobName: jobName,
				Instances: []infrastructure.InstanceRequest{
					{
						Name:     instance.Name,
						Provider: instance.ProviderID,
						Region:   instance.Region,
						Size:     instance.Size,
					},
				},
				Action: "delete",
			}

			// TODO: a hacky fix, but has to be addressed properly in another PR
			if instanceInfraReq.Provider == "" {
				instanceInfraReq.Provider = instanceInfraReq.Instances[0].Provider
			}

			// Create infrastructure client for this specific instance
			infra, err := infrastructure.NewInfrastructure(instanceInfraReq)
			if err != nil {
				fmt.Printf("❌ Failed to create infrastructure client for instance %s: %v\n", instance.Name, err)
				continue
			}

			// Execute the deletion for this specific instance
			_, err = infra.Execute()
			if err != nil {
				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
					fmt.Printf("⚠️ Warning: Instance %s was already deleted\n", instance.Name)
					// Instance doesn't exist in DO, safe to mark as deleted
					if err := s.repo.Terminate(ctx, instance.ID); err != nil {
						fmt.Printf("❌ Failed to mark instance %s as terminated in database: %v\n", instance.Name, err)
						continue
					}
					fmt.Printf("✅ Marked instance %s as terminated in database\n", instance.Name)
					if deleted, ok := deletionResult["deleted"].([]string); ok {
						deletionResult["deleted"] = append(deleted, instance.Name)
					}
				} else {
					fmt.Printf("❌ Error deleting instance %s: %v\n", instance.Name, err)
					continue
				}
			}

			// Deletion was successful, update database
			if err := s.repo.Terminate(ctx, instance.ID); err != nil {
				fmt.Printf("❌ Failed to mark instance %s as terminated in database: %v\n", instance.Name, err)
				continue
			}
			fmt.Printf("✅ Marked instance %s as terminated in database\n", instance.Name)
			if deleted, ok := deletionResult["deleted"].([]string); ok {
				deletionResult["deleted"] = append(deleted, instance.Name)
			}
		}

		deletionResult["status"] = "completed"
	}()
}
