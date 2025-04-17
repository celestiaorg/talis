// Package main provides the entry point for the server application
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

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

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a WaitGroup to wait for goroutines to finish
	var wg sync.WaitGroup

	// Launch worker with the cancellable context and WaitGroup
	wg.Add(1) // Increment counter before launching goroutine
	go services.LaunchWorker(ctx, &wg, taskService)

	// Start server in a goroutine so that it doesn't block.
	go func() {
		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}
		log.Info("Server starting on :" + port)
		if err := app.Listen(":" + port); err != nil {
			// Use Errorf here as Fatalf would exit the program immediately
			log.Errorf("Failed to start server: %v", err)
			// We might want to signal an error to the main goroutine here
			// For now, just logging is okay, but the shutdown might not be clean if Listen fails.
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")

	// Perform application shutdown with a timeout.
	// TODO: Add a mechanism here if needed to wait for LaunchWorker to fully complete its cleanup.
	// This might involve a WaitGroup or a channel signal from the worker.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 5 second timeout for shutdown
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Errorf("Server shutdown failed: %v", err) // Use Errorf instead of Fatalf
	} else {
		log.Info("Server shut down gracefully")
	}

	// Wait for background goroutines to finish.
	log.Info("Waiting for background processes to shut down...")
	wg.Wait() // Block until wg.Done() is called by all goroutines

	log.Info("Shut down successfully. Exiting.")
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
