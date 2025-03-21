package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// JobHandler handles HTTP requests for job operations
type JobHandler struct {
	service         services.Job
	instanceService services.Instance
}

// NewJobHandler creates a new job handler instance
func NewJobHandler(s services.Job, instanceService services.Instance) *JobHandler {
	return &JobHandler{
		service:         s,
		instanceService: instanceService,
	}
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
	var status = models.JobStatusUnknown

	// Parse status if provided
	if statusStr := c.Query("status"); statusStr != "" {
		var err error
		status, err = models.ParseJobStatus(statusStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid job status",
			})
		}
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", DefaultPageSize)
	paginationOpts := getPaginationOptions(page, limit)

	ownerID := 0 // TODO: get owner id from the JWT token

	jobs, err := h.service.ListJobs(c.Context(), status, uint(ownerID), paginationOpts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list jobs: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"jobs":   jobs,
		"total":  len(jobs),
		"page":   page,
		"limit":  limit,
		"offset": paginationOpts.Offset,
	})
}

// CreateJob handles the request to create a new job
func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	var req infrastructure.CreateJobRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if project name already exists
	existingJob, err := h.service.GetByProjectName(c.Context(), req.ProjectName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to check project name: %v", err),
		})
	}
	if existingJob != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": fmt.Sprintf("project name '%s' is already in use", req.ProjectName),
			"job":   existingJob,
		})
	}

	// Create job first
	job := &models.Job{
		Name:        req.JobName,
		ProjectName: req.ProjectName,
		Status:      models.JobStatusPending,
		WebhookURL:  req.WebhookURL,
		OwnerID:     0, // Default owner ID for now
	}

	job, err = h.service.CreateJob(c.Context(), job)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to create job: %v", err),
		})
	}

	// Create instances using the instance service
	if err := h.instanceService.CreateInstance(c.Context(), job.ID, req.InstanceName, req.Instances); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(job)
}
