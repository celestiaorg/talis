package routes

import (
	"github.com/celestiaorg/talis/internal/api/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes configures all the v1 routes
func RegisterRoutes(
	app *fiber.App,
	infraHandler *handlers.InfrastructureHandler,
	jobHandler *handlers.JobHandler,
) {
	// API v1 routes
	v1 := app.Group("/api/v1")

	// Infrastructure endpoints
	instances := v1.Group("/instances")
	instances.Post("/", infraHandler.CreateInfrastructure).Name("CreateInfrastructure")
	instances.Delete("/", infraHandler.DeleteInfrastructure).Name("DeleteInfrastructure")

	// Jobs endpoints
	jobs := v1.Group("/jobs")
	jobs.Get("/", jobHandler.ListJobs).Name("ListJobs")
	jobs.Get("/:id", jobHandler.GetJobStatus).Name("GetJobStatus")

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})
}
