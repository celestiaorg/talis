package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// InstanceHandler handles HTTP requests for instance operations
type InstanceHandler struct {
	service    services.Instance
	jobService services.Job
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(service services.Instance, jobService services.Job) *InstanceHandler {
	return &InstanceHandler{
		service:    service,
		jobService: jobService,
	}
}

// ListInstances handles the request to list all instances
func (h *InstanceHandler) ListInstances(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", DefaultPageSize)

	instances, err := h.service.ListInstances(c.Context(), getPaginationOptions(page, limit))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list instances: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"instances": instances,
		"page":      page,
		"limit":     limit,
	})
}

// CreateInstance handles the request to create a new instance
func (h *InstanceHandler) CreateInstance(c *fiber.Ctx) error {
	// Get jobId from path parameter
	jobID, err := c.ParamsInt("jobId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid job id",
		})
	}

	var req infrastructure.InstanceCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create instance using the service
	if err := h.service.CreateInstance(c.Context(), uint(jobID), req.InstanceName, req.Instances); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Instance creation initiated",
		"job_id":  jobID,
	})
}

// DeleteInstance handles the request to delete instance(s) from a job in FIFO order
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

	if req.InstanceName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "instance_name is required",
		})
	}

	if req.ProjectName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project_name is required",
		})
	}

	// TODO: I think if no instances are provided it should just delete all instances for the job.
	// In order to do this we need the provider info which will be a DB request from DeleteInstance.
	if len(req.Instances) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "at least one instance is required",
		})
	}

	// Delete instance using the service
	job, err := h.service.DeleteInstance(c.Context(), req.ID, req.InstanceName, req.ProjectName, req.Instances)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(job)
}

// GetInstance returns details of a specific instance
func (h *InstanceHandler) GetInstance(c *fiber.Ctx) error {
	instanceID := c.Params("instanceId")
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

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", DefaultPageSize)
	paginationOpts := getPaginationOptions(page, limit)

	// Get instances with their public IPs using the service
	instances, err := h.service.GetPublicIPs(c.Context(), paginationOpts)
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
		"offset":    paginationOpts.Offset,
	})
}

// GetAllMetadata returns a list of all instance details
func (h *InstanceHandler) GetAllMetadata(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting all instance metadata...")

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", DefaultPageSize)
	paginationOpts := getPaginationOptions(page, limit)

	// Get instances with their details using the service
	instances, err := h.service.GetPublicIPs(c.Context(), paginationOpts)
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
		"offset":    paginationOpts.Offset,
	})
}

// GetInstancesByJobID returns a list of instances for a specific job
func (h *InstanceHandler) GetInstancesByJobID(c *fiber.Ctx) error {
	jobID, err := c.ParamsInt("jobId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid job id",
		})
	}
	if jobID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "job id is required",
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

	// Return all instance details
	return c.JSON(fiber.Map{
		"instances": instances,
		"total":     len(instances),
		"job_id":    jobID,
	})
}
