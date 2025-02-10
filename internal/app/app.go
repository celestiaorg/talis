package app

import (
	v1 "github.com/tty47/talis/internal/api/v1/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func NewApp() *fiber.App {
	app := fiber.New()

	// Middleware (e.g., logging, CORS)
	app.Use(logger.New())

	// Register versioned routes
	v1.Register(app)

	return app
}
