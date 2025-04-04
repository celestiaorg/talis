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
	assert.NotNil(t, env.T(), "testing.T should be set")
	assert.Same(t, t, env.T())
	assert.NotNil(t, env.App, "app should be initialized")
	assert.NotNil(t, env.Server, "server should be initialized")
	assert.NotNil(t, env.APIClient, "API client should be initialized")
	assert.NotNil(t, env.db, "database should be initialized")
	assert.NotNil(t, env.jobRepo, "job repository should be initialized")
	assert.NotNil(t, env.instanceRepo, "instance repository should be initialized")
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
		require.NotNil(t, env.db, "database should be initialized")
		require.NotNil(t, env.jobRepo, "job repository should be initialized")
		require.NotNil(t, env.instanceRepo, "instance repository should be initialized")

		// Verify database is working
		job := &models.Job{
			Name:        "test-job",
			ProjectName: "test-project",
			OwnerID:     1, // Set owner ID for the test
		}
		result := env.db.Create(job)
		assert.NoError(t, result.Error, "should create job without error")
		assert.NotZero(t, job.ID, "job should have an ID")

		// Verify job repository is working
		savedJob, err := env.jobRepo.GetByID(env.ctx, job.OwnerID, job.ID)
		assert.NoError(t, err, "should get job without error")
		assert.Equal(t, job.Name, savedJob.Name, "job names should match")

		// Verify instance repository is working
		instance := &models.Instance{
			Name:    "test-instance",
			JobID:   job.ID,
			OwnerID: job.OwnerID,
			Status:  models.InstanceStatusPending,
		}
		result = env.db.Create(instance)
		assert.NoError(t, result.Error, "should create instance without error")
		assert.NotZero(t, instance.ID, "instance should have an ID")

		savedInstance, err := env.InstanceRepo.GetByID(env.ctx, instance.OwnerID, instance.JobID, instance.ID)
		assert.NoError(t, err, "should get instance without error")
		assert.Equal(t, instance.Name, savedInstance.Name, "instance names should match")

		// Verify user repository is working
		user := &models.User{
			Username: "user1",
			Email:    "user1@email.com",
			Role:     1,
		}
		result = env.DB.Create(user)
		assert.NoError(t, result.Error, "should create user without error")
		assert.NotZero(t, user.ID, "user should have an ID")

		usr, err := env.UserRepo.GetUserByID(env.ctx, user.ID)
		assert.NoError(t, err, "should get user without error")
		assert.Equal(t, usr.Username, user.Username, "user usernames should match")

		usr, err = env.UserRepo.GetUserByUsername(env.ctx, user.Username)
		assert.NoError(t, err, "should get user without error")
		assert.Equal(t, usr.Username, user.Username, "user usernames should match")
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
		env.db.Create(job)

		// Get the underlying sql.DB
		sqlDB, err := env.db.DB()
		require.NoError(t, err)

		// Cleanup should close the connection
		env.Cleanup()

		// Verify connection is closed
		err = sqlDB.Ping()
		assert.Error(t, err, "database connection should be closed")
	})
}
