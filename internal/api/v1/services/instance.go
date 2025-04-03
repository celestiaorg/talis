package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
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
		logger.Info("🚀 Starting async infrastructure creation...")

		// Get the job name from the database
		job, err := s.jobService.jobRepo.GetByID(ctx, 0, jobID) // ownerID 0 for now since we're not using it yet
		if err != nil {
			logger.Errorf("❌ Failed to get job details: %v", err)
			return
		}

		// Create a JobRequest for provisioning
		infra, err := infrastructure.NewInfrastructure(&infrastructure.InstancesRequest{
			JobName:     job.Name,
			Instances:   instances,
			ProjectName: "example-project", // TODO: replace it with the actual project name in another PR
			Action:      "create",
			Provider:    instances[0].Provider, // Use the provider directly from the instance
		})
		if err != nil {
			logger.Errorf("❌ Failed to create infrastructure: %v", err)
			return
		}

		result, err := infra.Execute()
		if err != nil {
			// Check if error is due to resource not found
			if strings.Contains(err.Error(), "404") &&
				strings.Contains(err.Error(), "could not be found") {
				logger.Warn("⚠️ Warning: Some old resources were not found (already deleted)")
			} else {
				logger.Errorf("❌ Failed to execute infrastructure: %v", err)
				return
			}
		}

		// Update instance information in database
		pInstances, ok := result.([]types.InstanceInfo)
		if !ok {
			logger.Errorf("❌ Invalid result type: %T", result)
			return
		}

		logger.Infof("📝 Created instances: %+v", pInstances)

		// Update instance information in database
		for _, instance := range pInstances {
			// Update IP and status
			if err := s.repo.UpdateIPByName(ctx, instance.Name, instance.PublicIP); err != nil {
				logger.Errorf("❌ Failed to update instance %s IP: %v", instance.Name, err)
				continue
			}
			logger.Infof("✅ Updated instance %s with IP %s", instance.Name, instance.PublicIP)

			// Update volumes
			if len(instance.Volumes) > 0 {
				if err := s.updateInstanceVolumes(ctx, instance.Name, instance.Volumes, instances[0].Volumes); err != nil {
					logger.Errorf("❌ %v", err)
					continue
				}
			}

			if err := s.repo.UpdateStatusByName(ctx, instance.Name, models.InstanceStatusReady); err != nil {
				logger.Errorf("❌ Failed to update instance %s status: %v", instance.Name, err)
				continue
			}
			logger.Infof("✅ Updated instance %s status to ready", instance.Name)
		}

		// Start Ansible provisioning if requested
		if instances[0].Provision {
			if err := infra.RunProvisioning(pInstances); err != nil {
				logger.Errorf("❌ Failed to run provisioning: %v", err)
				return
			}
		}

		logger.Infof("✅ Infrastructure creation completed for job ID %d", jobID)
	}()
}

// updateInstanceVolumes updates the volumes and volume details for an instance
func (s *Instance) updateInstanceVolumes(
	ctx context.Context,
	instanceName string,
	volumes []string,
	volumeConfigs []types.VolumeConfig,
) error {
	// Get instance from database
	dbInstance, err := s.repo.GetByName(ctx, instanceName)
	if err != nil {
		return fmt.Errorf("failed to get instance %s from database: %w", instanceName, err)
	}

	// Update volumes
	dbInstance.Volumes = volumes

	// Add volume details
	var volumeDetails []models.VolumeDetail
	for i, volumeID := range volumes {
		// Get the volume configuration from the request
		if i < len(volumeConfigs) {
			volumeConfig := volumeConfigs[i]
			volumeDetails = append(volumeDetails, models.VolumeDetail{
				ID:         volumeID,
				Name:       volumeConfig.Name,
				SizeGB:     volumeConfig.SizeGB,
				Region:     volumeConfig.Region,
				MountPoint: volumeConfig.MountPoint,
			})
		}
	}
	dbInstance.VolumeDetails = volumeDetails

	// Update instance in database
	if err := s.repo.Update(ctx, dbInstance.ID, dbInstance); err != nil {
		return fmt.Errorf("failed to update instance %s volumes: %w", instanceName, err)
	}

	logger.Infof("✅ Updated instance %s with volumes: %v and details: %+v", instanceName, volumes, volumeDetails)
	return nil
}

// GetInstancesByJobID retrieves all instances for a specific job
func (s *Instance) GetInstancesByJobID(ctx context.Context, jobID uint) ([]models.Instance, error) {
	logger.Infof("📥 Getting instances for job ID %d from database...", jobID)

	// Get instances for the specific job
	instances, err := s.repo.GetByJobID(ctx, jobID)
	if err != nil {
		logger.Errorf("❌ Error getting instances for job %d: %v", jobID, err)
		return nil, fmt.Errorf("failed to get instances for job %d: %w", jobID, err)
	}

	logger.Infof("✅ Retrieved %d instances for job %d from database", len(instances), jobID)
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

// terminate handles the infrastructure deletion process
func (s *Instance) terminate(ctx context.Context, jobName string, instances []models.Instance) {
	go func() {
		if len(instances) == 0 {
			logger.Error("❌ No instances found to terminate")
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
			logger.Infof("🗑️ Attempting to delete instance: %s", instance.Name)

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
				logger.Errorf("❌ Failed to create infrastructure client for instance %s: %v", instance.Name, err)
				continue
			}

			// Execute the deletion for this specific instance
			_, err = infra.Execute()
			if err != nil && !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") {
				logger.Errorf("❌ Error deleting instance %s: %v", instance.Name, err)
				continue
			}

			// If we get here, either:
			// 1. The deletion was successful
			// 2. The instance was already deleted (404/not found error)
			// In both cases, we want to mark it as terminated in our database
			if err := s.repo.Terminate(ctx, instance.ID); err != nil {
				logger.Errorf("❌ Failed to mark instance %s as terminated in database: %v", instance.Name, err)
				continue
			}
			logger.Infof("✅ Marked instance %s as terminated in database", instance.Name)
			if deleted, ok := deletionResult["deleted"].([]string); ok {
				deletionResult["deleted"] = append(deleted, instance.Name)
			}
		}

		deletionResult["status"] = "completed"
	}()
}
