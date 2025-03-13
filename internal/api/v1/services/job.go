package services

import (
	"context"
	"fmt"
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
		OwnerID:     ownerID,
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

// TerminateJob terminates a job and all its instances
func (s *JobService) TerminateJob(ctx context.Context, ownerID uint, jobID uint) error {
	job, err := s.jobRepo.GetByID(ctx, ownerID, jobID)
	if err != nil {
		return err
	}

	s.terminateJob(ctx, job)
	return nil
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
		err = s.UpdateJobStatus(ctx, job.ID, models.JobStatusCompleted, deletionResult, "")
		if err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure deletion completed for job %d\n", job.ID)
	}()
}
