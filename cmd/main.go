package main

import (
	"errors"
	"os"
	"strconv"

	fiber "github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/log"
	"github.com/joho/godotenv"

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
	jobService := services.NewJobService(jobRepo, instanceRepo)
	instanceService := services.NewInstanceService(instanceRepo, jobService)

	// Initialize handlers
	instanceHandler := handlers.NewInstanceHandler(instanceService)
	jobHandler := handlers.NewJobHandler(jobService, instanceService)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Add logger middleware
	app.Use(middleware.Logger())

	// Register routes
	routes.RegisterRoutes(app, instanceHandler, jobHandler)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	fiberlog.Info("Server starting on :" + port)
	if err := app.Listen(":" + port); err != nil {
		fiberlog.Fatalf("Failed to start server: %v", err)
	}
}

// customErrorHandler handles errors returned by the handlers
func customErrorHandler(c *fiber.Ctx, err error) error {
	// Default error
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a fiber.*Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		message = e.Message
	}

	// Return JSON response
	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
