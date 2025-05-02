// Package handlers provides HTTP request handling
package handlers

import (
	"errors"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/pkg/types"
	"gorm.io/gorm"

	fiber "github.com/gofiber/fiber/v2"
)

// ProjectHandlers contains all project related handlers
type ProjectHandlers struct {
	*APIHandler
}

// NewProjectHandlers creates a new project handlers instance
func NewProjectHandlers(api *APIHandler) *ProjectHandlers {
	return &ProjectHandlers{
		APIHandler: api,
	}
}

// Create handles creating a project
func (h *ProjectHandlers) Create(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[ProjectCreateParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	project := models.Project{
		OwnerID:     params.OwnerID,
		Name:        params.Name,
		Description: params.Description,
		Config:      params.Config,
	}

	if err := h.project.Create(c.Context(), &project); err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgProjCreateFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data:    project,
		Success: true,
		ID:      req.ID,
	})
}

// Get handles retrieving a project by name
func (h *ProjectHandlers) Get(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[ProjectGetParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	project, err := h.project.GetByName(c.Context(), params.OwnerID, params.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return respondWithRPCError(c, fiber.StatusNotFound, ErrMsgProjNotFound, err.Error(), req.ID)
		}
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgProjGetFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data:    project,
		Success: true,
		ID:      req.ID,
	})
}

// List handles listing all projects with pagination
func (h *ProjectHandlers) List(c *fiber.Ctx, req RPCRequest) error {
	page := 1

	params, err := parseParams[ProjectListParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	if params.Page > 0 {
		page = params.Page
	}

	listOpts := getPaginationOptions(page)

	projects, err := h.project.List(c.Context(), params.OwnerID, listOpts)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgProjListFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data: types.ListResponse[models.Project]{
			Rows: projects,
			Pagination: types.PaginationResponse{
				Total:  len(projects),
				Page:   page,
				Limit:  listOpts.Limit,
				Offset: listOpts.Offset,
			},
		},
		Success: true,
		ID:      req.ID,
	})
}

// Delete handles deleting a project by name
func (h *ProjectHandlers) Delete(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[ProjectDeleteParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	if err := h.project.Delete(c.Context(), params.OwnerID, params.Name); err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgProjDeleteFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Success: true,
		ID:      req.ID,
	})
}

// ListInstances handles listing all instances for a project
func (h *ProjectHandlers) ListInstances(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[ProjectListInstancesParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	listOpts := getPaginationOptions(params.Page)
	instances, err := h.project.ListInstances(c.Context(), params.OwnerID, params.ProjectID, listOpts)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, "Failed to list project instances", err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data: types.ListResponse[models.Instance]{
			Rows: instances,
			Pagination: types.PaginationResponse{
				Total:  len(instances),
				Page:   params.Page,
				Limit:  listOpts.Limit,
				Offset: listOpts.Offset,
			},
		},
		Success: true,
		ID:      req.ID,
	})
}
