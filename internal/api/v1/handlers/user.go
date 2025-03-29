package handlers

import (
	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service *services.User
}

func NewUserHandler(service *services.User) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput("invalid user id"))
	}

	user, err := h.service.GetUserByID(c.Context(), uint(userID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).
			JSON(infrastructure.ErrNotFound("cannot find user with id"))
	}

	return c.JSON(infrastructure.GetUserResponse{
		User: *user,
	})
}

func (h *UserHandler) GetUserByUsername(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(infrastructure.ErrInvalidInput("invalid username"))
	}
	user, err := h.service.GetUserByUsername(c.Context(), username)

	if err != nil {
		return c.Status(fiber.StatusNotFound).
			JSON(infrastructure.ErrNotFound("cannot find user with username"))
	}

	return c.JSON(infrastructure.GetUserResponse{
		User: *user,
	})
}
