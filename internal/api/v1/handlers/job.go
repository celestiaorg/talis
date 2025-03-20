package handlers

import (
	"fmt"
	"strconv"

	fiber "github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// JobHandler handles HTTP requests for job operations
type JobHandler struct {
	service         *services.Job
	instanceService *services.Instance
}

// NewJobHandler creates a new job handler instance
func NewJobHandler(s *services.Job, instanceService *services.Instance) *JobHandler {
	return &JobHandler{
		service:         s,
		instanceService: instanceService,
	}
}

// GetJobStatus handles the request to get a job's status
func (h *JobHandler) GetJobStatus(c *fiber.Ctx) error {
	jobIDStr := c.Params("id")
	if jobIDStr == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput("invalid job id"))
	}

	ownerID := 0 // TODO: get owner id from the JWT token
	jobID, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput("invalid job id"))
	}

	status, err := h.service.GetJobStatus(c.Context(), uint(ownerID), uint(jobID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(errServer(err.Error()))
	}

	return c.JSON(Response{
		Slug: SuccessSlug,
		Data: status,
	})
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
		return c.Status(fiber.StatusInternalServerError).
			JSON(errServer(err.Error()))
	}

	return c.JSON(Response{
		Slug: SuccessSlug,
		Data: jobs,
	})
}

// CreateJob handles the request to create a new job
func (h *JobHandler) CreateJob(c *fiber.Ctx) error {
	var jobReq infrastructure.JobRequest
	if err := c.BodyParser(&jobReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput(err.Error()))
	}

	if err := jobReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput(err.Error()))
	}

	ownerID := 0 // TODO: get owner id from the JWT token

	err := h.service.CreateJob(c.Context(), uint(ownerID), &jobReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(errServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).
		JSON(Response{
			Slug: SuccessSlug,
		})
}

// TerminateJob handles the request to terminate a job
func (h *JobHandler) TerminateJob(c *fiber.Ctx) error {
	// TODO: this should be job name as jobID is a db internal thing
	jobID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput(fmt.Sprintf("invalid job id: %v", err)))
	}

	ownerID := 0 // TODO: get owner id from the JWT token

	err = h.service.TerminateJob(c.Context(), uint(ownerID), uint(jobID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(errServer(err.Error()))
	}

	return c.Status(fiber.StatusOK).
		JSON(Response{
			Slug: SuccessSlug,
			Data: "Job terminated successfully",
		})
}

// UpdateJob handles the request to update a job
func (h *JobHandler) UpdateJob(c *fiber.Ctx) error {
	// Implementation for updating a job
	return nil
}

// SearchJobs handles the request to search jobs
func (h *JobHandler) SearchJobs(c *fiber.Ctx) error {
	// Implementation for searching jobs
	return nil
}
