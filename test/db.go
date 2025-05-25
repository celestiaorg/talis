// Package test provides utilities for setting up and running tests
package test

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
)

// NewFileBasedTestDB creates a new file-based SQLite database for testing.
// It returns the database connection and the path to the temporary directory.
func NewFileBasedTestDB() (*gorm.DB, string, error) {
	tmpDir, err := os.MkdirTemp("", "talis_test")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temporary directory: %w", err)
	}
	dbPath := filepath.Join(tmpDir, "talis_test.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		// Try to clean up the temporary directory, but don't fail if cleanup fails
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			fmt.Printf("Warning: failed to remove temporary directory after database error: %v\n", rmErr)
		}
		return nil, "", fmt.Errorf("failed to open database: %w", err)
	}
	return db, tmpDir, nil
}

// CleanupTestDB closes the database connection and removes the temporary directory.
func CleanupTestDB(db *gorm.DB, tmpDir string) {
	sqlDB, err := db.DB()
	if err == nil && sqlDB != nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			fmt.Printf("Error closing database connection: %v\n", closeErr)
		}
	}
	if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
		fmt.Printf("Error removing temporary directory: %v\n", rmErr)
	}
}

// RunMigrations runs all database migrations for the test database.
func RunMigrations(db *gorm.DB) error {
	// Run migrations for all models
	err := db.AutoMigrate(
		&models.Instance{},
		&models.User{},
		&models.Project{},
		&models.Task{},
		&models.SSHKey{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// SetupTestDB configures the test suite to use the provided database connection.
// If nil is provided, a new file-based database will be created.
func SetupTestDB(suite *Suite, database *gorm.DB) {
	if database != nil {
		suite.DB = database
	} else {
		// Create new file-based database
		dbConn, tmpDir, err := NewFileBasedTestDB()
		suite.Require().NoError(err, "Failed to create file-based database")
		suite.DB = dbConn

		// Run migrations
		err = RunMigrations(suite.DB)
		suite.Require().NoError(err, "Failed to run database migrations")

		// Add cleanup
		oldCleanup := suite.cleanup
		suite.cleanup = func() {
			if oldCleanup != nil {
				oldCleanup()
			}
			// Close database connection and remove temporary directory
			CleanupTestDB(suite.DB, tmpDir)
		}
	}

	// Initialize repositories
	suite.InstanceRepo = repos.NewInstanceRepository(suite.DB)
	suite.UserRepo = repos.NewUserRepository(suite.DB)
	suite.ProjectRepo = repos.NewProjectRepository(suite.DB)
	suite.TaskRepo = repos.NewTaskRepository(suite.DB)
}
