package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

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

// Define common errors
var (
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrUserNotFoundByID   = errors.New("user not found with provided id")
	ErrUserNotFoundByName = errors.New("user not found with provided username")
)

// CreateUser handles the creation of a new user
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var userReq types.CreateUserRequest
	if err := c.BodyParser(&userReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(err.Error()))
	}

	if err := userReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(err.Error()))
	}

	id, err := h.service.CreateUser(c.Context(), &userReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).
		JSON(types.CreateUserResponse{
			UserID: id,
		})
}

// GetUserByID retrieves a user by their ID
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(ErrInvalidUserID.Error()))
	}

	if userID <= 0 {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput("user ID must be positive"))
	}

	user, err := h.service.GetUserByID(c.Context(), uint(userID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).
			JSON(types.ErrNotFound(ErrUserNotFoundByID.Error()))
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(err.Error()))
	}

	return c.JSON(types.UserResponse{
		User: *user,
	})
}

// GetUsers retrieves all users or a single user if username is provided
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	username := c.Query("username")

	if username != "" {
		return h.getUserByUsername(c, username)
	}

	// return all users
	page := c.QueryInt("page", 1)
	paginationOpts := getPaginationOptions(page)

	users, err := h.service.GetAllUsers(c.Context(), paginationOpts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(err.Error()))
	}
	return c.JSON(types.UserResponse{
		Users: users,
		Pagination: &types.PaginationResponse{
			Total:  len(users),
			Page:   page,
			Offset: paginationOpts.Offset,
		},
	})
}

// getUserByUsername handles fetching a single user by username
func (h *UserHandler) getUserByUsername(c *fiber.Ctx, username string) error {
	user, err := h.service.GetUserByUsername(c.Context(), username)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).
			JSON(types.ErrNotFound(ErrUserNotFoundByName.Error()))
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(err.Error()))
	}

	return c.JSON(types.UserResponse{
		User:       *user,
		Pagination: nil,
	})
}

// DeleteUser handles the request to terminate a user
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput(ErrInvalidUserID.Error()))
	}

	if userID <= 0 {
		return c.Status(fiber.StatusBadRequest).
			JSON(types.ErrInvalidInput("user ID must be positive"))
	}

	err = h.service.DeleteUser(c.Context(), uint(userID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).
			JSON(types.ErrNotFound(ErrUserNotFoundByID.Error()))
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(types.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusOK).
		JSON(types.Success("user deleted successfully"))
}
