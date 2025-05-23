// Package routes defines the API routes and URL structure
package routes

import (
	"path/filepath"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/swaggo/swag"

	_ "github.com/celestiaorg/talis/docs/swagger" // Import generated docs
)

// RegisterSwaggerRoutes registers the Swagger UI routes
func RegisterSwaggerRoutes(app *fiber.App) {
	// Get the absolute path to the swagger-ui.html file
	_, b, _, _ := runtime.Caller(0)
	swaggerPath := filepath.Join(filepath.Dir(b), "swagger-ui.html")

	// Serve Swagger UI HTML
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.SendFile(swaggerPath)
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
