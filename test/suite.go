package test

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/api/v1/client"
	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
)

// DefaultTestTimeout is the default timeout for test suites.
const DefaultTestTimeout = 30 * time.Second

// TestSuite is a test suite for the API
type TestSuite struct {
	*suite.Suite
	db *gorm.DB

	// Context for the test
	ctx        context.Context
	cancelFunc context.CancelFunc

	// Repositories
	jobRepo      *repos.JobRepository
	instanceRepo *repos.InstanceRepository
	userRepo     *repos.UserRepository

	// Mock clients
	MockDOClient *compute.MockDOClient

	// Test data
	testUser *models.User

	// Server components
	App    *fiber.App
	Server *httptest.Server

	// Client components
	APIClient client.Client

	// Cleanup function
	cleanup func()
}

// SetupSuite sets up the test suite
func (s *TestSuite) SetupSuite() {
	// Create a temporary database file
	dbFile := filepath.Join(os.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		s.T().Fatalf("failed to connect database: %v", err)
	}

	s.db = db

	// Create repositories
	s.jobRepo = repos.NewJobRepository(db)
	s.instanceRepo = repos.NewInstanceRepository(db)
	s.userRepo = repos.NewUserRepository(db)

	// Create mock clients
	s.MockDOClient = compute.NewMockDOClient()

	// Create test user
	s.testUser = &models.User{
		Username: "test",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
	}
}

// TearDownSuite tears down the test suite
func (s *TestSuite) TearDownSuite() {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	if s.cleanup != nil {
		s.cleanup()
	}
}

// Run runs the test suite
func Run(t *testing.T) {
	suite.Run(t, NewTestSuite(t))
}

// NewTestSuite creates a new test suite with the given options.
// The suite must be cleaned up after use by calling Cleanup.
func NewTestSuite(t *testing.T) *TestSuite {
	s := &TestSuite{
		Suite: new(suite.Suite),
	}
	s.SetT(t)

	// Create context
	s.ctx, s.cancelFunc = context.WithTimeout(context.Background(), DefaultTestTimeout)

	// Setup database
	SetupTestDB(s, nil)

	// Setup mock client
	SetupMockDOClient(s)

	// Setup server
	SetupServer(s)

	return s
}

// Cleanup tears down the test suite, releasing all resources.
// This should be deferred immediately after creating the suite.
func (s *TestSuite) Cleanup() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// Context returns the suite's context, which is automatically
// canceled when the suite is cleaned up.
func (s *TestSuite) Context() context.Context {
	return s.ctx
}

// Require returns a require.Assertions instance for this suite.
// This is a convenience method to avoid passing t around.
func (s *TestSuite) Require() *require.Assertions {
	return require.New(s.T())
}

// Retry retries a function until it succeeds or the number of retries is reached.
func (s *TestSuite) Retry(fn func() error, retries int, interval time.Duration) (err error) {
	for i := 0; i < retries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(interval)
	}
	return
}
