package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
)

func TestNewTestSuite(t *testing.T) {
	// Create environment
	env := NewTestSuite(t)
	defer env.Cleanup()

	// Basic environment checks
	assert.NotNil(t, env.t, "testing.T should be set")
	assert.Same(t, t, env.t)
	assert.NotNil(t, env.App, "app should be initialized")
	assert.NotNil(t, env.Server, "server should be initialized")
	assert.NotNil(t, env.APIClient, "API client should be initialized")
	assert.NotNil(t, env.DB, "database should be initialized")
	assert.NotNil(t, env.JobRepo, "job repository should be initialized")
	assert.NotNil(t, env.InstanceRepo, "instance repository should be initialized")
	assert.NotNil(t, env.MockDOClient, "mock DO client should be initialized")
	assert.NotNil(t, env.ctx, "context should be set")
	assert.NotNil(t, env.cancelFunc, "cancel function should be set")
	assert.NotNil(t, env.cleanup, "cleanup function should be set")
}

func TestTestEnvironment_Database(t *testing.T) {
	t.Run("default in-memory database", func(t *testing.T) {
		env := NewTestSuite(t)
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

		// Verify job repository is working
		savedJob, err := env.JobRepo.GetByID(env.ctx, job.OwnerID, job.ID)
		assert.NoError(t, err, "should get job without error")
		assert.Equal(t, job.Name, savedJob.Name, "job names should match")

		// Verify instance repository is working
		instance := &models.Instance{
			Name:    "test-instance",
			JobID:   job.ID,
			OwnerID: job.OwnerID,
			Status:  models.InstanceStatusPending,
		}
		result = env.DB.Create(instance)
		assert.NoError(t, result.Error, "should create instance without error")
		assert.NotZero(t, instance.ID, "instance should have an ID")

		savedInstance, err := env.InstanceRepo.GetByID(env.ctx, instance.OwnerID, instance.JobID, instance.ID)
		assert.NoError(t, err, "should get instance without error")
		assert.Equal(t, instance.Name, savedInstance.Name, "instance names should match")
	})
}

func TestTestEnvironment_Cleanup(t *testing.T) {
	t.Run("multiple cleanup calls", func(t *testing.T) {
		env := NewTestSuite(t)

		// First cleanup should work
		env.Cleanup()

		// Second cleanup should not panic
		env.Cleanup()
	})

	t.Run("database cleanup", func(t *testing.T) {
		env := NewTestSuite(t)

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
