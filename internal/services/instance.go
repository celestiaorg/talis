// Package services provides business logic implementation for the API
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// Instance represents the instance service
type Instance struct {
	repo       *repos.InstanceRepository
	jobService *Job
}

// NewInstanceService creates a new instance service
func NewInstanceService(repo *repos.InstanceRepository, jobService *Job) *Instance {
	return &Instance{
		repo:       repo,
		jobService: jobService,
	}
}

// ListInstances retrieves a paginated list of instances
func (s *Instance) ListInstances(ctx context.Context, ownerID uint, opts *models.ListOptions) ([]models.Instance, error) {
	return s.repo.List(ctx, ownerID, opts)
}

// CreateInstance creates a new instance
func (s *Instance) CreateInstance(ctx context.Context, ownerID uint, jobName string, instances []types.InstanceRequest) error {
	job, err := s.jobService.jobRepo.GetByName(ctx, ownerID, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	for _, i := range instances {
		// Sanity check the ownerID fields.
		// TODO: this is a little verbose, maybe we can clean it up?
		bothZero := i.OwnerID == 0 && ownerID == 0
		bothNonZero := i.OwnerID != 0 && ownerID != 0
		// At least one of the ownerID fields is required
		if bothZero {
			return fmt.Errorf("instance owner_id is required")
		}
		// Sanity check that a user is not trying to create an instance for another user
		if bothNonZero && i.OwnerID != ownerID {
			return fmt.Errorf("instance owner_id does not match job owner_id")
		}
		// Ensure the instance owner_id is set
		if i.OwnerID == 0 {
			i.OwnerID = ownerID
		}

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
				Name:          instanceName,
				OwnerID:       i.OwnerID,
				JobID:         job.ID,
				ProviderID:    i.Provider,
				Status:        models.InstanceStatusPending,
				Region:        i.Region,
				Size:          i.Size,
				Tags:          i.Tags,
				Volumes:       []string{},
				VolumeDetails: models.VolumeDetails{},
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
func (s *Instance) GetInstance(ctx context.Context, ownerID, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, ownerID, id)
}

// updateInstanceVolumes updates the volumes and volume details for an instance
func (s *Instance) updateInstanceVolumes(
	ctx context.Context,
	instanceName string,
	volumes []string,
	volumeDetails []types.VolumeDetails,
) error {
	// Get the instance first to get its ownerID
	instance, err := s.repo.GetByName(ctx, instanceName)
	if err != nil {
		return fmt.Errorf("failed to get instance %s: %w", instanceName, err)
	}

	logger.Debugf("🔄 Converting volume details for instance %s", instanceName)
	logger.Debugf("📥 Input data:")
	logger.Debugf("  - Volumes: %v", volumes)
	logger.Debugf("  - Volume Details: %+v", volumeDetails)

	// Convert volume details to database model
	dbVolumeDetails := make(models.VolumeDetails, 0, len(volumeDetails))
	for _, vd := range volumeDetails {
		dbVolumeDetails = append(dbVolumeDetails, models.VolumeDetail{
			ID:         vd.ID,
			Name:       vd.Name,
			Region:     vd.Region,
			SizeGB:     vd.SizeGB,
			MountPoint: vd.MountPoint,
		})
	}

	logger.Debugf("📦 Preparing to update instance %s", instanceName)
	logger.Debugf("📝 Data to update:")
	logger.Debugf("  - Volumes: %#v", volumes)
	logger.Debugf("  - Volume Details: %#v", dbVolumeDetails)

	// Create update instance with only the fields we want to update
	updateInstance := &models.Instance{
		Volumes:       volumes,
		VolumeDetails: dbVolumeDetails,
	}

	// Update instance in database using a transaction to ensure atomicity
	err = s.repo.UpdateByName(ctx, instance.OwnerID, instanceName, updateInstance)
	if err != nil {
		logger.Errorf("❌ Failed to update instance %s volumes: %v", instanceName, err)
		return fmt.Errorf("failed to update instance %s volumes: %w", instanceName, err)
	}

	// Verify the update immediately
	updatedInstance, err := s.repo.GetByName(ctx, instanceName)
	if err != nil {
		logger.Warnf("⚠️ Could not verify volume update for instance %s: %v", instanceName, err)
		return nil // Don't return error here as the update might have succeeded
	}

	logger.Debugf("✅ Verified volumes update for instance %s:", instanceName)
	logger.Debugf("📊 Database state after update:")
	logger.Debugf("  - Volumes: %#v", updatedInstance.Volumes)
	logger.Debugf("  - Volume Details: %#v", updatedInstance.VolumeDetails)

	// Verify data integrity
	if len(updatedInstance.Volumes) != len(volumes) {
		logger.Warnf("⚠️ Volume count mismatch - Expected: %d, Got: %d", len(volumes), len(updatedInstance.Volumes))
	}
	if len(updatedInstance.VolumeDetails) != len(dbVolumeDetails) {
		logger.Warnf("⚠️ Volume details count mismatch - Expected: %d, Got: %d", len(dbVolumeDetails), len(updatedInstance.VolumeDetails))
	}

	return nil
}

// provisionInstances provisions the job asynchronously
func (s *Instance) provisionInstances(ctx context.Context, jobID uint, instances []types.InstanceRequest) {
	go func() {
		// Get job name
		job, err := s.jobService.jobRepo.GetByID(ctx, 0, jobID)
		if err != nil {
			logger.Errorf("❌ Failed to get job details: %v", err)
			return
		}

		// Create infrastructure client
		infraReq := &types.InstancesRequest{
			JobName:   job.Name,
			Instances: instances,
			Action:    "create",
			Provider:  instances[0].Provider,
		}

		infra, err := NewInfrastructure(infraReq)
		if err != nil {
			logger.Errorf("❌ Failed to create infrastructure client: %v", err)
			return
		}

		// Execute infrastructure creation
		result, err := infra.Execute()
		if err != nil {
			logger.Errorf("❌ Failed to create infrastructure: %v", err)
			return
		}

		// Type assert the result
		pInstances, ok := result.([]types.InstanceInfo)
		if !ok {
			logger.Errorf("❌ Invalid result type: %T", result)
			return
		}

		logger.Debugf("📝 Created instances: %+v", pInstances)

		// Update instances with IP and status
		// Update instance information in database
		ownerID := instances[0].OwnerID
		for _, instance := range pInstances {
			logger.Debugf("🔄 Processing instance update for %s", instance.Name)
			logger.Debugf("  - Volumes: %v", instance.Volumes)
			logger.Debugf("  - Volume Details: %+v", instance.VolumeDetails)

			// Update volumes first if present
			if len(instance.Volumes) > 0 || len(instance.VolumeDetails) > 0 {
				logger.Debugf("🔄 Updating volumes for instance %s", instance.Name)
				if err := s.updateInstanceVolumes(ctx, instance.Name, instance.Volumes, instance.VolumeDetails); err != nil {
					logger.Errorf("❌ Failed to update volumes for instance %s: %v", instance.Name, err)
					continue
				}
				logger.Debugf("✅ Successfully updated volumes for instance %s", instance.Name)

				// Verify the update was successful
				updatedInstance, err := s.repo.GetByName(ctx, instance.Name)
				if err != nil {
					logger.Errorf("❌ Failed to verify volume update for instance %s: %v", instance.Name, err)
				} else {
					logger.Debugf("📊 Current instance state after volume update:")
					logger.Debugf("  - Volumes: %v", updatedInstance.Volumes)
					logger.Debugf("  - Volume Details: %+v", updatedInstance.VolumeDetails)
				}
			}

			// Then update instance status and IP
			updateInstance := &models.Instance{
				PublicIP: instance.PublicIP,
				Status:   models.InstanceStatusReady,
			}

			if err := s.repo.UpdateByName(ctx, ownerID, instance.Name, updateInstance); err != nil {
				logger.Errorf("❌ Failed to update instance %s: %v", instance.Name, err)
				continue
			}
			logger.Debugf("✅ Updated instance %s with IP %s and status ready", instance.Name, instance.PublicIP)
		}

		// The event is now emitted in infrastructure.Execute(), so we don't need to emit it here
		logger.Debugf("✅ Infrastructure creation completed for job ID %d", jobID)
	}()
}

// GetInstancesByJobID retrieves all instances for a specific job
func (s *Instance) GetInstancesByJobID(ctx context.Context, ownerID uint, jobID uint) ([]models.Instance, error) {
	logger.Infof("📥 Getting instances for job ID [%d] from database...", jobID)

	// Get instances for the specific job
	instances, err := s.repo.GetByJobID(ctx, ownerID, jobID)
	if err != nil {
		logger.Errorf("❌ Error getting instances for job ID [%d]: %v", jobID, err)
		return nil, fmt.Errorf("failed to get instances for job ID [%d]: %w", jobID, err)
	}

	logger.Infof("✅ Retrieved %d instances for job ID [%d] from database", len(instances), jobID)
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
	instances, err := s.repo.GetByJobIDAndNames(ctx, ownerID, job.ID, instanceNames)
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

// deleteRequest represents a single instance deletion request with tracking
type deleteRequest struct {
	instance     models.Instance
	infraRequest *types.InstancesRequest
	attempts     int
	lastError    error
	maxAttempts  int
}

// deletionResults tracks the overall results of the deletion operation
type deletionResults struct {
	successful []string         // names of successfully deleted instances
	failed     map[string]error // instance name -> last error for failed deletions
}

// terminate handles the infrastructure deletion process
func (s *Instance) terminate(ctx context.Context, jobName string, instances []models.Instance) {
	go func() {
		if len(instances) == 0 {
			logger.Error("❌ No instances found to terminate")
			return
		}

		// Create queue of delete requests
		queue := make([]*deleteRequest, 0, len(instances))
		for _, instance := range instances {
			logger.Infof("🗑️ Attempting to delete instance: %s", instance.Name)

			// Create a new infrastructure request for each instance
			infraReq := &types.InstancesRequest{
				JobName: jobName,
				Instances: []types.InstanceRequest{
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
			if infraReq.Provider == "" {
				infraReq.Provider = infraReq.Instances[0].Provider
			}

			queue = append(queue, &deleteRequest{
				instance:     instance,
				infraRequest: infraReq,
				maxAttempts:  10,
			})
		}

		results := &deletionResults{
			successful: make([]string, 0),
			failed:     make(map[string]error),
		}

		// Process queue until empty or all requests have failed
		defaultErrorSleep := 100 * time.Millisecond
		ownerID := instances[0].OwnerID
	REQUESTLOOP:
		for len(queue) > 0 {
			select {
			case <-ctx.Done():
				// Context cancelled, mark remaining requests as failed
				for _, req := range queue {
					results.failed[req.instance.Name] = fmt.Errorf("operation cancelled: %w", ctx.Err())
				}
				queue = nil // Clear the queue
				break REQUESTLOOP
			default:
			}

			request := queue[0]
			queue = queue[1:] // pop from front

			// Skip if max attempts reached
			if request.attempts >= request.maxAttempts {
				results.failed[request.instance.Name] = fmt.Errorf("max attempts reached (%d): %w", request.maxAttempts, request.lastError)
				continue
			}

			request.attempts++

			// Try to delete infrastructure
			infra, err := NewInfrastructure(request.infraRequest)
			if err != nil {
				request.lastError = fmt.Errorf("failed to create infrastructure client: %w", err)
				queue = append(queue, request) // add back to queue
				time.Sleep(defaultErrorSleep)
				continue
			}

			_, err = infra.Execute()
			// If error exists and it's not a "not found" error, add back to queue
			if err != nil && (!strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found")) {
				request.lastError = fmt.Errorf("failed to delete infrastructure: %w", err)
				queue = append(queue, request) // add back to queue
				time.Sleep(defaultErrorSleep)
				continue
			}

			// Try to update database
			if err := s.repo.Terminate(ctx, ownerID, request.instance.ID); err != nil {
				request.lastError = fmt.Errorf("failed to terminate in database: %w", err)
				queue = append(queue, request) // add back to queue
				time.Sleep(defaultErrorSleep)
				continue
			}

			// Successfully deleted both infrastructure and database record
			results.successful = append(results.successful, request.instance.Name)
		}

		// Log final results
		if len(results.successful) > 0 {
			logger.Infof("Successfully deleted instances: %v", results.successful)
		}
		if len(results.failed) > 0 {
			logger.Infof("Failed to delete instances:")
			for name, err := range results.failed {
				logger.Infof("  %s: %v", name, err)
			}
		}

		// Create deletion result for API response
		deletionResult := map[string]interface{}{
			"status":    "completed",
			"deleted":   results.successful,
			"failed":    results.failed,
			"completed": time.Now().UTC(),
		}

		// Update final status with result
		if err := s.jobService.UpdateJobStatus(ctx, instances[0].JobID, models.JobStatusCompleted, deletionResult, ""); err != nil {
			logger.Errorf("Failed to update final job status: %v", err)
		}
	}()
}
