package handlers

import (
	"strconv"

	fiber "github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// JobHandler handles HTTP requests for job operations
type JobHandler struct {
	jobService      *services.Job
	instanceService *services.Instance
}

// NewJobHandler creates a new job handler instance
func NewJobHandler(s *services.Job, instanceService *services.Instance) *JobHandler {
	return &JobHandler{
		jobService:      s,
		instanceService: instanceService,
	}
}

// GetJob handles the request to get a job
func (h *JobHandler) GetJob(c *fiber.Ctx) error {
	jobIDStr := c.Params("id")
	if jobIDStr == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput("invalid job id"))
	}

	jobID, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput("invalid job id"))
	}

	job, err := h.jobService.GetJob(c.Context(), uint(jobID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.JSON(infrastructure.Success(job))
}

// GetJobStatus handles the request to get a job's status
func (h *JobHandler) GetJobStatus(c *fiber.Ctx) error {
	jobIDStr := c.Params("id")
	if jobIDStr == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput("invalid job id"))
	}

	// Using 32-bit limit for ParseUint
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput("invalid job id"))
	}
	// Using uint(jobID) is safe because ParseUint guarantees non-negative values

	status, err := h.jobService.GetJobStatus(c.Context(), models.AdminID, uint(jobID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.JSON(infrastructure.Success(status))
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
	paginationOpts := getPaginationOptions(page)

	jobs, err := h.jobService.ListJobs(c.Context(), status, models.AdminID, paginationOpts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.JSON(infrastructure.ListJobsResponse{
		Slug: infrastructure.SuccessSlug,
		Jobs: jobs,
		Pagination: infrastructure.PaginationResponse{
			Total: len(jobs),
		},
	})
}

// CreateJob handles the request to create a new job
// TODO: this should return the Job ID so that it can be immediately queried.
func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	var jobReq infrastructure.JobRequest
	if err := c.BodyParser(&jobReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput(err.Error()))
	}

	if err := jobReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput(err.Error()))
	}

	err := h.jobService.CreateJob(c.Context(), models.AdminID, &jobReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).
		JSON(infrastructure.Success(nil))
}

// TerminateJob handles the request to terminate a job
func (h *JobHandler) TerminateJob(c *fiber.Ctx) error {
	// TODO: this should be job name as jobID is a db internal thing
	jobIDStr := c.Params("id")
	if jobIDStr == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput("invalid job id"))
	}

	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput("invalid job id"))
	}

	err = h.jobService.TerminateJob(c.Context(), models.AdminID, uint(jobID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusOK).
		JSON(infrastructure.Success("Job terminated successfully"))
}

// UpdateJob handles the request to update a job
func (h *JobHandler) UpdateJob(_ *fiber.Ctx) error {
	// Implementation for updating a job
	return nil
}
