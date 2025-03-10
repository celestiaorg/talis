package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// JobServiceInterface defines the interface for job operations.
// This interface is used by InstanceService to manage job states during infrastructure operations.
// It provides a minimal set of methods needed to create and update job statuses,
// following the Interface Segregation Principle by only exposing the methods that are actually used.
type JobServiceInterface interface {
	CreateJob(ctx context.Context, job *models.Job) (*models.Job, error)
	UpdateJobStatus(ctx context.Context, id uint, status models.JobStatus, result interface{}, errMsg string) error
}

// InstanceService provides business logic for instance operations
type InstanceService struct {
	repo       *repos.InstanceRepository
	jobService JobServiceInterface
}

// NewInstanceService creates a new instance service instance
func NewInstanceService(repo *repos.InstanceRepository, jobService JobServiceInterface) *InstanceService {
	return &InstanceService{
		repo:       repo,
		jobService: jobService,
	}
}

// ListInstances retrieves a paginated list of instances
func (s *InstanceService) ListInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	return s.repo.List(ctx, opts)
}

// CreateInstance creates a new instance
func (s *InstanceService) CreateInstance(
	ctx context.Context,
	instanceName,
	projectName,
	webhookURL string,
	instances []infrastructure.InstanceRequest,
) (*models.Job, error) {
	// Create job request
	jobReq := &infrastructure.JobRequest{
		InstanceName: instanceName,
		ProjectName:  projectName,
		Provider:     instances[0].Provider,
		Instances:    instances,
		Action:       "create",
	}

	if err := jobReq.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create job first
	job := &models.Job{
		Name:        instanceName,
		ProjectName: projectName,
		Status:      models.JobStatusPending,
		WebhookURL:  webhookURL,
	}

	job, err := s.jobService.CreateJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Start async infrastructure creation
	go func() {
		if err := s.handleInfrastructureCreation(context.Background(), job, jobReq, instances); err != nil {
			fmt.Printf("❌ Failed to handle infrastructure creation: %v\n", err)
			s.updateJobStatusWithError(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error())
		}
	}()

	return job, nil
}

// DeleteInstance deletes an instance
func (s *InstanceService) DeleteInstance(
	ctx context.Context,
	jobID uint,
	instanceName,
	projectName string,
	instances []infrastructure.InstanceRequest,
) (*models.Job, error) {
	// Create job for deletion
	job := &models.Job{
		Model: gorm.Model{
			ID: jobID,
		},
		Name:        instanceName,
		ProjectName: projectName,
		Status:      models.JobStatusPending,
	}

	// Create infrastructure request
	infraReq := &infrastructure.JobRequest{
		InstanceName: instanceName,
		ProjectName:  projectName,
		Provider:     instances[0].Provider,
		Instances:    instances,
		Action:       "delete",
	}

	// Start async infrastructure deletion
	go func() {
		if err := s.handleInfrastructureDeletion(context.Background(), job, infraReq); err != nil {
			fmt.Printf("❌ Failed to handle infrastructure deletion: %v\n", err)
			s.updateJobStatusWithError(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error())
		}
	}()

	return job, nil
}

// GetInstance retrieves an instance by ID
func (s *InstanceService) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, id)
}

// handleInfrastructureCreation handles the infrastructure creation process
func (s *InstanceService) handleInfrastructureCreation(
	ctx context.Context,
	job *models.Job,
	jobReq *infrastructure.JobRequest,
	instances []infrastructure.InstanceRequest,
) error {
	fmt.Println("🚀 Starting async infrastructure creation...")

	// Update to initializing when starting setup
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusInitializing, nil, "")

	infra, err := infrastructure.NewInfrastructure(jobReq)
	if err != nil {
		return fmt.Errorf("failed to create infrastructure client: %w", err)
	}

	// Update to provisioning when creating infrastructure
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusProvisioning, nil, "")

	result, err := infra.Execute()
	if err != nil {
		// Check if error is due to resource not found
		if strings.Contains(err.Error(), "404") && strings.Contains(err.Error(), "could not be found") {
			fmt.Printf("⚠️ Warning: Some old resources were not found (already deleted)\n")
		} else {
			return fmt.Errorf("failed to execute infrastructure: %w", err)
		}
	}

	// Save instances to database
	instanceInfos, ok := result.([]infrastructure.InstanceInfo)
	if !ok {
		return fmt.Errorf("invalid result type: expected []infrastructure.InstanceInfo, got %T", result)
	}

	// Get the request configuration that applies to all instances
	instanceConfig := instances[0] // Safe because we validate there's at least one instance in the request

	for _, info := range instanceInfos {
		// Create default tags if none provided
		var tags []string
		if len(instanceConfig.Tags) > 0 {
			tags = instanceConfig.Tags
		} else {
			tags = []string{fmt.Sprintf("%s-do-instance", job.Name)}
		}

		// Create instance in database
		instance := &models.Instance{
			JobID:      job.ID,
			ProviderID: models.ProviderID(jobReq.Provider),
			Name:       info.Name,
			PublicIP:   info.IP,
			Region:     info.Region,
			Size:       info.Size,
			Image:      instanceConfig.Image,
			Tags:       tags,
			Status:     models.InstanceStatusReady,
		}
		if err := s.repo.Create(ctx, instance); err != nil {
			fmt.Printf("❌ Failed to save instance to database: %v\n", err)
		} else {
			fmt.Printf("✅ Instance %s saved to database with ID %d\n", instance.Name, instance.ID)
		}
	}

	// Handle provisioning if requested
	if instances[0].Provision {
		if err := s.handleProvisioning(ctx, job, infra, instanceInfos); err != nil {
			return err
		}
	}

	// Update final status with result
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusCompleted, instanceInfos, "")

	fmt.Printf("✅ Infrastructure creation completed for job ID %d\n", job.ID)
	return nil
}

