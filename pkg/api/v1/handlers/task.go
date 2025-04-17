// Package handlers provides HTTP request handling
package handlers

import (
	"errors"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/internal/types"
	"gorm.io/gorm"

	fiber "github.com/gofiber/fiber/v2"
)

// TaskHandlers contains all task related handlers
type TaskHandlers struct {
	service *services.Task
}

// NewTaskHandlers creates a new task handlers instance
func NewTaskHandlers(taskService *services.Task) *TaskHandlers {
	return &TaskHandlers{
		service: taskService,
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

	task, err := h.service.GetByName(c.Context(), ownerID, params.TaskName)
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

	tasks, err := h.service.ListByProject(c.Context(), ownerID, params.ProjectName, listOpts)
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

	// First update the task status to "terminated"
	if err := h.service.UpdateStatusByName(c.Context(), ownerID, params.TaskName, models.TaskStatusTerminated); err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgTaskTerminateFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Success: true,
		ID:      req.ID,
	})
}
