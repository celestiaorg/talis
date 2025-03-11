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
	var (
		limit  = c.QueryInt("limit", 10)
		offset = c.QueryInt("offset", 0)
	)
	status, err := models.ParseJobStatus(c.Query("status"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput("invalid job status"))
	}

	ownerID := 0 // TODO: get owner id from the JWT token

	jobs, err := h.service.ListJobs(c.Context(), status, uint(ownerID), &models.ListOptions{
		Limit:  limit,
		Offset: offset,
	})
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
	jobReq.Action = "create"

	if err := jobReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput(err.Error()))
	}

	ownerID := 0 // TODO: get owner id from the JWT token

	job, err := h.service.CreateJob(c.Context(), uint(ownerID), &jobReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(errServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).
		JSON(Response{
			Slug: SuccessSlug,
			Data: job,
		})
}

// TerminateJob handles the request to terminate a job
func (h *JobHandler) TerminateJob(c *fiber.Ctx) error {
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
