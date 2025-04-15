// Package main provides the entry point for the server application
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/celestiaorg/talis/internal/db"
	"github.com/celestiaorg/talis/internal/db/repos"
	log "github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/pkg/api/v1/routes"
)

func main() {
	// Load .env file first
	if err := godotenv.Load(); err != nil {
		// Use fmt.Printf here since logger isn't initialized yet
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	// Configure logger after loading .env
	log.InitializeAndConfigure()

	// Log that the application is starting
	log.Info("Starting application...")

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
	instanceRepo := repos.NewInstanceRepository(DB)
	userRepo := repos.NewUserRepository(DB)
	projectRepo := repos.NewProjectRepository(DB)
	taskRepo := repos.NewTaskRepository(DB)

	// Initialize services
	projectService := services.NewProjectService(projectRepo)
	taskService := services.NewTaskService(taskRepo, projectService)
	instanceService := services.NewInstanceService(instanceRepo, taskService, projectService)
	userService := services.NewUserService(userRepo)

	// Initialize handlers
	instanceHandler := handlers.NewInstanceHandler(instanceService)
	userHandler := handlers.NewUserHandler(userService)

	// Create RPC handler and assign handlers directly
	rpcHandler := &handlers.RPCHandler{
		ProjectHandlers: handlers.NewProjectHandlers(projectService),
		TaskHandlers:    handlers.NewTaskHandlers(taskService),
	}

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Add logger for API requests
	app.Use(log.APILogger())

	// Register routes - no need for project and task handlers as they're handled via RPC
	routes.RegisterRoutes(app, instanceHandler, userHandler, rpcHandler)

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
