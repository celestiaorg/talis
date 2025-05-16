package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
)

func TestNewSuite(t *testing.T) {
	suite := NewSuite(t)
	defer suite.Cleanup()

	// Should have valid components
	assert.NotNil(t, suite.App, "app should be initialized")
	assert.NotNil(t, suite.Server, "server should be initialized")
	assert.NotNil(t, suite.APIClient, "API client should be initialized")

	assert.NotNil(t, suite.DB, "database should be initialized")
	assert.NotNil(t, suite.InstanceRepo, "instance repository should be initialized")
	assert.NotNil(t, suite.UserRepo, "user repository should be initialized")
	assert.NotNil(t, suite.ProjectRepo, "project repository should be initialized")
	assert.NotNil(t, suite.TaskRepo, "task repository should be initialized")
	assert.NotNil(t, suite.MockDOClient, "mock DO client should be initialized")
}

func TestTestEnvironment_Database(t *testing.T) {
	t.Run("default in-memory database", func(t *testing.T) {
		env := NewSuite(t)
		defer env.Cleanup()

		// Check database components
		require.NotNil(t, env.DB, "database should be initialized")
		require.NotNil(t, env.InstanceRepo, "instance repository should be initialized")

		// Verify instance repository is working
		instance := &models.Instance{
			OwnerID: 1,
			Status:  models.InstanceStatusPending,
		}
		result := env.DB.Create(instance)
		assert.NoError(t, result.Error, "should create instance without error")
		assert.NotZero(t, instance.ID, "instance should have an ID")

		savedInstance, err := env.InstanceRepo.Get(env.ctx, instance.OwnerID, instance.ID)
		assert.NoError(t, err, "should get instance without error")
		assert.Equal(t, instance.ID, savedInstance.ID, "instance IDs should match")

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
		assert.Equal(t, user.Username, usr.Username, "usernames should match")
		assert.Equal(t, user.Email, usr.Email, "emails should match")
	})
}

func TestTestEnvironment_Cleanup(t *testing.T) {
	t.Run("multiple cleanup calls", func(t *testing.T) {
		env := NewSuite(t)

		// First cleanup should work
		env.Cleanup()

		// Second cleanup should not panic
		env.Cleanup()
	})

	t.Run("database cleanup", func(t *testing.T) {
		env := NewSuite(t)

		// Create a test record
		// job := &models.Job{Name: "cleanup-test"} // Removed as Jobs are deprecated
		// env.DB.Create(job)

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
