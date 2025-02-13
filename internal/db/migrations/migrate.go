package migrations

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Config holds migration configuration
type Config struct {
	MigrationsPath string
	DatabaseURL    string
	RetryAttempts  int
	RetryDelay     time.Duration
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		MigrationsPath: "file://migrations",
		RetryAttempts:  5,
		RetryDelay:     time.Second * 3,
	}
}

// MigrationService handles database migrations
type MigrationService struct {
	config  Config
	migrate *migrate.Migrate
}

// NewMigrationService creates a new migration service
func NewMigrationService(config Config) (*MigrationService, error) {
	var m *migrate.Migrate
	var err error

	// Retry connection a few times before giving up
	for i := 0; i < config.RetryAttempts; i++ {
		m, err = migrate.New(config.MigrationsPath, config.DatabaseURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database, attempt %d/%d: %v", i+1, config.RetryAttempts, err)
		time.Sleep(config.RetryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance after %d attempts: %v", config.RetryAttempts, err)
	}

	return &MigrationService{
		config:  config,
		migrate: m,
	}, nil
}

// Up runs all pending migrations
func (s *MigrationService) Up() error {
	if err := s.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")
	return nil
}

// Down rolls back all migrations
func (s *MigrationService) Down() error {
	if err := s.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %v", err)
	}
	log.Println("Rollback completed successfully")
	return nil
}

// Steps runs n migrations up or down
func (s *MigrationService) Steps(n int) error {
	if err := s.migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d migrations: %v", n, err)
	}
	return nil
}

// Version returns the current migration version
func (s *MigrationService) Version() (uint, bool, error) {
	return s.migrate.Version()
}

// Force forces a specific version
func (s *MigrationService) Force(version int) error {
	return s.migrate.Force(version)
}
