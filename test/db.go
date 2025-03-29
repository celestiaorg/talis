package test

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
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

// SetupTestDB configures the test suite to use the provided database connection.
// If nil is provided, a new in-memory database will be created.
func SetupTestDB(suite *TestSuite, database *gorm.DB) {
	if database != nil {
		suite.DB = database
	} else {
		// Create new in-memory database
		dbConn, err := NewInMemoryDB()
		suite.Require().NoError(err, "Failed to create in-memory database")
		suite.DB = dbConn

		// Run migrations
		err = RunMigrations(suite.DB)
		suite.Require().NoError(err, "Failed to run database migrations")
	}

	// Initialize repositories
	suite.JobRepo = repos.NewJobRepository(suite.DB)
	suite.InstanceRepo = repos.NewInstanceRepository(suite.DB)
	suite.UserRepo = repos.NewUserRepository(suite.DB)

	// Add cleanup
	oldCleanup := suite.cleanup
	suite.cleanup = func() {
		if oldCleanup != nil {
			oldCleanup()
		}
		// Close database connection
		sqlDB, err := suite.DB.DB()
		if err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}
