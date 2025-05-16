// Package handlers provides HTTP request handling
package handlers

import (
	"errors"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
	"gorm.io/gorm"

	fiber "github.com/gofiber/fiber/v2"
)

// TaskHandlers contains all task related handlers
type TaskHandlers struct {
	*APIHandler
}

// NewTaskHandlers creates a new task handlers instance
func NewTaskHandlers(api *APIHandler) *TaskHandlers {
	return &TaskHandlers{
		APIHandler: api,
	}
}

// Get godoc
// @Summary Get task by ID
// @Description Retrieves a task by its ID via RPC
// @Tags tasks,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with TaskGetParams"
// @Success 200 {object} RPCResponse{data=models.Task} "Task details"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 404 {object} RPCResponse "Task not found"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId getTaskById
func (h *TaskHandlers) Get(c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[TaskGetParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	task, err := h.task.Get(c.Context(), ownerID, params.TaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return respondWithRPCError(c, fiber.StatusNotFound, ErrMsgTaskNotFound, err.Error(), req.ID)
		}
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgTaskGetFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data:    task,
		Success: true,
		ID:      req.ID,
	})
}

// List godoc
// @Summary List tasks for a project
// @Description Returns a list of tasks for a specific project with pagination via RPC
// @Tags tasks,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with TaskListParams"
// @Success 200 {object} RPCResponse{data=types.TaskListResponse} "List of tasks"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId listTasksByProject
func (h *TaskHandlers) List(c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[TaskListParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	page := 1
	if params.Page > 0 {
		page = params.Page
	}

	listOpts := getPaginationOptions(page)

	tasks, err := h.task.ListByProject(c.Context(), ownerID, params.ProjectName, listOpts)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgTaskListFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data: types.ListResponse[models.Task]{
			Rows: tasks,
			Pagination: types.PaginationResponse{
				Total:  len(tasks),
				Page:   page,
				Limit:  listOpts.Limit,
				Offset: listOpts.Offset,
			},
		},
		Success: true,
		ID:      req.ID,
	})
}

// Terminate godoc
// @Summary Terminate a task
// @Description Terminates a running task via RPC
// @Tags tasks,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with TaskTerminateParams"
// @Success 200 {object} RPCResponse "Task terminated successfully"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId terminateTask
func (h *TaskHandlers) Terminate(c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[TaskTerminateParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	if err := h.task.UpdateStatus(c.Context(), ownerID, params.TaskID, models.TaskStatusTerminated); err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgTaskTerminateFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Success: true,
		ID:      req.ID,
	})
}

// ListByInstanceID godoc
// @Summary List tasks for an instance
// @Description Returns a list of tasks for a specific instance with optional filtering and pagination
// @Tags tasks
// @Accept json
// @Produce json
// @Param instance_id path int true "Instance ID"
// @Param action query string false "Filter by task action"
// @Param limit query int false "Number of items to return (default 10)"
// @Param offset query int false "Number of items to skip (default 0)"
// @Success 200 {object} types.SuccessResponse{data=types.TaskListResponse} "List of tasks"
// @Failure 400 {object} types.ErrorResponse "Invalid input"
// @Failure 500 {object} types.ErrorResponse "Internal server error"
// @Router /instances/{instance_id}/tasks [get]
func (h *TaskHandlers) ListByInstanceID(c *fiber.Ctx) error {
	// TODO: Extract OwnerID from authenticated context (e.g., c.Locals("userID").(uint))
	// For now, using a placeholder or admin ID. This needs proper auth integration.
	ownerID := models.AdminID

	instanceID, err := c.ParamsInt("instance_id")
	if err != nil || instanceID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Invalid or missing instance_id parameter"))
	}

	params := TaskListByInstanceParams{
		InstanceID: uint(instanceID),
		OwnerID:    ownerID,
		Action:     c.Query("action"),
		Limit:      c.QueryInt("limit", DefaultPageSize),
		Offset:     c.QueryInt("offset", 0),
	}

	if err := params.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput(err.Error()))
	}

	listOpts := &models.ListOptions{
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	tasks, err := h.task.ListTasksByInstanceID(c.Context(), params.OwnerID, params.InstanceID, models.TaskAction(params.Action), listOpts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrServer("Failed to retrieve tasks for the instance"))
	}

	page := 0
	if params.Limit > 0 {
		page = (params.Offset / params.Limit) + 1
	}

	return c.Status(fiber.StatusOK).JSON(types.Success(types.ListResponse[models.Task]{
		Rows: tasks,
		Pagination: types.PaginationResponse{
			Total:  len(tasks),
			Limit:  params.Limit,
			Offset: params.Offset,
			Page:   page,
		},
	}))
}
