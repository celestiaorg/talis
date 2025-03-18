// Package test provides integration testing infrastructure for Talis
package test

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/repos"
)

// Version represents the current version of the test package.
// This will be updated as new features are added.
const Version = "0.1.0"

// DefaultTestTimeout is the default timeout for test environments.
const DefaultTestTimeout = 30 * time.Second

// Option represents a configuration option for the test environment.
type Option func(*TestEnvironment)

// WithTimeout returns an option that sets the test environment timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(env *TestEnvironment) {
		if env.cancelFunc != nil {
			env.cancelFunc()
		}
		env.ctx, env.cancelFunc = context.WithTimeout(context.Background(), timeout)
	}
}

// WithCleanupFunc returns an option that adds a cleanup function to be
// called when the environment is cleaned up.
func WithCleanupFunc(cleanup func()) Option {
	return func(env *TestEnvironment) {
		oldCleanup := env.cleanup
		env.cleanup = func() {
			if cleanup != nil {
				cleanup()
			}
			if oldCleanup != nil {
				oldCleanup()
			}
		}
	}
}

// WithDB returns an option that configures the test environment to use
// the provided database connection. If nil, a new in-memory database
// will be created.
func WithDB(database *gorm.DB) Option {
	return func(env *TestEnvironment) {
		if database != nil {
			env.DB = database
		} else {
			// Create new in-memory database
			dbConn, err := NewInMemoryDB()
			env.Require().NoError(err, "Failed to create in-memory database")
			env.DB = dbConn

			// Run migrations
			err = RunMigrations(env.DB)
			env.Require().NoError(err, "Failed to run database migrations")
		}

		// Initialize repositories
		env.JobRepo = repos.NewJobRepository(env.DB)
		env.InstanceRepo = repos.NewInstanceRepository(env.DB)

		// Add cleanup
		oldCleanup := env.cleanup
		env.cleanup = func() {
			if oldCleanup != nil {
				oldCleanup()
			}
			// Close database connection
			sqlDB, err := env.DB.DB()
			if err == nil && sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
	}
}
