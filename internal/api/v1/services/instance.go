package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// InstanceService provides business logic for instance operations
type InstanceService struct {
	repo       *repos.InstanceRepository
	jobService *JobService
}

// NewInstanceService creates a new instance service instance
func NewInstanceService(repo *repos.InstanceRepository, jobService *JobService) *InstanceService {
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
func (s *InstanceService) CreateInstance(ctx context.Context, ownerID uint, jobName string, instances []infrastructure.InstanceRequest) error {
	job, err := s.jobService.jobRepo.GetByName(ctx, ownerID, jobName)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	for _, i := range instances {
		if i.Name == "" {
			i.Name = fmt.Sprintf("instance-%s", uuid.New().String())
		}

		instance := &models.Instance{
			Name:       i.Name,
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

	s.provisionInstances(ctx, job.ID, instances)

	return nil
}

// GetInstance retrieves an instance by ID
func (s *InstanceService) GetInstance(ctx context.Context, id uint) (*models.Instance, error) {
	return s.repo.Get(ctx, id)
}

// provisionInstances provisions the job asynchronously
func (s *InstanceService) provisionInstances(ctx context.Context, jobID uint, instances []infrastructure.InstanceRequest) {
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

		// Start Ansible provisioning if creation was successful and provisioning is requested
		if instances[0].Provision {
			pInstances, ok := result.([]infrastructure.InstanceInfo)
			if !ok {
				fmt.Printf("❌ Invalid result type: %T\n", result)
				return
			}

			fmt.Printf("📝 Created instances: %+v\n", pInstances)

			for _, instance := range pInstances {
				err := s.repo.UpdateIPByName(ctx, instance.Name, instance.IP)
				if err != nil {
					fmt.Printf("❌ Failed to update instance %s IP: %v\n", instance.Name, err)
				}

				err = s.repo.UpdateStatusByName(ctx, instance.Name, models.InstanceStatusProvisioning)
				if err != nil {
					fmt.Printf("❌ Failed to update instance %s status to provisioning: %v\n", instance.Name, err)
				}
			}

			// TODO: instance provisioning should be done in a way that is async and updates the instance status one by one in the db
			if err := infra.RunProvisioning(pInstances); err != nil {
				fmt.Printf("❌ Failed to run provisioning: %v\n", err)
				return
			}

			for _, instance := range pInstances {
				// TODO:Consider passing OwnerID to update the status as well
				// Not sure if Ready is the best status to set here
				if err := s.repo.UpdateStatusByName(ctx, instance.Name, models.InstanceStatusReady); err != nil {
					fmt.Printf("❌ Failed to update instance status to provisioning: %v\n", err)
				}
			}
		}

		fmt.Printf("✅ Infrastructure creation completed for job ID %d\n", jobID)
	}()
}
