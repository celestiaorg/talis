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
	jobRepo      *repos.JobRepository
	instanceRepo *repos.InstanceRepository
}

// NewJobService creates a new job service instance
func NewJobService(jobRepo *repos.JobRepository, instanceRepo *repos.InstanceRepository) *JobService {
	return &JobService{jobRepo: jobRepo, instanceRepo: instanceRepo}
}

// ListJobs retrieves a paginated list of jobs
func (s *JobService) ListJobs(ctx context.Context, status models.JobStatus, ownerID uint, opts *models.ListOptions) ([]models.Job, error) {
	return s.jobRepo.List(ctx, status, ownerID, opts)
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

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, err
	}

	s.provisionJob(ctx, job.ID, jobReq)
	return job, nil
}

// GetJobStatus retrieves the status of a job
func (s *JobService) GetJobStatus(ctx context.Context, ownerID uint, id uint) (models.JobStatus, error) {
	j, err := s.jobRepo.GetByID(ctx, ownerID, id)
	if err != nil {
		return models.JobStatusUnknown, err
	}
	return j.Status, nil
}

// UpdateJobStatus updates the status of a job
func (s *JobService) UpdateJobStatus(ctx context.Context, id uint, status models.JobStatus, result interface{}, errMsg string) error {
	return s.jobRepo.UpdateStatus(ctx, id, status, result, errMsg)
}

func (s *JobService) TerminateJob(ctx context.Context, ownerID uint, jobID uint) error {
	job, err := s.jobRepo.GetByID(ctx, ownerID, jobID)
	if err != nil {
		return err
	}

	s.terminateJob(ctx, job)
	return nil
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

		if jobReq.Provider == "" {
			jobReq.Provider = jobReq.Instances[0].Provider
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
				fmt.Printf("⚠️ Warning: Some old resources were not found (already deleted)\n")
			} else {
				fmt.Printf("❌ Failed to execute infrastructure: %v\n", err)
				if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed, nil, err.Error()); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				return
			}
		}

		// Start Ansible provisioning if creation was successful and provisioning is requested
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
			for _, instance := range instances {
				err := s.instanceRepo.Create(ctx, &models.Instance{
					JobID:      jobID,
					Name:       instance.Name,
					ProviderID: instance.Provider,
					Status:     models.InstanceStatusProvisioning,
					Region:     instance.Region,
					Size:       instance.Size,
				})
				if err != nil {
					fmt.Printf("❌ Failed to create instance: %v\n", err)
				}
			}

			// Update to configuring when setting up Ansible
			if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusConfiguring, instances, ""); err != nil {
				fmt.Printf("❌ Failed to update job status to configuring: %v\n", err)
				return
			}

			// TODO: instance provisioning should be done in a way that is async and updates the instance status one by one in the db
			if err := infra.RunProvisioning(instances); err != nil {
				fmt.Printf("❌ Failed to run provisioning: %v\n", err)
				if err := s.UpdateJobStatus(ctx, jobID, models.JobStatusFailed, instances, fmt.Sprintf("infrastructure created but provisioning failed: %v", err)); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				return
			}

			for _, instance := range instances {
				// TODO:Consider passing JobID and OwnerID to update the status as well
				// Not sure if Ready is the best status to set here
				if err := s.instanceRepo.UpdateStatusByName(ctx, instance.Name, models.InstanceStatusReady); err != nil {
					fmt.Printf("❌ Failed to update instance status to provisioning: %v\n", err)
				}
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

// GetByProjectName retrieves a job by its project name
func (s *JobService) GetByProjectName(ctx context.Context, projectName string) (*models.Job, error) {
	return s.jobRepo.GetByProjectName(ctx, projectName)
}

// handleInfrastructureDeletion handles the infrastructure deletion process
func (s *JobService) terminateJob(ctx context.Context, job *models.Job) {
	go func() {
		fmt.Printf("🗑️ Starting async deletion for job %d\n", job.ID)

		// Get instances from database for this job, ordered by creation time
		instances, err := s.instanceRepo.GetByJobIDOrdered(ctx, job.ID)
		if err != nil {
			fmt.Printf("❌ Failed to get instances: %v\n", err)
			return
		}

		if len(instances) == 0 {
			fmt.Printf("❌ No instances found for job %d\n", job.ID)
			return
		}

		// Prepare deletion result
		deletionResult := map[string]interface{}{
			"status":  "deleting",
			"deleted": []string{},
		}

		// Try to delete each selected instance
		for _, instance := range instances {
			fmt.Printf("🗑️ Attempting to delete instance: %s\n", instance.Name)

			// Create a new infrastructure request for each instance
			instanceInfraReq := &infrastructure.JobRequest{
				Name:        instance.Name,
				ProjectName: job.ProjectName,
				Provider:    instance.ProviderID,
				Instances: []infrastructure.InstanceRequest{{
					Provider: instance.ProviderID,
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
					if err := s.instanceRepo.Terminate(ctx, instance.ID); err != nil {
						fmt.Printf("❌ Failed to mark instance %s as terminated in database: %v\n", instance.Name, err)
					} else {
						fmt.Printf("✅ Marked instance %s as terminated in database\n", instance.Name)
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
				if err := s.instanceRepo.Terminate(ctx, instance.ID); err != nil {
					fmt.Printf("❌ Failed to mark instance %s as terminated in database: %v\n", instance.Name, err)
				} else {
					fmt.Printf("✅ Marked instance %s as terminated in database\n", instance.Name)
					if deleted, ok := deletionResult["deleted"].([]string); ok {
						deletionResult["deleted"] = append(deleted, instance.Name)
					}
				}
			}
		}

		deletionResult["status"] = "completed"

		// Update final status with result
		s.UpdateJobStatus(ctx, job.ID, models.JobStatusCompleted, deletionResult, "")

		fmt.Printf("✅ Infrastructure deletion completed for job %d\n", job.ID)
	}()
}
