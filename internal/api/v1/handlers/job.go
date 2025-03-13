package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// JobHandler handles HTTP requests for job operations
type JobHandler struct {
	service *services.JobService
}

// NewJobHandler creates a new job handler instance
func NewJobHandler(s *services.JobService) *JobHandler {
	return &JobHandler{service: s}
}

// GetJobStatus handles the request to get a job's status
func (h *JobHandler) GetJobStatus(c *fiber.Ctx) error {
	jobIDStr := c.Params("id")
	if jobIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "job id is required",
		})
	}

	ownerID := 0 // TODO: get owner id from the JWT token
	jobID, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid job id",
		})
	}

	status, err := h.service.GetJobStatus(c.Context(), uint(ownerID), uint(jobID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get job status: %v", err),
		})
	}

	return c.JSON(status)
}

// ListJobs handles the request to list jobs
func (h *JobHandler) ListJobs(c *fiber.Ctx) error {
	var (
		limit  = c.QueryInt("limit", 10)
		offset = c.QueryInt("offset", 0)
	)
	status, err := models.ParseJobStatus(c.Query("status"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid job status",
		})
	}

	ownerID := 0 // TODO: get owner id from the JWT token

	jobs, err := h.service.ListJobs(c.Context(), status, uint(ownerID), &models.ListOptions{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list jobs: %v", err),
		})
	}

	return c.JSON(jobs)
}

// CreateJob handles the request to create a new job
func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	var req infrastructure.CreateJobRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to Request and validate
	JobReq := &infrastructure.JobRequest{
		JobName:      req.JobName,
		InstanceName: req.InstanceName,
		ProjectName:  req.ProjectName,
		Provider:     req.Instances[0].Provider,
		Instances:    req.Instances,
		Action:       "create",
	}

	if err := JobReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ownerID := 0 // TODO: get owner id from the JWT token

	job, err := h.service.CreateJob(c.Context(), &models.Job{
		Name:         req.JobName,
		InstanceName: req.InstanceName,
		OwnerID:      uint(ownerID),
		ProjectName:  req.ProjectName,
		Status:       models.JobStatusPending,
		WebhookURL:   req.WebhookURL,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to create job: %v", err),
		})
	}

	// TODO: this logic should be moved to a service
	// Create job first
	go func() {
		fmt.Println("🚀 Starting async infrastructure creation...")

		// Update to initializing when starting Pulumi setup
		if err := h.service.UpdateJobStatus(context.Background(), job.ID, models.JobStatusInitializing, nil, ""); err != nil {
			fmt.Printf("❌ Failed to update job status to initializing: %v\n", err)
			return
		}

		infra, err := infrastructure.NewInfrastructure(JobReq)
		if err != nil {
			fmt.Printf("❌ Failed to create infrastructure: %v\n", err)
			if err := h.service.UpdateJobStatus(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error()); err != nil {
				log.Printf("Failed to update job status: %v", err)
			}
			return
		}

		// Update to provisioning when creating infrastructure
		if err := h.service.UpdateJobStatus(context.Background(), job.ID, models.JobStatusProvisioning, nil, ""); err != nil {
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
				if err := h.service.UpdateJobStatus(context.Background(), job.ID, models.JobStatusFailed, nil, err.Error()); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				return
			}
		}

		// Start Ansible provisioning if creation was successful and provisioning is requested
		if req.Instances[0].Provision {
			instances, ok := result.([]infrastructure.InstanceInfo)
			if !ok {
				fmt.Printf("❌ Invalid result type: %T\n", result)
				if err := h.service.UpdateJobStatus(context.Background(), job.ID, models.JobStatusFailed, nil,
					fmt.Sprintf("invalid result type: %T, expected []infrastructure.InstanceInfo", result)); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				return
			}

			fmt.Printf("📝 Created instances: %+v\n", instances)

			// Update to configuring when setting up Ansible
			if err := h.service.UpdateJobStatus(context.Background(), job.ID, models.JobStatusConfiguring, instances, ""); err != nil {
				fmt.Printf("❌ Failed to update job status to configuring: %v\n", err)
				return
			}

			if err := infra.RunProvisioning(instances); err != nil {
				fmt.Printf("❌ Failed to run provisioning: %v\n", err)
				if err := h.service.UpdateJobStatus(context.Background(), job.ID, models.JobStatusFailed, instances, fmt.Sprintf("infrastructure created but provisioning failed: %v", err)); err != nil {
					log.Printf("Failed to update job status: %v", err)
				}
				return
			}
		}

		// Update final status with result
		if err := h.service.UpdateJobStatus(context.Background(), job.ID, models.JobStatusCompleted, result, ""); err != nil {
			fmt.Printf("❌ Failed to update final job status: %v\n", err)
			return
		}

		fmt.Printf("✅ Infrastructure creation completed for job ID %d and job name %s\n", job.ID, job.Name)
	}()

	return c.Status(fiber.StatusCreated).JSON(job)
}
