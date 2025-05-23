// Package routes defines the API routes and URL structure
package routes

import (
	"embed"
	"io/fs"

	"github.com/gofiber/fiber/v2"
	"github.com/swaggo/swag"

	_ "github.com/celestiaorg/talis/docs/swagger" // Import generated docs
)

//go:embed swagger-ui.html
var swaggerUI embed.FS

// RegisterSwaggerRoutes registers the Swagger UI routes
func RegisterSwaggerRoutes(app *fiber.App) {
	// Create a sub-filesystem for swagger-ui.html
	swaggerFS, err := fs.Sub(swaggerUI, ".")
	if err != nil {
		panic(err)
	}

	// Serve Swagger UI HTML
	app.Get("/swagger", func(c *fiber.Ctx) error {
		content, err := fs.ReadFile(swaggerFS, "swagger-ui.html")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		c.Set("Content-Type", "text/html")
		return c.Send(content)
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
