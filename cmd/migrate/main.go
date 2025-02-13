// This file is used to run database migrations
// How to run:
// go run cmd/migrate/main.go              # Run all pending migrations
// go run cmd/migrate/main.go -down        # Rollback all migrations
// go run cmd/migrate/main.go -steps 1     # Run one migration
// go run cmd/migrate/main.go -steps -1    # Rollback one migration
// go run cmd/migrate/main.go -force 1     # Force version 1
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/celestiaorg/talis/internal/db/migrations"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
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

	var (
		dbURLFlag = flag.String("db", "", "Database URL (optional, defaults to env vars)")
		migPath   = flag.String("path", "file://migrations", "Path to migration files")
		down      = flag.Bool("down", false, "Roll back migrations")
		steps     = flag.Int("steps", 0, "Number of migrations to apply (up or down)")
		force     = flag.Int("force", -1, "Force a specific version")
		retries   = flag.Int("retries", 5, "Number of connection retries")
		retryWait = flag.Duration("retry-wait", 3*time.Second, "Wait time between retries")
	)
	flag.Parse()

	// Use command line flag if provided, otherwise use env vars
	if *dbURLFlag != "" {
		dbURL = *dbURLFlag
	}

	config := migrations.Config{
		MigrationsPath: *migPath,
		DatabaseURL:    dbURL,
		RetryAttempts:  *retries,
		RetryDelay:     *retryWait,
	}

	service, err := migrations.NewMigrationService(config)
	if err != nil {
		log.Fatalf("Failed to create migration service: %v", err)
	}

	// Handle force version
	if *force >= 0 {
		if err := service.Force(*force); err != nil {
			log.Fatalf("Failed to force version %d: %v", *force, err)
		}
		log.Printf("Successfully forced version to %d", *force)
		os.Exit(0)
	}

	// Handle steps
	if *steps != 0 {
		if err := service.Steps(*steps); err != nil {
			log.Fatalf("Failed to apply %d steps: %v", *steps, err)
		}
		log.Printf("Successfully applied %d steps", *steps)
		os.Exit(0)
	}

	// Handle up/down
	if *down {
		if err := service.Down(); err != nil {
			log.Fatalf("Migration rollback failed: %v", err)
		}
	} else {
		if err := service.Up(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	version, dirty, err := service.Version()
	if err != nil {
		log.Printf("Warning: could not get final version: %v", err)
	} else {
		log.Printf("Current migration version: %d (dirty: %v)", version, dirty)
	}
}
