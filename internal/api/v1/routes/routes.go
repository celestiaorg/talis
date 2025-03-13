package routes

import (
	"github.com/celestiaorg/talis/internal/api/v1/handlers"

	fiber "github.com/gofiber/fiber/v2"
)

// RegisterRoutes configures all the v1 routes
func RegisterRoutes(
	app *fiber.App,
	instanceHandler *handlers.InstanceHandler,
	jobHandler *handlers.JobHandler,
) {
	// API v1 routes
	v1 := app.Group("/api/v1")

	// Instances endpoints
	instances := v1.Group("/instances")
	instances.Get("/", instanceHandler.ListInstances).Name("ListInstances")
	instances.Post("/", instanceHandler.CreateInstance).Name("CreateInstance")
	instances.Get("/:id", instanceHandler.GetInstance).Name("GetInstance")

	// Jobs endpoints
	jobs := v1.Group("/jobs")
	jobs.Get("/", jobHandler.ListJobs).Name("ListJobs")
	jobs.Get("/:id", jobHandler.GetJobStatus).Name("GetJobStatus")
	jobs.Post("/", jobHandler.CreateJob).Name("CreateJob")
	jobs.Delete("/:id", jobHandler.TerminateJob).Name("TerminateJob")
	jobs.Put("/:id", jobHandler.UpdateJob).Name("UpdateJob")
	jobs.Get("/search", jobHandler.SearchJobs).Name("SearchJobs")

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})
}
