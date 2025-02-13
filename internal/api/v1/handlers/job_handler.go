package handlers

import (
	"fmt"

	"github.com/celestiaorg/talis/internal/domain/job"
	"github.com/gofiber/fiber/v2"
)

type JobHandler struct {
	jobService job.Service
}

func NewJobHandler(jobService job.Service) *JobHandler {
	return &JobHandler{
		jobService: jobService,
	}
}

// GetJobStatus returns the status of a specific job
func (h *JobHandler) GetJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "job id is required",
		})
	}

	status, err := h.jobService.GetJobStatus(c.Context(), jobID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get job status: %v", err),
		})
	}

	return c.JSON(status)
}

// ListJobs returns a list of all jobs
func (h *JobHandler) ListJobs(c *fiber.Ctx) error {
	// Check if a specific job_id was requested
	if jobID := c.Query("job_id"); jobID != "" {
		status, err := h.jobService.GetJobStatus(c.Context(), jobID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("failed to get job status: %v", err),
			})
		}
		return c.JSON(status)
	}

	// Otherwise, list all jobs with filters
	limit := c.Query("limit")
	status := c.Query("status")

	jobs, err := h.jobService.ListJobs(c.Context(), &job.ListOptions{
		Limit:  limit,
		Status: status,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list jobs: %v", err),
		})
	}

	return c.JSON(jobs)
}
