package v1

import (
	"github.com/tty47/talis/internal/api/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all the v1 routes
func SetupRoutes(router fiber.Router) {
	// User routes
	users := router.Group("/users")
	users.Get("/", handlers.GetUsers)
	users.Post("/login", handlers.Login)

	// Infrastructure routes - single endpoint for both create and delete
	instances := router.Group("/instances")
	instances.Post("/", handlers.HandleInfrastructure)
}

// Register registers the v1 routes
func Register(app *fiber.App) {
	// API v1 routes
	v1Group := app.Group("/api/v1")
	SetupRoutes(v1Group)
}
