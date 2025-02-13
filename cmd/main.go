package main

import (
	"database/sql"
	"fmt"
	"os"

	fiberlog "github.com/gofiber/fiber/v2/log"

	"github.com/gofiber/fiber/v2"

	"github.com/joho/godotenv"

	"github.com/celestiaorg/talis/internal/api/v1/handlers"
	"github.com/celestiaorg/talis/internal/api/v1/middleware"
	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/application/job"
	"github.com/celestiaorg/talis/internal/db/migrations"
	"github.com/celestiaorg/talis/internal/infrastructure/persistence/postgres"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fiberlog.Fatal("Error loading .env file")
	}

	// Build database URL from env vars
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Configure logger
	fiberlog.SetLevel(fiberlog.LevelInfo)

	// Initialize database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fiberlog.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	config := migrations.DefaultConfig()
	config.DatabaseURL = dbURL
	migrationService, err := migrations.NewMigrationService(config)
	if err != nil {
		fiberlog.Fatalf("Failed to create migration service: %v", err)
	}
	if err := migrationService.Up(); err != nil {
		fiberlog.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	jobRepo := postgres.NewJobRepository(db)

	// Initialize services
	jobService := job.NewService(jobRepo)

	// Initialize handlers
	infraHandler := handlers.NewInfrastructureHandler(jobService)
	jobHandler := handlers.NewJobHandler(jobService)

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: customErrorHandler,
	})

	// Add logger middleware
	app.Use(middleware.Logger())

	// Register routes
	routes.RegisterRoutes(app, infraHandler, jobHandler)

	// Start server
	fiberlog.Info("Server starting on :8080")
	fiberlog.Fatal(app.Listen(":8080"))
}

// customErrorHandler is a custom error handler for the Fiber app
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
