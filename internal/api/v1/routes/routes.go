package routes

import (
	"github.com/celestiaorg/talis/internal/api/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes configures all the v1 routes
func RegisterRoutes(
	app *fiber.App,
	instanceHandler *handlers.InstanceHandler,
	jobHandler *handlers.JobHandler,
) {
	// API v1 routes
	v1 := app.Group("/api/v1")

	// ---------------------------
	// Jobs endpoints
	// ---------------------------
	jobs := v1.Group("/jobs")
	jobs.Get("/:id", jobHandler.GetJobStatus).Name("GetJobStatus")
	jobs.Post("/", jobHandler.CreateJob).Name("CreateJob")

	// Job instances endpoints
	jobInstances := jobs.Group("/:jobId/instances")
	jobInstances.Get("/", instanceHandler.GetInstancesByJobID).Name("GetInstancesByJobID")
	jobInstances.Get("/public-ips", instanceHandler.GetPublicIPs).Name("GetPublicIPs")
	jobInstances.Get("/:instanceId", instanceHandler.GetInstance).Name("GetInstance")
	jobInstances.Post("/", instanceHandler.CreateInstance).Name("CreateInstance")
	jobInstances.Delete("/", instanceHandler.DeleteInstance).Name("DeleteInstance")

	// Admin endpoints for instances (all jobs)
	instances := v1.Group("/instances")
	instances.Get("/", instanceHandler.ListInstances).Name("ListInstances")
	instances.Get("/all-metadata", instanceHandler.GetAllMetadata).Name("GetAllMetadata")

	// ---------------------------
	// Health check
	// ---------------------------
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})
}
