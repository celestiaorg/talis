// Package handlers provides HTTP request handlers for the API
package handlers

import (
	"fmt"

	fiber "github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/internal/types"
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
	var opts models.ListOptions
	opts.Limit = c.QueryInt("limit", DefaultPageSize)
	opts.Offset = c.QueryInt("offset", 0)
	opts.IncludeDeleted = c.QueryBool("include_deleted", false)

	// Handle status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status, err := models.ParseInstanceStatus(statusStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(types.ErrInvalidInput(fmt.Sprintf("invalid instance status: %v", err)))
		}
		opts.InstanceStatus = &status
	} else if !opts.IncludeDeleted && opts.InstanceStatus == nil {
		// By default, exclude terminated instances if not including deleted
		defaultStatus := models.InstanceStatusTerminated
		opts.InstanceStatus = &defaultStatus
		opts.StatusFilter = models.StatusFilterNotEqual
	}

	// TODO: should check for OwnerID and filter by it

	instances, err := h.service.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list instances: %v", err),
		})
	}

	return c.JSON(types.ListResponse[models.Instance]{
		Rows: instances,
		Pagination: types.PaginationResponse{
			Total:  len(instances),
			Page:   1,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	})
}

// GetInstance returns details of a specific instance
func (h *InstanceHandler) GetInstance(c *fiber.Ctx) error {
	instanceID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(fmt.Sprintf("instance id is required: %v", err)))
	}

	if instanceID <= 0 {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput("instance id must be positive"))
	}

	// Get instance using the service
	instance, err := h.service.GetInstance(c.Context(), models.AdminID, uint(instanceID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instance: %v", err),
		})
	}

	return c.JSON(instance)
}

// CreateInstance handles the request to create instances
// TODO: this should return the Instance ID so that it can be immediately queried.
func (h *InstanceHandler) CreateInstance(c *fiber.Ctx) error {
	var instancesReq types.InstancesRequest
	if err := c.BodyParser(&instancesReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(err.Error()))
	}
	instancesReq.Action = "create"

	if err := instancesReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(err.Error()))
	}

	err := h.service.CreateInstance(c.Context(), models.AdminID, instancesReq.JobName, instancesReq.Instances)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).
		JSON(types.Success(nil))
}

// GetPublicIPs returns a list of public IPs and their associated job IDs
func (h *InstanceHandler) GetPublicIPs(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting public IPs...")

	var opts models.ListOptions
	opts.Limit = c.QueryInt("limit", DefaultPageSize)
	opts.Offset = c.QueryInt("offset", 0)
	opts.IncludeDeleted = c.QueryBool("include_deleted", false)

	// Only apply default status filter if IncludeDeleted is false
	if !opts.IncludeDeleted && opts.InstanceStatus == nil {
		defaultStatus := models.InstanceStatusTerminated
		opts.InstanceStatus = &defaultStatus
		opts.StatusFilter = models.StatusFilterNotEqual
	}

	// Get instances
	instances, err := h.service.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		fmt.Printf("❌ Error getting instances: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get public IPs: %v", err),
		})
	}

	fmt.Printf("✅ Found %d instances\n", len(instances))

	// Extract the public IPs from the instances
	publicIPs := make([]types.PublicIPs, len(instances))
	for i, instance := range instances {
		publicIPs[i] = types.PublicIPs{
			JobID:    instance.JobID,
			PublicIP: instance.PublicIP,
		}
	}

	// Return instances with pagination info
	return c.JSON(types.PublicIPsResponse{
		PublicIPs: publicIPs,
		Pagination: types.PaginationResponse{
			Total:  len(instances),
			Page:   1,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	})
}

// GetAllMetadata returns a list of all instance details
func (h *InstanceHandler) GetAllMetadata(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting all instance metadata...")

	var opts models.ListOptions
	opts.Limit = c.QueryInt("limit", DefaultPageSize)
	opts.Offset = c.QueryInt("offset", 0)
	opts.IncludeDeleted = c.QueryBool("include_deleted", false)

	// Only apply default status filter if IncludeDeleted is false
	if !opts.IncludeDeleted && opts.InstanceStatus == nil {
		defaultStatus := models.InstanceStatusTerminated
		opts.InstanceStatus = &defaultStatus
		opts.StatusFilter = models.StatusFilterNotEqual
	}

	// TODO: should check for JobID and filter by it
	// TODO: should check for OwnerID and filter by it

	// Get instances with their details using the service
	instances, err := h.service.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		fmt.Printf("❌ Error getting instance: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instance metadata: %v", err),
		})
	}

	fmt.Printf("✅ Found %d instances\n", len(instances))

	// Return instances with pagination info
	return c.JSON(types.ListResponse[models.Instance]{
		Rows: instances,
		Pagination: types.PaginationResponse{
			Total:  len(instances),
			Page:   1,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
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
	instances, err := h.service.GetInstancesByJobID(c.Context(), models.AdminID, uint(jobID))
	if err != nil {
		fmt.Printf("❌ Error getting instances for job %d: %v\n", jobID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instances for job %d: %v", jobID, err),
		})
	}

	fmt.Printf("✅ Found %d instances for job %d\n", len(instances), jobID)

	// Return response using JobInstancesResponse type
	return c.JSON(types.JobInstancesResponse{
		Instances: instances,
		Total:     len(instances),
		JobID:     uint(jobID),
	})
}

// TerminateInstances handles the request to terminate instances
func (h *InstanceHandler) TerminateInstances(c *fiber.Ctx) error {
	var deleteReq types.DeleteInstanceRequest
	if err := c.BodyParser(&deleteReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(err.Error()))
	}

	if deleteReq.JobName == "" || len(deleteReq.InstanceNames) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "job name and instance names are required",
		})
	}

	err := h.service.Terminate(c.Context(), models.AdminID, deleteReq.JobName, deleteReq.InstanceNames)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to terminate instances: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(types.Success(nil))
}

// GetInstances handles the request to list instances
func (h *InstanceHandler) GetInstances(c *fiber.Ctx) error {
	var opts models.ListOptions
	opts.Limit = c.QueryInt("limit", DefaultPageSize)
	opts.Offset = c.QueryInt("offset", 0)
	opts.IncludeDeleted = c.QueryBool("include_deleted", false)

	// Handle status filter
	if statusStr := c.Query("status"); statusStr != "" {
		status, err := models.ParseInstanceStatus(statusStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid instance status: %v", err),
			})
		}
		opts.InstanceStatus = &status
	}

	instances, err := h.service.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to list instances: %v", err),
		})
	}

	return c.JSON(types.ListResponse[models.Instance]{
		Rows: instances,
		Pagination: types.PaginationResponse{
			Total:  len(instances),
			Page:   1,
			Limit:  opts.Limit,
			Offset: opts.Offset,
		},
	})
}
