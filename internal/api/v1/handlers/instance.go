package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// InstanceHandler handles HTTP requests for instance operations
type InstanceHandler struct {
	service *services.InstanceService
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(service *services.InstanceService) *InstanceHandler {
	return &InstanceHandler{
		service: service,
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

// CreateInstance handles the request to create a new instance
func (h *InstanceHandler) CreateInstance(c *fiber.Ctx) error {
	var req struct {
		Name        string                           `json:"name"`
		ProjectName string                           `json:"project_name"`
		WebhookURL  string                           `json:"webhook_url"`
		Instances   []infrastructure.InstanceRequest `json:"instances"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create instance using the service
	job, err := h.service.CreateInstance(c.Context(), req.Name, req.ProjectName, req.WebhookURL, req.Instances)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(job)
}

// DeleteInstance handles the request to delete an instance
func (h *InstanceHandler) DeleteInstance(c *fiber.Ctx) error {
	instanceID := c.Params("id")
	if instanceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "instance id is required",
		})
	}

	// Delete instance using the service
	job, err := h.service.DeleteInstance(c.Context(), instanceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(job)
}

// GetInstance returns details of a specific instance
func (h *InstanceHandler) GetInstance(c *fiber.Ctx) error {
	instanceID := c.Params("id")
	if instanceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "instance id is required",
		})
	}

	id, err := strconv.ParseUint(instanceID, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid instance id",
		})
	}

	// Get instance using the service
	instance, err := h.service.GetInstance(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instance: %v", err),
		})
	}

	return c.JSON(instance)
}
