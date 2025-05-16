// Package routes defines the API routes and URL structure
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/swaggo/swag"

	_ "github.com/celestiaorg/talis/docs" // Import generated docs
)

// RegisterSwaggerRoutes registers the Swagger UI routes
func RegisterSwaggerRoutes(app *fiber.App) {
	// Serve Swagger UI HTML
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger-ui.html")
	})

	// Serve Swagger JSON
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		doc, err := swag.ReadDoc()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		return c.Type("json").Send([]byte(doc))
	})
}
