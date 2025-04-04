package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
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
	var userReq infrastructure.CreateUserRequest
	if err := c.BodyParser(&userReq); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput(err.Error()))
	}

	if err := userReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput(err.Error()))
	}

	id, err := h.service.CreateUser(c.Context(), &userReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusCreated).
		JSON(infrastructure.CreateUserResponse{
			UserId: id,
		})
}

// GetUserByID retrieves a user by their ID
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput(ErrInvalidUserID.Error()))
	}

	user, err := h.service.GetUserByID(c.Context(), uint(userID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).
			JSON(infrastructure.ErrNotFound(ErrUserNotFoundByID.Error()))
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.JSON(infrastructure.UserResponse{
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
			JSON(infrastructure.ErrServer(err.Error()))
	}
	return c.JSON(infrastructure.UserResponse{
		Users: users,
		Pagination: infrastructure.PaginationResponse{
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
			JSON(infrastructure.ErrNotFound(ErrUserNotFoundByName.Error()))
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.JSON(infrastructure.UserResponse{
		User: *user,
	})
}

// DeleteUser handles the request to terminate a user
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput(ErrInvalidUserID.Error()))
	}

	err = h.service.DeleteUser(c.Context(), uint(userID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).
			JSON(infrastructure.ErrNotFound(ErrUserNotFoundByID.Error()))
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(infrastructure.ErrServer(err.Error()))
	}

	return c.Status(fiber.StatusOK).
		JSON(infrastructure.Success("user deleted successfully"))
}
