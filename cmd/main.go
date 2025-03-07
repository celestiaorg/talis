package main

import (
	"os"
	"strconv"

	fiberlog "github.com/gofiber/fiber/v2/log"

	"github.com/gofiber/fiber/v2"

	"github.com/joho/godotenv"

	"errors"

	"github.com/celestiaorg/talis/internal/api/v1/handlers"
	"github.com/celestiaorg/talis/internal/api/v1/middleware"
	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db"
	"github.com/celestiaorg/talis/internal/db/repos"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fiberlog.Fatal("Error loading .env file")
	}

	// Configure logger
	fiberlog.SetLevel(fiberlog.LevelInfo)

	// This is temporary, we will pass them through the CLI later
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		fiberlog.Fatalf("Failed to convert DB_PORT to int: %v", err)
	}

	// Initialize database
	DB, err := db.New(db.Options{
		Host:     os.Getenv("DB_HOST"),
		Port:     dbPort,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		// SSLEnabled: os.Getenv("DB_SSL_MODE") == "true",
	})
	if err != nil {
		fiberlog.Fatalf("Failed to connect to database: %v", err)
	}
	// We will use connection pooling later
	// defer DB.Close()

	// Initialize repositories
	jobRepo := repos.NewJobRepository(DB)
	instanceRepo := repos.NewInstanceRepository(DB)

	// Initialize services
	jobService := services.NewJobService(jobRepo)
	instanceService := services.NewInstanceService(instanceRepo, jobService)

	// Initialize handlers
	instanceHandler := handlers.NewInstanceHandler(instanceService, jobService)
	jobHandler := handlers.NewJobHandler(jobService)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Add logger middleware
	app.Use(middleware.Logger())

	// Register routes
	routes.RegisterRoutes(app, instanceHandler, jobHandler)

	// Error handler
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				return c.Status(fiberErr.Code).JSON(fiber.Map{
					"error": fiberErr.Message,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return nil
	})

	// Start server
	fiberlog.Info("Server starting on :8080")
	fiberlog.Fatal(app.Listen(":8080"))
}

// customErrorHandler is a custom error handler for the Fiber app
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Default error
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
