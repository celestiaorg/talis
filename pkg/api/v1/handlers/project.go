// Package handlers provides HTTP request handling
package handlers

import (
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/internal/types/infrastructure"

	fiber "github.com/gofiber/fiber/v2"
)

// ProjectHandler handles HTTP requests for projects
type ProjectHandler struct {
	projectService *services.ProjectService
}

// NewProjectHandler creates a new instance of ProjectHandler
func NewProjectHandler(projectService *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// CreateProject handles the creation of a new project
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var project models.Project
	if err := c.BodyParser(&project); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(infrastructure.ErrInvalidInput(err.Error()))
	}

	if project.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(infrastructure.ErrInvalidInput("Project name is required"))
	}

	ownerID := uint(0) // TODO: get owner id from the JWT token
	project.OwnerID = ownerID

	if err := h.projectService.Create(c.Context(), &project); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(infrastructure.Success(project))
}

// GetProject handles retrieving a project by name
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(infrastructure.ErrInvalidInput("Project name is required"))
	}

	ownerID := uint(0) // TODO: get owner id from the JWT token

	project, err := h.projectService.GetByName(c.Context(), ownerID, name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(infrastructure.ErrNotFound(err.Error()))
	}

	return c.JSON(infrastructure.Success(project))
}

// GetProjectByName handles retrieving a project by name
func (h *ProjectHandler) GetProjectByName(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(infrastructure.ErrInvalidInput("Project name is required"))
	}

	ownerID := uint(0) // TODO: get owner id from the JWT token

	project, err := h.projectService.GetByName(c.Context(), ownerID, name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(infrastructure.ErrNotFound(err.Error()))
	}

	return c.JSON(infrastructure.Success(project))
}

// ListProjects handles retrieving all projects with pagination
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	listOpts := getPaginationOptions(page)

	ownerID := uint(0) // TODO: get owner id from the JWT token

	projects, err := h.projectService.List(c.Context(), ownerID, listOpts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.JSON(infrastructure.Success(map[string]interface{}{
		"projects": projects,
		"pagination": infrastructure.PaginationResponse{
			Total:  len(projects),
			Page:   page,
			Limit:  listOpts.Limit,
			Offset: listOpts.Offset,
		},
	}))
}

// DeleteProject handles deleting a project by name
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(infrastructure.ErrInvalidInput("Project name is required"))
	}

	ownerID := uint(0) // TODO: get owner id from the JWT token

	if err := h.projectService.Delete(c.Context(), ownerID, name); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListProjectInstances handles retrieving all instances for a specific project
func (h *ProjectHandler) ListProjectInstances(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(infrastructure.ErrInvalidInput("Project name is required"))
	}

	ownerID := uint(0) // TODO: get owner id from the JWT token

	instances, err := h.projectService.ListInstances(c.Context(), ownerID, name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.JSON(infrastructure.Success(instances))
}
