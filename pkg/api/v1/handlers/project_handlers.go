// Package handlers provides HTTP request handling
package handlers

import (
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"

	fiber "github.com/gofiber/fiber/v2"
)

// ProjectHandlers contains all project related handlers
type ProjectHandlers struct{}

// CreateProject handles creating a project
func CreateProject(h *RPCHandler, c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[ProjectCreateParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	project := models.Project{
		OwnerID:     ownerID,
		Name:        params.Name,
		Description: params.Description,
		Config:      params.Config,
	}

	if err := h.projectService.Create(c.Context(), &project); err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgProjCreateFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data:    project,
		Success: true,
		ID:      req.ID,
	})
}

// GetProject handles retrieving a project by name
func GetProject(h *RPCHandler, c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[ProjectGetParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	project, err := h.projectService.GetByName(c.Context(), ownerID, params.Name)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusNotFound, ErrMsgProjNotFound, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data:    project,
		Success: true,
		ID:      req.ID,
	})
}

// ListProjects handles listing all projects with pagination
func ListProjects(h *RPCHandler, c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	page := 1

	if req.Params != nil {
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
	}

	listOpts := getPaginationOptions(page)

	projects, err := h.projectService.List(c.Context(), ownerID, listOpts)
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

// DeleteProject handles deleting a project by name
func DeleteProject(h *RPCHandler, c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[ProjectDeleteParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	if err := h.projectService.Delete(c.Context(), ownerID, params.Name); err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgProjDeleteFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Success: true,
		ID:      req.ID,
	})
}

// ListProjectInstances handles listing all instances for a project
func ListProjectInstances(h *RPCHandler, c *fiber.Ctx, ownerID uint, req RPCRequest) error {
	params, err := parseParams[ProjectListInstancesParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	listOpts := getPaginationOptions(params.Page)
	instances, err := h.projectService.ListInstances(c.Context(), ownerID, params.Name, listOpts)
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