// handleProvisioning handles the provisioning process for instances
func (s *InstanceService) handleProvisioning(
	ctx context.Context,
	job *models.Job,
	infra *infrastructure.Infrastructure,
	result interface{},
) error {
	instances, ok := result.([]infrastructure.InstanceInfo)
	if !ok {
		return fmt.Errorf("invalid result type: %T, expected []infrastructure.InstanceInfo", result)
	}

	fmt.Printf("📝 Created instances: %+v\n", instances)

	// Update to configuring when setting up Ansible
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusConfiguring, instances, "")

	if err := infra.RunProvisioning(instances); err != nil {
		return fmt.Errorf("infrastructure created but provisioning failed: %w", err)
	}

	return nil
}

// handleInfrastructureDeletion handles the infrastructure deletion process
func (s *InstanceService) handleInfrastructureDeletion(
	ctx context.Context,
	job *models.Job,
	infraReq *infrastructure.JobRequest,
) error {
	fmt.Printf("🗑️ Starting async deletion for job %d\n", job.ID)

	// Update to initializing when starting setup
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusInitializing, nil, "")

	// Get instances from database for this job, ordered by creation time
	instances, err := s.repo.GetByJobIDOrdered(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("failed to get instances: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("no instances found for job %d", job.ID)
	}

	// Check if we have specific instance names to delete
	var instancesToDelete []models.Instance
	var specificNamesToDelete []string

	// Collect specific instance names to delete
	for _, instanceReq := range infraReq.Instances {
		if instanceReq.Name != "" {
			specificNamesToDelete = append(specificNamesToDelete, instanceReq.Name)
		}
	}

	// If specific names are provided, find those instances
	if len(specificNamesToDelete) > 0 {
		fmt.Printf("ℹ️ Deleting specific instances by name: %v\n", specificNamesToDelete)
		
		// Find instances with the specified names
		for _, name := range specificNamesToDelete {
			found := false
			for _, instance := range instances {
				if instance.Name == name {
					instancesToDelete = append(instancesToDelete, instance)
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("⚠️ Warning: Instance with name '%s' not found\n", name)
			}
		}
		
		if len(instancesToDelete) == 0 {
			return fmt.Errorf("none of the specified instances were found")
		}
	} else {
		// No specific names provided, use FIFO order (oldest first)
		// Calculate how many instances to delete
		numberOfInstancesToDelete := 0
		for _, instanceReq := range infraReq.Instances {
			numberOfInstancesToDelete += instanceReq.NumberOfInstances
		}

		if numberOfInstancesToDelete <= 0 {
			return fmt.Errorf("invalid number of instances to delete: %d", numberOfInstancesToDelete)
		}

		// Limit the number of instances to delete to the available ones
		if numberOfInstancesToDelete > len(instances) {
			fmt.Printf("⚠️ Warning: Requested to delete %d instances but only %d are available\n",
				numberOfInstancesToDelete, len(instances))
			numberOfInstancesToDelete = len(instances)
		}

		// Select only the oldest instances to delete
		instancesToDelete = instances[:numberOfInstancesToDelete]
		
		fmt.Printf("ℹ️ Will delete %d oldest instances out of %d total instances\n",
			numberOfInstancesToDelete, len(instances))
	}

	// Prepare deletion result
	deletionResult := map[string]interface{}{
		"status":  "deleting",
		"deleted": []string{},
	}

	// Try to delete each selected instance
	for _, instance := range instancesToDelete {
		fmt.Printf("🗑️ Attempting to delete instance: %s\n", instance.Name)

		// Create a new infrastructure request for each instance
		instanceInfraReq := &infrastructure.JobRequest{
			InstanceName: instance.Name,
			ProjectName:  infraReq.ProjectName,
			Provider:     infraReq.Provider,
			Instances: []infrastructure.InstanceRequest{{
				Provider: string(instance.ProviderID),
				Region:   instance.Region,
				Size:     instance.Size,
			}},
			Action: "delete",
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
				if err := s.repo.Delete(ctx, instance.ID); err != nil {
					fmt.Printf("❌ Failed to mark instance %s as deleted in database: %v\n", instance.Name, err)
				} else {
					fmt.Printf("✅ Marked instance %s as deleted in database\n", instance.Name)
					if deleted, ok := deletionResult["deleted"].([]string); ok {
						deletionResult["deleted"] = append(deleted, instance.Name)
					}
				}
			} else {
				fmt.Printf("❌ Error deleting instance %s: %v\n", instance.Name, err)
				continue
			}
		} else {
			// Deletion was successful, update database
			if err := s.repo.Delete(ctx, instance.ID); err != nil {
				fmt.Printf("❌ Failed to mark instance %s as deleted in database: %v\n", instance.Name, err)
			} else {
				fmt.Printf("✅ Marked instance %s as deleted in database\n", instance.Name)
				if deleted, ok := deletionResult["deleted"].([]string); ok {
					deletionResult["deleted"] = append(deleted, instance.Name)
				}
			}
		}
	}

	deletionResult["status"] = "completed"

	// Update final status with result
	s.updateJobStatusWithError(ctx, job.ID, models.JobStatusCompleted, deletionResult, "")

	fmt.Printf("✅ Infrastructure deletion completed for job %d\n", job.ID)
	return nil
}

// Update job status with error handling
func (s *InstanceService) updateJobStatusWithError(
	ctx context.Context,
	jobID uint,
	status models.JobStatus,
	result interface{},
	errMsg string,
) {
	if err := s.jobService.UpdateJobStatus(ctx, jobID, status, result, errMsg); err != nil {
		log.Printf("Failed to update job status: %v", err)
	}
}
