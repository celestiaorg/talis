// Package handlers provides HTTP request handlers for the API
package handlers

import (
	"fmt"

	fiber "github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
)

// InstanceHandler handles HTTP requests for instance operations
type InstanceHandler struct {
	*APIHandler
}

// NewInstanceHandler creates a new instance handler instance
func NewInstanceHandler(api *APIHandler) *InstanceHandler {
	return &InstanceHandler{
		APIHandler: api,
	}
}

// ListInstances godoc
// @Summary List all instances
// @Description Returns a list of all instances with pagination
// @Tags instances
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return (default 10)"
// @Param offset query int false "Number of items to skip (default 0)"
// @Param include_deleted query bool false "Include deleted instances (default false)"
// @Param status query string false "Filter by instance status"
// @Success 200 {object} types.InstanceListResponse "List of instances"
// @Failure 400 {object} types.ErrorResponse "Invalid input"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances [get]
// @OperationId listAllInstances
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

	instances, err := h.instance.ListInstances(c.Context(), models.AdminID, &opts)
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

// GetInstance godoc
// @Summary Get instance details
// @Description Returns details of a specific instance by ID
// @Tags instances
// @Accept json
// @Produce json
// @Param id path int true "Instance ID"
// @Success 200 {object} models.Instance "Instance details"
// @Failure 400 {object} types.ErrorResponse "Invalid input"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances/{id} [get]
// @OperationId getInstanceById
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
	instance, err := h.instance.GetInstance(c.Context(), models.AdminID, uint(instanceID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get instance: %v", err),
		})
	}

	return c.JSON(instance)
}

// CreateInstance godoc
// @Summary Create new instances
// @Description Creates one or more new instances
// @Tags instances
// @Accept json
// @Produce json
// @Param request body []types.InstanceRequest true "Instance creation requests"
// @Success 201 {object} types.SuccessResponse "Created instances"
// @Failure 400 {object} types.ErrorResponse "Invalid input"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances [post]
// @OperationId createInstances
func (h *InstanceHandler) CreateInstance(c *fiber.Ctx) error {
	var instanceReqs []types.InstanceRequest
	if err := c.BodyParser(&instanceReqs); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(err.Error()))
	}

	if len(instanceReqs) == 0 {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput("at least one instance request is required"))
	}

	// NOTE: in order to update the underlying instanceReqs, we need to iterate over the slice with the index. If you use range, you will get a copy of the slice and not the original.
	for i := range instanceReqs {
		instanceReqs[i].Action = "create"
		if err := instanceReqs[i].Validate(); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(types.ErrInvalidInput(err.Error()))
		}
	}

	createdInstances, err := h.instance.CreateInstance(c.Context(), instanceReqs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).
		JSON(types.Success(createdInstances))
}

// GetPublicIPs godoc
// @Summary Get public IPs
// @Description Returns a list of public IPs for all instances
// @Tags instances
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return (default 10)"
// @Param offset query int false "Number of items to skip (default 0)"
// @Param include_deleted query bool false "Include deleted instances (default false)"
// @Success 200 {object} types.PublicIPsResponse "List of public IPs"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances/public-ips [get]
// @OperationId getInstancePublicIPs
func (h *InstanceHandler) GetPublicIPs(c *fiber.Ctx) error {
	fmt.Println("🔍 Getting all public IPs...")

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

	// Get instances with their details using the service
	instances, err := h.instance.ListInstances(c.Context(), models.AdminID, &opts)
	if err != nil {
		fmt.Printf("❌ Error getting public IPs: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to get public IPs: %v", err),
		})
	}

	fmt.Printf("✅ Found %d instances\n", len(instances))

	// Extract the public IPs from the instances
	publicIPs := make([]types.PublicIPs, len(instances))
	for i, instance := range instances {
		publicIPs[i] = types.PublicIPs{
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

// GetAllMetadata godoc
// @Summary Get all instance metadata
// @Description Returns detailed metadata for all instances
// @Tags instances
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return (default 10)"
// @Param offset query int false "Number of items to skip (default 0)"
// @Param include_deleted query bool false "Include deleted instances (default false)"
// @Success 200 {object} types.InstanceListResponse "List of instance metadata"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances/all-metadata [get]
// @OperationId getAllInstanceMetadata
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

	// Get instances with their details using the service
	instances, err := h.instance.ListInstances(c.Context(), models.AdminID, &opts)
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

// GetInstances godoc
// @Summary List instances
// @Description Returns a list of instances with pagination and optional filtering
// @Tags instances
// @Accept json
// @Produce json
// @Param limit query int false "Number of items to return (default 10)"
// @Param offset query int false "Number of items to skip (default 0)"
// @Param include_deleted query bool false "Include deleted instances (default false)"
// @Param status query string false "Filter by instance status"
// @Success 200 {object} types.InstanceListResponse "List of instances"
// @Failure 400 {object} types.ErrorResponse "Invalid input"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances [get]
// @OperationId getInstancesList
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

	instances, err := h.instance.ListInstances(c.Context(), models.AdminID, &opts)
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

// TerminateInstances godoc
// @Summary Terminate instances
// @Description Terminates one or more instances
// @Tags instances
// @Accept json
// @Produce json
// @Param request body types.DeleteInstancesRequest true "Termination request"
// @Success 200 {object} types.SuccessResponse "Instances terminated successfully"
// @Failure 400 {object} types.ErrorResponse "Invalid input"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances [delete]
// @OperationId terminateInstances
func (h *InstanceHandler) TerminateInstances(c *fiber.Ctx) error {
	var deleteReq types.DeleteInstancesRequest
	if err := c.BodyParser(&deleteReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(err.Error()))
	}

	if deleteReq.ProjectName == "" || len(deleteReq.InstanceIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "project name and instance IDs are required",
		})
	}

	err := h.instance.Terminate(c.Context(), deleteReq.OwnerID, deleteReq.ProjectName, deleteReq.InstanceIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to terminate instances: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).
		JSON(types.Success(nil))
}
