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

// GetPublicIPs returns a list of public IPs and their associated job IDs
func (h *InstanceHandler) GetPublicIPs(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting public IPs...")

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 10)
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Get instances with their public IPs using the service
	instances, err := h.service.GetPublicIPs(c.Context(), &models.ListOptions{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		fmt.Printf("❌ Error getting public IPs: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get public IPs: %v", err),
		})
	}

	fmt.Printf("✅ Found %d instances\n", len(instances))

	// Convert instances to simplified format with only public IPs and job IDs
	publicIPs := make([]map[string]interface{}, len(instances))
	for i, instance := range instances {
		publicIPs[i] = map[string]interface{}{
			"job_id":    instance.JobID,
			"public_ip": instance.PublicIP,
		}
	}

	// Return instances with pagination info
	return c.JSON(fiber.Map{
		"instances": publicIPs,
		"total":     len(instances),
		"page":      page,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetAllMetadata returns a list of all instance details
func (h *InstanceHandler) GetAllMetadata(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting all instance metadata...")

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit := c.QueryInt("limit", 10)
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Get instances with their details using the service
	instances, err := h.service.GetPublicIPs(c.Context(), &models.ListOptions{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		fmt.Printf("❌ Error getting instance metadata: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instance metadata: %v", err),
		})
	}

	fmt.Printf("✅ Found %d instances\n", len(instances))

	// Return instances with pagination info
	return c.JSON(fiber.Map{
		"instances": instances,
		"total":     len(instances),
		"page":      page,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetInstancesByJobID returns a list of instances for a specific job
func (h *InstanceHandler) GetInstancesByJobID(c *fiber.Ctx) error {
	jobIDStr := c.Params("jobId")
	if jobIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "job id is required",
		})
	}

	jobID, err := strconv.ParseUint(jobIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid job id",
		})
	}

	fmt.Printf("🔍 Getting instances for job ID %d...\n", jobID)

	// Get instances using the service
	instances, err := h.service.GetInstancesByJobID(c.Context(), uint(jobID))
	if err != nil {
		fmt.Printf("❌ Error getting instances for job %d: %v\n", jobID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instances for job %d: %v", jobID, err),
		})
	}

	fmt.Printf("✅ Found %d instances for job %d\n", len(instances), jobID)

	// Convert instances to simplified format with only public IPs
	publicIPs := make([]string, len(instances))
	for i, instance := range instances {
		publicIPs[i] = instance.PublicIP
	}

	// Return the public IPs
	return c.JSON(fiber.Map{
		"public_ips": publicIPs,
		"total":      len(instances),
		"job_id":     jobID,
	})
}
