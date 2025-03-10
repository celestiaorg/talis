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
	service    services.InstanceServiceInterface
	jobService services.JobServiceInterface
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(service services.InstanceServiceInterface, jobService services.JobServiceInterface) *InstanceHandler {
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

	// Check if project name already exists
	existingJob, err := h.jobService.GetByProjectName(c.Context(), req.ProjectName)
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
	var req infrastructure.DeleteInstanceRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("invalid request body: %v", err),
		})
	}

	if req.ID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "job id is required",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	if req.ProjectName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project_name is required",
		})
	}

	if len(req.Instances) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "at least one instance is required",
		})
	}

	// Delete instance using the service
	job, err := h.service.DeleteInstance(c.Context(), req.ID, req.Name, req.ProjectName, req.Instances)
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

// GetPublicIPs returns a list of all public IPs and instance details
func (h *InstanceHandler) GetPublicIPs(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting public IPs...")

	// Get instances with their public IPs using the service
	instances, err := h.service.GetPublicIPs(c.Context())
	if err != nil {
		fmt.Printf("❌ Error getting public IPs: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get public IPs: %v", err),
		})
	}

	fmt.Printf("✅ Found %d instances\n", len(instances))

	// Return the instances with their details
	return c.JSON(fiber.Map{
		"instances": instances,
		"total":     len(instances),
	})
}
