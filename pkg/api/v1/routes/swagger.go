package routes

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterSwaggerRoutes adds Swagger documentation routes to the app
func RegisterSwaggerRoutes(app *fiber.App) {
	// Serve Swagger UI HTML
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger-ui.html")
	})

	// Serve Swagger JSON
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// Serve Swagger YAML
	app.Get("/swagger/doc.yaml", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.yaml")
	})
}
