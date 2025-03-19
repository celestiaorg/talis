package handlers

import (
	"fmt"

	fiber "github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// InstanceHandler handles HTTP requests for instance operations
type InstanceHandler struct {
	service *services.Instance
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(service *services.Instance) *InstanceHandler {
	return &InstanceHandler{
		service: service,
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
	var instancesReq infrastructure.InstancesRequest
	if err := c.BodyParser(&instancesReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput(err.Error()))
	}
	instancesReq.Action = "create"

	if err := instancesReq.Validate(); err != nil {
		fmt.Printf("err val: %v\n", err)
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput(err.Error()))
	}

	ownerID := 0 // TODO: get owner id from the JWT token

	err := h.service.CreateInstance(c.Context(), uint(ownerID), instancesReq.JobName, instancesReq.Instances)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(errServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).
		JSON(Response{
			Slug: SuccessSlug,
		})
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

// TerminateInstances handles the request to terminate instances
func (h *InstanceHandler) TerminateInstances(c *fiber.Ctx) error {
	var deleteReq struct {
		JobName     string   `json:"job_name"`
		InstanceIDs []string `json:"instance_ids"`
	}
	if err := c.BodyParser(&deleteReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(errInvalidInput(err.Error()))
	}

	if deleteReq.JobName == "" || len(deleteReq.InstanceIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "job name and instance names are required",
		})
	}

	ownerID := 0 // TODO: get owner id from the JWT token

	err := h.service.Terminate(c.Context(), uint(ownerID), deleteReq.JobName, deleteReq.InstanceIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to terminate instances: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "instances terminated successfully",
	})
}
