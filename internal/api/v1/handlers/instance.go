package handlers

import (
	"fmt"

	fiber "github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db/models"
)

// InstanceHandler handles HTTP requests for instance operations
type InstanceHandler struct {
	service    *services.InstanceService
	jobService *services.JobService
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(service *services.InstanceService, jobService *services.JobService) *InstanceHandler {
	return &InstanceHandler{
		service:    service,
		jobService: jobService,
	}
}

// ListInstances handles the request to list all instances
func (h *InstanceHandler) ListInstances(c *fiber.Ctx) error {
	var (
		limit  = c.QueryInt("limit", 10)
		offset = c.QueryInt("offset", 0)
	)

	instances, err := h.service.ListInstances(c.Context(), &models.ListOptions{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list instances: %v", err),
		})
	}

	return c.JSON(instances)
}

// GetInstance returns details of a specific instance
func (h *InstanceHandler) GetInstance(c *fiber.Ctx) error {
	instanceID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput(fmt.Sprintf("instance id is required: %v", err)))
	}

	// Get instance using the service
	// TODO: Consider passing OwnerID for security purposes
	instance, err := h.service.GetInstance(c.Context(), uint(instanceID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instance: %v", err),
		})
	}

	return c.JSON(instance)
}

// CreateInstance handles the request to create instances
func (h *InstanceHandler) CreateInstance(c *fiber.Ctx) error {
	// Implementation for creating instances
	return nil
}
