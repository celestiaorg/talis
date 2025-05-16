package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	*APIHandler
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(api *APIHandler) *UserHandler {
	return &UserHandler{
		APIHandler: api,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Creates a new user via RPC
// @Tags users,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with CreateUserParams"
// @Success 200 {object} RPCResponse{data=types.CreateUserResponse} "Created user ID"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId createUser
func (h *UserHandler) CreateUser(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[CreateUserParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err = params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	user := &models.User{
		Username:     params.Username,
		Email:        params.Email,
		Role:         params.Role,
		PublicSSHKey: params.PublicSSHKey,
	}

	id, err := h.user.CreateUser(c.Context(), user)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgCreateUserFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Data: types.CreateUserResponse{
			UserID: id,
		},
		Success: true,
		ID:      req.ID,
	})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Retrieves a user by their ID via RPC
// @Tags users,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with UserGetByIDParams"
// @Success 200 {object} RPCResponse{data=models.User} "User details"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 404 {object} RPCResponse "User not found"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId getUserById
func (h *UserHandler) GetUserByID(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[UserGetByIDParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err = params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	user, err := h.user.GetUserByID(c.Context(), params.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return respondWithRPCError(c, fiber.StatusNotFound, ErrMsgUserNotFoundByID, nil, req.ID)
	} else if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgGetUserFailed, err.Error(), req.ID)
	}

	if user == nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgNilUserObject, nil, req.ID)
	}

	return c.JSON(RPCResponse{
		Data:    user,
		Success: true,
		ID:      req.ID,
	})
}

// GetUsers godoc
// @Summary Get users
// @Description Retrieves all users or a single user by username via RPC
// @Tags users,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with UserGetParams"
// @Success 200 {object} RPCResponse{data=types.UserResponse} "User list or single user"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 404 {object} RPCResponse "User not found"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId getUsers
func (h *UserHandler) GetUsers(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[UserGetParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if params.Username != "" {
		return h.getUserByUsername(c, params.Username, req)
	}

	// return all users
	page := 1
	if params.Page > 1 {
		page = params.Page
	}
	paginationOpts := getPaginationOptions(page)

	users, err := h.user.GetAllUsers(c.Context(), paginationOpts)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgGetUsersFailed, err.Error(), req.ID)
	}
	return c.JSON(RPCResponse{
		Data: types.UserResponse{
			Users: users,
			Pagination: &types.PaginationResponse{
				Total:  len(users),
				Page:   page,
				Offset: paginationOpts.Offset,
			},
		},
		Success: true,
		ID:      req.ID,
	})
}

// getUserByUsername godoc
// @Summary Get user by username
// @Description Retrieves a user by their username (internal helper method)
// @Tags users,rpc
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param request body RPCRequest true "RPC request"
// @Success 200 {object} RPCResponse{data=types.UserResponse} "User details"
// @Failure 404 {object} RPCResponse "User not found"
// @Failure 500 {object} RPCResponse "Internal server error"
func (h *UserHandler) getUserByUsername(c *fiber.Ctx, username string, req RPCRequest) error {
	user, err := h.user.GetUserByUsername(c.Context(), username)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return respondWithRPCError(c, fiber.StatusNotFound, ErrMsgUserNotFoundByUsername, nil, req.ID)
	} else if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgGetUserFailed, err.Error(), req.ID)
	}

	if user == nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgNilUserObject, nil, req.ID)
	}

	return c.JSON(RPCResponse{
		Data: types.UserResponse{
			User:       *user,
			Pagination: nil,
		},
		Success: true,
		ID:      req.ID,
	})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Deletes a user by their ID via RPC
// @Tags users,rpc
// @Accept json
// @Produce json
// @Param request body RPCRequest true "RPC request with DeleteUserParams"
// @Success 200 {object} RPCResponse "User deleted successfully"
// @Failure 400 {object} RPCResponse "Invalid parameters"
// @Failure 404 {object} RPCResponse "User not found"
// @Failure 500 {object} RPCResponse "Internal server error"
// @OperationId deleteUser
func (h *UserHandler) DeleteUser(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[DeleteUserParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err = params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	err = h.user.DeleteUser(c.Context(), params.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return respondWithRPCError(c, fiber.StatusNotFound, ErrMsgUserNotFoundByID, nil, req.ID)
	}
	if err != nil {
		return respondWithRPCError(c, fiber.StatusInternalServerError, ErrMsgDeleteUserFailed, err.Error(), req.ID)
	}

	return c.JSON(RPCResponse{
		Success: true,
		ID:      req.ID,
	})
}
