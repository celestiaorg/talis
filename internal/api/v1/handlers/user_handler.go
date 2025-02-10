package handlers

import "github.com/gofiber/fiber/v2"

// GetUsers handles the retrieval of users
func GetUsers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "List of users",
	})
}

// Login handles user authentication
func Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// TODO: Implement actual authentication logic
	return c.JSON(fiber.Map{
		"message": "Logged in successfully",
		"token":   "dummy-token",
	})
}
