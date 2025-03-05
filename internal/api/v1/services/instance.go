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

// CreateInstance creates a new instance
func (s *InstanceService) CreateInstance(
	ctx context.Context,
	name,
	projectName,
	webhookURL string,
	instances []infrastructure.InstanceRequest,
) (*models.Job, error) {
	// Convert to Request and validate
	jobReq := &infrastructure.JobRequest{
		Name:        name,
		ProjectName: projectName,
		Provider:    instances[0].Provider,
		Instances:   instances,
		Action:      "create",
	}

	if err := jobReq.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create job first
	job := &models.Job{
		Name:        name,
		ProjectName: projectName,
		Status:      models.JobStatusPending,
		WebhookURL:  webhookURL,
	}

	j, err := s.jobService.CreateJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Start async infrastructure creation
	go func() {
		fmt.Println("🚀 Starting async infrastructure creation...")

		// Update to initializing when starting setup
		if err := s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusInitializing, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
			return
		}

		infra, err := infrastructure.NewInfrastructure(jobReq)
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure client: %v\n", err)
			s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
			return
		}

		// Update to provisioning when creating infrastructure
		if err := s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusProvisioning, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to provisioning: %v\n", err)
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
				s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
				return
			}
		}

		// Start Ansible provisioning if creation was successful and provisioning is requested
		if instances[0].Provision {
			instances, ok := result.([]infrastructure.InstanceInfo)
			if !ok {
				fmt.Printf("❌ Invalid result type: %T\n", result)
				s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil,
					fmt.Sprintf("invalid result type: %T, expected []infrastructure.InstanceInfo", result))
				return
			}

			fmt.Printf("📝 Created instances: %+v\n", instances)

			// Update to configuring when setting up Ansible
			if err := s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusConfiguring, instances, ""); err != nil {
				fmt.Printf("❌ Failed to update job status to configuring: %v\n", err)
				return
			}

			if err := infra.RunProvisioning(instances); err != nil {
				fmt.Printf("❌ Failed to run provisioning: %v\n", err)
				s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, instances, fmt.Sprintf("infrastructure created but provisioning failed: %v", err))
				return
			}
		}

		// Update final status with result
		if err := s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusCompleted, result, ""); err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure creation completed for job %s\n", j.ID)
	}()

	return j, nil
}

// DeleteInstance deletes an instance
func (s *InstanceService) DeleteInstance(ctx context.Context, instanceID string) (*models.Job, error) {
	// Create job for deletion
	job := &models.Job{
		Name:        fmt.Sprintf("delete-%s", instanceID),
		ProjectName: "talis", // Default project for now
		Status:      models.JobStatusPending,
	}

	j, err := s.jobService.CreateJob(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Create infrastructure request
	infraReq := &infrastructure.JobRequest{
		Name:        instanceID,
		ProjectName: "talis",
		Provider:    "digitalocean",
		Action:      "delete",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 1,
				Region:            "nyc3",
			},
		},
	}

	// Start async infrastructure deletion
	go func() {
		fmt.Printf("🗑️ Starting async deletion of instance: %s\n", instanceID)

		// Update to initializing when starting setup
		if err := s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusInitializing, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
			return
		}

		infra, err := infrastructure.NewInfrastructure(infraReq)
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure client: %v\n", err)
			s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
			return
		}

		// Update to deleting status
		if err := s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusDeleting, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to deleting: %v\n", err)
			return
		}

		result, err := infra.Execute()
		if err != nil {
			// Check if the error is due to resources already being deleted
			if strings.Contains(err.Error(), "404") ||
				strings.Contains(err.Error(), "not found") {
				fmt.Printf("⚠️ Warning: Resources were already deleted\n")
				result = map[string]string{
					"status": "deleted",
					"note":   "resources were already deleted",
				}
				err = nil // Clear the error since this is an acceptable condition
			} else {
				fmt.Printf("❌ Failed to delete infrastructure: %v\n", err)
				s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusFailed, nil, err.Error())
				return
			}
		}

		// Update final status with result
		if err := s.jobService.UpdateJobStatus(context.Background(), j.ID, models.JobStatusCompleted, result, ""); err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure deletion completed for job %d\n", j.ID)
	}()

	return j, nil
}

// GetInstance retrieves an instance by ID
func (s *InstanceService) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, id)
}
