package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/internal/types"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	service *services.User
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(service *services.User) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// CreateUser handles the creation of a new user
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

	id, err := h.service.CreateUser(c.Context(), user)
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

// GetUserByID retrieves a user by their ID
func (h *UserHandler) GetUserByID(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[UserGetByIDParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err = params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	user, err := h.service.GetUserByID(c.Context(), params.ID)
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

// GetUsers retrieves all users or a single user if username is provided
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

	users, err := h.service.GetAllUsers(c.Context(), paginationOpts)
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

// getUserByUsername handles fetching a single user by username
func (h *UserHandler) getUserByUsername(c *fiber.Ctx, username string, req RPCRequest) error {
	user, err := h.service.GetUserByUsername(c.Context(), username)

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

// DeleteUser handles the request to terminate a user
func (h *UserHandler) DeleteUser(c *fiber.Ctx, req RPCRequest) error {
	params, err := parseParams[DeleteUserParams](req)
	if err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, ErrMsgInvalidParams, err.Error(), req.ID)
	}

	if err = params.Validate(); err != nil {
		return respondWithRPCError(c, fiber.StatusBadRequest, err.Error(), nil, req.ID)
	}

	err = h.service.DeleteUser(c.Context(), params.ID)
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
