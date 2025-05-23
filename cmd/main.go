// Package main provides the entry point for the server application
// @title Talis API
// @version 1.0
// @description API for Talis - Web3 infrastructure management service
// @host localhost:8000
// @BasePath /talis/api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
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

	"github.com/celestiaorg/talis/docs/swagger"
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

	// Set dynamic Swagger host from environment variable
	if apiHost := os.Getenv("API_HOST"); apiHost != "" {
		swagger.SwaggerInfo.Host = apiHost
		swagger.SwaggerInfo.BasePath = "/talis/api/v1" // Ensure base path is set correctly even with dynamic host
		log.Infof("Swagger host set dynamically to: %s", apiHost)
	} else {
		log.Info("API_HOST environment variable not set, using default Swagger host from annotations.")
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
	apiHandler := handlers.NewAPIHandler(instanceService, projectService, taskService, userService)
	instanceHandler := handlers.NewInstanceHandler(apiHandler)
	projectHandler := handlers.NewProjectHandlers(apiHandler)
	taskHandler := handlers.NewTaskHandlers(apiHandler)
	userHandler := handlers.NewUserHandler(apiHandler)

	// Create RPC handler and assign handlers directly
	rpcHandler := &handlers.RPCHandler{
		ProjectHandlers: projectHandler,
		TaskHandlers:    taskHandler,
		UserHandlers:    userHandler,
	}

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Add logger for API requests
	app.Use(log.APILogger())

	// Register routes - no need for project and task handlers as they're handled via RPC
	// The above comment is no longer entirely true as ListByInstanceID is a direct REST endpoint on TaskHandler
	routes.RegisterRoutes(app, instanceHandler, rpcHandler, taskHandler)

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a WaitGroup to wait for goroutines to finish
	var wg sync.WaitGroup

	// Get worker count from environment or use default
	workerCount := services.DefaultWorkerCount
	if workerCountStr := os.Getenv("WORKER_COUNT"); workerCountStr != "" {
		if count, err := strconv.Atoi(workerCountStr); err == nil && count > 0 {
			workerCount = count
			log.Infof("Using configured worker count: %d", workerCount)
		} else if err != nil {
			log.Warnf("Invalid WORKER_COUNT value: %s, using default: %d", workerCountStr, workerCount)
		}
	}

	// Get high priority ratio from environment or use default
	highPriorityRatio := services.DefaultHighPriorityRatio
	if ratioStr := os.Getenv("HIGH_PRIORITY_RATIO"); ratioStr != "" {
		if ratio, err := strconv.ParseFloat(ratioStr, 64); err == nil && ratio > 0 && ratio <= 1.0 {
			highPriorityRatio = ratio
			log.Infof("Using configured high priority ratio: %.2f", highPriorityRatio)
		} else if err != nil {
			log.Warnf("Invalid HIGH_PRIORITY_RATIO value: %s, using default: %.2f", ratioStr, highPriorityRatio)
		} else {
			log.Warnf("HIGH_PRIORITY_RATIO must be between 0 and 1, using default: %.2f", highPriorityRatio)
		}
	}

	// Launch worker pool with the cancellable context and WaitGroup
	wg.Add(1) // Increment counter before launching goroutine
	workerPool := services.NewWorkerPool(instanceService, projectService, taskService, userService, services.DefaultBackoff)
	workerPool.WithWorkerCount(workerCount).WithHighPriorityRatio(highPriorityRatio)

	// Recover any stale tasks before starting the worker pool
	log.Info("Starting worker pool...")
	go workerPool.LaunchWorkerPool(ctx, &wg)

	// Start server in a goroutine so that it doesn't block.
	var errChan = make(chan error)
	go func() {
		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}
		log.Info("Server starting on :" + port)
		if err := app.Listen(":" + port); err != nil {
			// Send error to the channel
			errChan <- err
		}
	}()

	// Listen for the interrupt signal or the server error.
	select {
	case err := <-errChan:
		log.Errorf("Server failed to start: %v", err)
	case <-ctx.Done():
		log.Info("Received interrupt signal, shutting down gracefully")
	}

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Info("Shutting down gracefully, press Ctrl+C again to force")

	// Perform application shutdown with a timeout.
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
