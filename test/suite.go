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
	"github.com/celestiaorg/talis/test/mocks"
)

// DefaultTestTimeout is the default timeout for test suites.
const DefaultTestTimeout = 30 * time.Second

// Suite encapsulates all components needed for integration testing.
// It provides a complete test setup with:
//   - In-memory database
//   - Real API server
//   - Real API client
//   - Mocked external providers
type Suite struct {
	t *testing.T // The testing.T instance for this suite

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
	MockDOClient *mocks.MockDOClient

	// Context management
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Cleanup function
	cleanup func()
}

// NewTestSuite creates a new test suite with the given options.
// The suite must be cleaned up after use by calling Cleanup.
func NewTestSuite(t *testing.T) *Suite {
	t.Helper()

	// Create suite with default timeout
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTestTimeout)

	suite := &Suite{
		t:          t,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	// Initialize cleanup function
	suite.cleanup = func() {
		if suite.Server != nil {
			suite.Server.Close()
		}
		if suite.cancelFunc != nil {
			suite.cancelFunc()
		}
		// Close database if it exists
		if suite.DB != nil {
			sqlDB, err := suite.DB.DB()
			if err == nil && sqlDB != nil {
				_ = sqlDB.Close()
			}
		}
	}

	// Setup database by default
	SetupTestDB(suite, nil)

	// Setup server by default
	SetupServer(suite)

	// Setup mock DO client by default
	SetupMockDOClient(suite)

	return suite
}

// Cleanup tears down the test suite, releasing all resources.
// This should be deferred immediately after creating the suite.
func (s *Suite) Cleanup() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// Context returns the suite's context, which is automatically
// canceled when the suite is cleaned up.
func (s *Suite) Context() context.Context {
	return s.ctx
}

// Require returns a require.Assertions instance for this suite.
// This is a convenience method to avoid passing t around.
func (s *Suite) Require() *require.Assertions {
	return require.New(s.t)
}

// Retry retries a function until it succeeds or the number of retries is reached.
func (s *Suite) Retry(fn func() error, retries int, interval time.Duration) (err error) {
	for i := 0; i < retries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(interval)
	}
	return
}
