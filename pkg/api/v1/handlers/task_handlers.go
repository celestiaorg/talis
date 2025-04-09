// Package handlers provides HTTP request handling
package handlers

import (
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/internal/types"

	fiber "github.com/gofiber/fiber/v2"
)

// TaskHandlers contains all task related handlers
type TaskHandlers struct {
	taskService *services.TaskService
}

// NewTaskHandlers creates a new task handlers instance
func NewTaskHandlers(taskService *services.TaskService) *TaskHandlers {
	return &TaskHandlers{
		taskService: taskService,
	}
}

// Get handles retrieving a task by name
func (h *TaskHandlers) Get(c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[TaskGetParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	task, err := h.taskService.GetByName(c.Context(), ownerID, params.ProjectName, params.TaskName)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusNotFound, ErrMsgTaskNotFound, err.Error(), req.ID)
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

	tasks, err := h.taskService.ListByProject(c.Context(), ownerID, params.ProjectName, listOpts)
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

// Abort handles aborting a task by name
func (h *TaskHandlers) Abort(c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[TaskAbortParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	// First update the task status to "aborted"
	if err := h.taskService.UpdateStatus(c.Context(), ownerID, params.ProjectName, params.TaskName, "aborted"); err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgTaskAbortFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Success: true,
		ID:      req.ID,
	})
}

// UpdateStatus handles updating a task's status
func (h *TaskHandlers) UpdateStatus(c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[TaskUpdateStatusParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	if err := h.taskService.UpdateStatus(c.Context(), ownerID, params.ProjectName, params.TaskName, params.Status); err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgTaskStatusFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Success: true,
		ID:      req.ID,
	})
}
