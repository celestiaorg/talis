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

		// Get the job name from the database
		job, err := s.jobService.jobRepo.GetByID(ctx, 0, jobID) // ownerID 0 for now since we're not using it yet
		if err != nil {
			fmt.Printf("❌ Failed to get job details: %v\n", err)
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
			// Create update instance with only the fields we want to update
			updateInstance := &models.Instance{
				PublicIP: instance.IP,
				Status:   models.InstanceStatusReady,
			}
			if err := s.repo.UpdateByName(ctx, instance.Name, updateInstance); err != nil {
				fmt.Printf("❌ Failed to update instance %s: %v\n", instance.Name, err)
				continue
			}
			fmt.Printf("✅ Updated instance %s with IP %s and status ready\n", instance.Name, instance.IP)
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

// deleteRequest represents a single instance deletion request with tracking
type deleteRequest struct {
	instance     models.Instance
	infraRequest *infrastructure.InstancesRequest
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
			fmt.Printf("❌ No instances provided to terminate\n")
			return
		}

		// Create queue of delete requests
		queue := make([]*deleteRequest, 0, len(instances))
		for _, instance := range instances {
			infraReq := &infrastructure.InstancesRequest{
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
			infra, err := infrastructure.NewInfrastructure(request.infraRequest)
			if err != nil {
				request.lastError = fmt.Errorf("failed to create infrastructure client: %w", err)
				queue = append(queue, request) // add back to queue
				time.Sleep(defaultErrorSleep)
				continue
			}

			_, err = infra.Execute()
			// If error exists and it's not a "not found" error, add back to queue
			if err != nil && !(strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found")) {
				request.lastError = fmt.Errorf("failed to delete infrastructure: %w", err)
				queue = append(queue, request) // add back to queue
				time.Sleep(defaultErrorSleep)
				continue
			}

			// Try to update database
			if err := s.repo.Terminate(ctx, request.instance.ID); err != nil {
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

		// Update deletion result for API response
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
