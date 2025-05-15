// Package handlers provides HTTP request handling
package handlers

import (
	"errors"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
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

// Create godoc
// @Summary Create a new project
// @Description Creates a new project via RPC
// @Tags projects,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with ProjectCreateParams"
// @Success 200 {object} RPCResponse{data=models.Project} "Created project"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId createProject
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

// Get godoc
// @Summary Get project by name
// @Description Retrieves a project by its name via RPC
// @Tags projects,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with ProjectGetParams"
// @Success 200 {object} RPCResponse{data=models.Project} "Project details"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 404 {object} RPCResponse "Project not found"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId getProjectByName
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

// List godoc
// @Summary List projects
// @Description Returns a list of projects with pagination via RPC
// @Tags projects,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with ProjectListParams"
// @Success 200 {object} RPCResponse{data=types.ListResponse[models.Project]} "List of projects"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId listProjects
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

// Delete godoc
// @Summary Delete a project
// @Description Deletes a project by its name via RPC
// @Tags projects,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with ProjectDeleteParams"
// @Success 200 {object} RPCResponse "Project deleted successfully"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId deleteProject
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

// ListInstances godoc
// @Summary List instances for a project
// @Description Returns a list of instances for a specific project via RPC
// @Tags projects,rpc,instances
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with ProjectListInstancesParams"
// @Success 200 {object} RPCResponse{data=types.ListResponse[models.Instance]} "List of instances"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId listProjectInstances
func (h *ProjectHandlers) ListInstances(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[ProjectListInstancesParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err := params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	listOpts := getPaginationOptions(params.Page)
	instances, err := h.project.ListInstances(c.Context(), params.OwnerID, params.Name, listOpts)
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
