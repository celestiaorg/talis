// Package handlers provides HTTP request handling
package handlers

import (
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/internal/types"

	fiber "github.com/gofiber/fiber/v2"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	taskService *services.TaskService
}

// NewTaskHandler creates a new instance of TaskHandler
func NewTaskHandler(taskService *services.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// GetTask handles retrieving a task by name within a project
func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	projectName := c.Params("name")
	if projectName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Project name is required"))
	}

	taskName := c.Params("taskName")
	if taskName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Task name is required"))
	}

	ownerID := uint(0) // TODO: get owner id from the JWT token

	task, err := h.taskService.GetByName(c.Context(), ownerID, projectName, taskName)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(types.ErrNotFound(err.Error()))
	}

	return c.JSON(types.Success(task))
}

// ListProjectTasks handles retrieving all tasks for a project with pagination
func (h *TaskHandler) ListProjectTasks(c *fiber.Ctx) error {
	projectName := c.Params("name")
	if projectName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Project name is required"))
	}

	page := c.QueryInt("page", 1)
	listOpts := getPaginationOptions(page)

	ownerID := uint(0) // TODO: get owner id from the JWT token

	tasks, err := h.taskService.ListByProject(c.Context(), ownerID, projectName, listOpts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrServer(err.Error()))
	}

	return c.JSON(types.Success(map[string]interface{}{
		"tasks": tasks,
		"pagination": types.PaginationResponse{
			Total:  len(tasks),
			Page:   page,
			Limit:  listOpts.Limit,
			Offset: listOpts.Offset,
		},
	}))
}

// DeleteTask handles deleting a task by name
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	projectName := c.Params("name")
	if projectName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Project name is required"))
	}

	taskName := c.Params("taskName")
	if taskName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Task name is required"))
	}

	ownerID := uint(0) // TODO: get owner id from the JWT token

	if err := h.taskService.Delete(c.Context(), ownerID, projectName, taskName); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrServer(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateTaskStatus handles updating the status of a task
func (h *TaskHandler) UpdateTaskStatus(c *fiber.Ctx) error {
	projectName := c.Params("name")
	if projectName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Project name is required"))
	}

	taskName := c.Params("taskName")
	if taskName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Task name is required"))
	}

	var statusUpdate struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&statusUpdate); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.ErrInvalidInput("Invalid request body"))
	}

	ownerID := uint(0) // TODO: get owner id from the JWT token

	if err := h.taskService.UpdateStatus(c.Context(), ownerID, projectName, taskName, statusUpdate.Status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(types.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(types.Success(nil))
}
