package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
)

func TestNewTestEnvironment(t *testing.T) {
	// Create environment
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Basic environment checks
	assert.NotNil(t, env.t, "testing.T should be set")
	assert.NotNil(t, env.ctx, "context should be set")
	assert.NotNil(t, env.cancelFunc, "cancel function should be set")
	assert.NotNil(t, env.cleanup, "cleanup function should be set")
}

func TestTestEnvironment_Database(t *testing.T) {
	t.Run("default in-memory database", func(t *testing.T) {
		env := NewTestEnvironment(t)
		defer env.Cleanup()

		// Check database components
		require.NotNil(t, env.DB, "database should be initialized")
		require.NotNil(t, env.JobRepo, "job repository should be initialized")
		require.NotNil(t, env.InstanceRepo, "instance repository should be initialized")

		// Verify database is working
		job := &models.Job{
			Name:        "test-job",
			ProjectName: "test-project",
			OwnerID:     1, // Set owner ID for the test
		}
		result := env.DB.Create(job)
		assert.NoError(t, result.Error, "should create job without error")
		assert.NotZero(t, job.ID, "job should have an ID")

		// Verify repositories are working
		savedJob, err := env.JobRepo.GetByID(env.Context(), job.OwnerID, job.ID)
		assert.NoError(t, err, "should get job without error")
		assert.Equal(t, job.Name, savedJob.Name, "job names should match")
	})

	t.Run("custom database connection", func(t *testing.T) {
		// Create a custom database
		customDB, err := NewInMemoryDB()
		require.NoError(t, err, "should create custom database")
		err = RunMigrations(customDB)
		require.NoError(t, err, "should run migrations")

		// Create environment with custom database
		env := NewTestEnvironment(t, WithDB(customDB))
		defer env.Cleanup()

		assert.Same(t, customDB, env.DB, "should use provided database")
	})
}

func TestTestEnvironment_Context(t *testing.T) {
	t.Run("default timeout", func(t *testing.T) {
		env := NewTestEnvironment(t)
		defer env.Cleanup()

		deadline, ok := env.Context().Deadline()
		require.True(t, ok, "context should have deadline")
		assert.True(t, deadline.After(time.Now()), "deadline should be in the future")
	})

	t.Run("custom timeout", func(t *testing.T) {
		customTimeout := 5 * time.Second
		env := NewTestEnvironment(t, WithTimeout(customTimeout))
		defer env.Cleanup()

		deadline, ok := env.Context().Deadline()
		require.True(t, ok, "context should have deadline")
		expectedDeadline := time.Now().Add(customTimeout)
		assert.WithinDuration(t, expectedDeadline, deadline, time.Second)
	})

	t.Run("context cancellation", func(t *testing.T) {
		env := NewTestEnvironment(t)

		// Get a channel that will be closed when the context is done
		done := make(chan struct{})
		go func() {
			<-env.Context().Done()
			close(done)
		}()

		// Cleanup should cancel the context
		env.Cleanup()

		// Wait for context cancellation or timeout
		select {
		case <-done:
			// Context was cancelled as expected
		case <-time.After(time.Second):
			t.Error("context was not cancelled by cleanup")
		}
	})
}

func TestTestEnvironment_Cleanup(t *testing.T) {
	t.Run("multiple cleanup calls", func(t *testing.T) {
		env := NewTestEnvironment(t)

		// First cleanup should work
		env.Cleanup()

		// Second cleanup should not panic
		env.Cleanup()
	})

	t.Run("custom cleanup function", func(t *testing.T) {
		cleanupCalled := false
		customCleanup := func() {
			cleanupCalled = true
		}

		env := NewTestEnvironment(t, WithCleanupFunc(customCleanup))
		env.Cleanup()

		assert.True(t, cleanupCalled, "custom cleanup should be called")
	})

	t.Run("database cleanup", func(t *testing.T) {
		env := NewTestEnvironment(t)

		// Create a test record
		job := &models.Job{Name: "cleanup-test"}
		env.DB.Create(job)

		// Get the underlying sql.DB
		sqlDB, err := env.DB.DB()
		require.NoError(t, err)

		// Cleanup should close the connection
		env.Cleanup()

		// Verify connection is closed
		err = sqlDB.Ping()
		assert.Error(t, err, "database connection should be closed")
	})
}

func TestTestEnvironment_WithTimeout(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Create a child context with timeout
	timeout := 100 * time.Millisecond
	ctx, cancel := env.WithTimeout(timeout)
	defer cancel()

	// Wait for timeout
	<-ctx.Done()

	assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
}

func TestTestEnvironment_Require(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// Require() should return a require.Assertions
	assert.IsType(t, &require.Assertions{}, env.Require())
}

func TestTestEnvironment_T(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Cleanup()

	// T() should return the original testing.T
	assert.Same(t, t, env.T())
}
