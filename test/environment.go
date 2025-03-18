package test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/api/v1/client"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/test/mocks/compute"
)

// TestEnvironment encapsulates all components needed for integration testing.
// It provides a complete test setup with:
//   - In-memory database
//   - Real API server
//   - Real API client
//   - Mocked external providers
type TestEnvironment struct {
	t *testing.T // The testing.T instance for this environment

	// Server components
	App    *fiber.App
	Server *httptest.Server

	// Client components
	APIClient client.Client

	// Database components
	DB           *gorm.DB
	JobRepo      *repos.JobRepository
	InstanceRepo *repos.InstanceRepository

	// Mock providers
	MockDOClient *compute.MockDOClient

	// Context management
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Cleanup function
	cleanup func()
}

// NewTestEnvironment creates a new test environment with the given options.
// The environment must be cleaned up after use by calling Cleanup.
func NewTestEnvironment(t *testing.T, opts ...Option) *TestEnvironment {
	t.Helper()

	// Create environment with default timeout
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTestTimeout)

	env := &TestEnvironment{
		t:          t,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	// Initialize cleanup function
	env.cleanup = func() {
		if env.Server != nil {
			env.Server.Close()
		}
		if env.cancelFunc != nil {
			env.cancelFunc()
		}
		// Close database if it exists
		if env.DB != nil {
			sqlDB, err := env.DB.DB()
			if err == nil && sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
	}

	// Setup database by default
	WithDB(nil)(env)

	// Apply additional options
	for _, opt := range opts {
		opt(env)
	}

	return env
}

// Context returns the environment's context, which is automatically
// canceled when the environment is cleaned up.
func (e *TestEnvironment) Context() context.Context {
	return e.ctx
}

// Cleanup tears down the test environment, releasing all resources.
// This should be deferred immediately after creating the environment.
func (e *TestEnvironment) Cleanup() {
	if e.cleanup != nil {
		e.cleanup()
	}
}

// Require returns a require.Assertions instance for this environment.
// This is a convenience method to avoid passing t around.
func (e *TestEnvironment) Require() *require.Assertions {
	return require.New(e.t)
}

// WithTimeout returns a new context with the specified timeout.
// The returned context is a child of the environment's context.
func (e *TestEnvironment) WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(e.ctx, timeout)
}

// T returns the testing.T instance for this environment.
// This is useful for test helpers that need access to the test instance.
func (e *TestEnvironment) T() *testing.T {
	return e.t
}
