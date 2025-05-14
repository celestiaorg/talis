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

// Get handles retrieving a task by id
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

// List handles listing all tasks for a project with pagination
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

// Terminate handles terminating a running task
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

// ListByInstanceID handles listing tasks for a specific instance ID, with optional action filter and pagination.
// This is a direct REST endpoint, not an RPC one.
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
