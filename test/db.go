package test

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// NewInMemoryDB creates a new in-memory SQLite database for testing.
func NewInMemoryDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create in-memory database: %w", err)
	}
	return db, nil
}

// RunMigrations runs all database migrations for the test database.
func RunMigrations(db *gorm.DB) error {
	// Run migrations for all models
	err := db.AutoMigrate(
		&models.Job{},
		&models.Instance{},
		// Add other models as needed
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
