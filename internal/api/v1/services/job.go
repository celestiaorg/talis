package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// JobService provides business logic for job operations
type JobService struct {
	repo *repos.JobRepository
}

// NewJobService creates a new job service instance
func NewJobService(repo *repos.JobRepository) *JobService {
	return &JobService{repo: repo}
}

// ListJobs retrieves a paginated list of jobs
func (s *JobService) ListJobs(ctx context.Context, status models.JobStatus, ownerID uint, opts *models.ListOptions) ([]models.Job, error) {
	return s.repo.List(ctx, status, ownerID, opts)
}

// CreateJob creates a new job
func (s *JobService) CreateJob(ctx context.Context, ownerID uint, jobReq *infrastructure.JobRequest) (*models.Job, error) {
	job := &models.Job{
		Name:        jobReq.Name,
		OwnerID:     uint(ownerID),
		ProjectName: jobReq.ProjectName,
		Status:      models.JobStatusPending,
		WebhookURL:  jobReq.WebhookURL,
	}

	if job.Name == "" {
		job.Name = fmt.Sprintf("job-%s", time.Now().Format("20060102-150405"))
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	s.provisionJob(ctx, job.ID, jobReq)
	return job, nil
}

// GetJobStatus retrieves the status of a job
func (s *JobService) GetJobStatus(ctx context.Context, ownerID uint, id uint) (models.JobStatus, error) {
	j, err := s.repo.GetByID(ctx, ownerID, id)
	if err != nil {
		return models.JobStatusUnknown, err
	}
	return j.Status, nil
}

// UpdateJobStatus updates the status of a job
func (s *JobService) UpdateJobStatus(ctx context.Context, id uint, status models.JobStatus, result interface{}, errMsg string) error {
	return s.repo.UpdateStatus(ctx, id, status, result, errMsg)
}

// provisionJob provisions the job asynchronously
func (s *JobService) provisionJob(ctx context.Context, jobID uint, jobReq *infrastructure.JobRequest) {
	// TODO: do something with the logs as now it makes the server logs messy
	go func() {
		fmt.Println("🚀 Starting async infrastructure creation...")

		// Update to initializing when starting Pulumi setup
		if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusInitializing, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
			return
		}

		infra, err := infrastructure.NewInfrastructure(jobReq)
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure: %v\n", err)
			if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed, nil, err.Error()); err != nil {
				log.Printf("Failed to update job status: %v", err)
			}
			return
		}

		// Update to provisioning when creating infrastructure
		if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusProvisioning, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to provisioning: %v\n", err)
			return
		}

		result, err := infra.Execute()
		if err != nil {
			// Check if error is due to resource not found
			if strings.Contains(err.Error(), "404") &&
				strings.Contains(err.Error(), "could not be found") {
				// Get Pulumi output result
				outputs, outputErr := infra.GetOutputs()
				if outputErr != nil {
					fmt.Printf("❌ Failed to get outputs: %v\n", outputErr)
					if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed, nil, outputErr.Error()); err != nil {
						log.Printf("Failed to update job status: %v", err)
					}
					return
				}
				result = outputs
				fmt.Printf("⚠️ Warning: Some old resources were not found (already deleted)\n")
			} else {
				fmt.Printf("❌ Failed to execute infrastructure: %v\n", err)
				if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed, nil, err.Error()); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				return
			}
		}

		// Start Nix provisioning if creation was successful and provisioning is requested
		if jobReq.Instances[0].Provision {
			instances, ok := result.([]infrastructure.InstanceInfo)
			if !ok {
				fmt.Printf("❌ Invalid result type: %T\n", result)
				if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed, nil,
					fmt.Sprintf("invalid result type: %T, expected []infrastructure.InstanceInfo", result)); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				return
			}

			fmt.Printf("📝 Created instances: %+v\n", instances)

			// Update to configuring when setting up Nix
			if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusConfiguring, instances, ""); err != nil {
				fmt.Printf("❌ Failed to update job status to configuring: %v\n", err)
				return
			}

			if err := infra.RunProvisioning(instances); err != nil {
				fmt.Printf("❌ Failed to run provisioning: %v\n", err)
				if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed, instances, fmt.Sprintf("infrastructure created but provisioning failed: %v", err)); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				return
			}
		}

		// Update final status with result
		if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusCompleted, result, ""); err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure creation completed for job ID %d and job name %s\n", jobID, jobReq.Name)
	}()
}
