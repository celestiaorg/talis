package main

import (
	"errors"
	"os"
	"strconv"

	log "github.com/celestiaorg/talis/internal/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/celestiaorg/talis/internal/api/v1/handlers"
	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/db"
	"github.com/celestiaorg/talis/internal/db/repos"
)

func main() {
	// Configure logger
	log.InitializeAndConfigure()

	// Log that the application is starting
	log.Info("Starting application...")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// This is temporary, we will pass them through the CLI later
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatalf("Failed to convert DB_PORT to int: %v", err)
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
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// We will use connection pooling later
	// defer DB.Close()

	// Initialize repositories
	jobRepo := repos.NewJobRepository(DB)
	instanceRepo := repos.NewInstanceRepository(DB)
	userRepo := repos.NewUserRepository(DB)

	// Initialize services
	jobService := services.NewJobService(jobRepo, instanceRepo)
	instanceService := services.NewInstanceService(instanceRepo, jobService)
	userService := services.NewUserService(userRepo)

	// Initialize handlers
	instanceHandler := handlers.NewInstanceHandler(instanceService)
	jobHandler := handlers.NewJobHandler(jobService, instanceService)
	userHandler := handlers.NewUserHandler(userService)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Add logger for API requests
	app.Use(log.APILogger())

	// Register routes
	routes.RegisterRoutes(app, instanceHandler, jobHandler, userHandler)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("Server starting on :" + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
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
