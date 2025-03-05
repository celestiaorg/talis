package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// JobServiceInterface defines the interface for job operations
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

// createJobRequest creates a new job request for instance operations
func (s *InstanceService) createJobRequest(
	name,
	projectName string,
	instances []infrastructure.InstanceRequest,
	action string,
) *infrastructure.JobRequest {
	return &infrastructure.JobRequest{
		Name:        name,
		ProjectName: projectName,
		Provider:    instances[0].Provider,
		Instances:   instances,
		Action:      action,
	}
}

// createJob creates a new job in the database
func (s *InstanceService) createJob(
	ctx context.Context,
	name,
	projectName,
	webhookURL string,
	status models.JobStatus,
) (*models.Job, error) {
	job := &models.Job{
		Name:        name,
		ProjectName: projectName,
		Status:      status,
		WebhookURL:  webhookURL,
	}

	return s.jobService.CreateJob(ctx, job)
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
	if err := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusInitializing, nil, ""); err != nil {
		return fmt.Errorf("failed to update job status to initializing: %w", err)
	}

	infra, err := infrastructure.NewInfrastructure(jobReq)
	if err != nil {
		return fmt.Errorf("failed to create infrastructure client: %w", err)
	}

	// Update to provisioning when creating infrastructure
	if err := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusProvisioning, nil, ""); err != nil {
		return fmt.Errorf("failed to update job status to provisioning: %w", err)
	}

	result, err := infra.Execute()
	if err != nil {
		// Check if error is due to resource not found
		if strings.Contains(err.Error(), "404") && strings.Contains(err.Error(), "could not be found") {
			fmt.Printf("⚠️ Warning: Some old resources were not found (already deleted)\n")
		} else {
			return fmt.Errorf("failed to execute infrastructure: %w", err)
		}
	}

	// Handle provisioning if requested
	if instances[0].Provision {
		if err := s.handleProvisioning(ctx, job, infra, result); err != nil {
			return err
		}
	}

	// Update final status with result
	if err := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusCompleted, result, ""); err != nil {
		return fmt.Errorf("failed to update final job status: %w", err)
	}

	fmt.Printf("✅ Infrastructure creation completed for job %s\n", job.ID)
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
	if err := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusConfiguring, instances, ""); err != nil {
		return fmt.Errorf("failed to update job status to configuring: %w", err)
	}

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
	fmt.Printf("🗑️ Starting async deletion of instance: %s\n", infraReq.Name)

	// Update to initializing when starting setup
	if err := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusInitializing, nil, ""); err != nil {
		return fmt.Errorf("failed to update job status to initializing: %w", err)
	}

	infra, err := infrastructure.NewInfrastructure(infraReq)
	if err != nil {
		return fmt.Errorf("failed to create infrastructure client: %w", err)
	}

	// Update to deleting status
	if err := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusDeleting, nil, ""); err != nil {
		return fmt.Errorf("failed to update job status to deleting: %w", err)
	}

	result, err := infra.Execute()
	if err != nil {
		// Check if the error is due to resources already being deleted
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			fmt.Printf("⚠️ Warning: Resources were already deleted\n")
			result = map[string]string{
				"status": "deleted",
				"note":   "resources were already deleted",
			}
			err = nil // Clear the error since this is an acceptable condition
		} else {
			return fmt.Errorf("failed to delete infrastructure: %w", err)
		}
	}

	// Update final status with result
	if err := s.jobService.UpdateJobStatus(ctx, job.ID, models.JobStatusCompleted, result, ""); err != nil {
		return fmt.Errorf("failed to update final job status: %w", err)
	}

	fmt.Printf("✅ Infrastructure deletion completed for job %d\n", job.ID)
	return nil
}

// CreateInstance creates a new instance
func (s *InstanceService) CreateInstance(
	ctx context.Context,
	name,
	projectName,
	webhookURL string,
	instances []infrastructure.InstanceRequest,
) (*models.Job, error) {
	// Create and validate job request
	jobReq := s.createJobRequest(name, projectName, instances, "create")
	if err := jobReq.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create job first
	job, err := s.createJob(ctx, name, projectName, webhookURL, models.JobStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Start async infrastructure creation
	go func() {
		if err := s.handleInfrastructureCreation(context.Background(), job, jobReq, instances); err != nil {
			fmt.Printf("❌ Failed to handle infrastructure creation: %v\n", err)
			s.jobService.UpdateJobStatus(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error())
		}
	}()

	return job, nil
}

// DeleteInstance deletes an instance
func (s *InstanceService) DeleteInstance(
	ctx context.Context,
	instanceID string,
) (*models.Job, error) {
	// Create job for deletion
	job, err := s.createJob(ctx, fmt.Sprintf("delete-%s", instanceID), "talis", "", models.JobStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Create infrastructure request
	infraReq := s.createJobRequest(instanceID, "talis", []infrastructure.InstanceRequest{
		{
			Provider:          "digitalocean",
			NumberOfInstances: 1,
			Region:            "nyc3",
		},
	}, "delete")

	// Start async infrastructure deletion
	go func() {
		if err := s.handleInfrastructureDeletion(context.Background(), job, infraReq); err != nil {
			fmt.Printf("❌ Failed to handle infrastructure deletion: %v\n", err)
			s.jobService.UpdateJobStatus(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error())
		}
	}()

	return job, nil
}

// GetInstance retrieves an instance by ID
func (s *InstanceService) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, id)
}
