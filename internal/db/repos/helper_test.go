package repos

import (
	"context"
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/celestiaorg/talis/internal/db/models"
)

// DBRepositoryTestSuite provides a base test suite for repository tests
type DBRepositoryTestSuite struct {
	suite.Suite
	db           *gorm.DB
	ctx          context.Context
	jobRepo      *JobRepository
	instanceRepo *InstanceRepository
	userRepo     *UserRepository
}

// randomOwnerID creates a random owner ID using crypto/rand
func (s *DBRepositoryTestSuite) randomOwnerID() uint {
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	s.Require().NoError(err, "Failed to generate random owner ID")
	return uint(n.Uint64() + 1) // +1 to avoid 0
}

// Retry retries a function until it succeeds or the number of retries is reached.
func (s *DBRepositoryTestSuite) Retry(fn func() error, retries int, interval time.Duration) (err error) {
	for i := 0; i < retries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(interval)
	}
	return
}

func (s *DBRepositoryTestSuite) SetupTest() {
	// Create new in-memory database with JSON support
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_json=1"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		DryRun:                                   false,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	require.NoError(s.T(), err, "Failed to create in-memory database")

	// Run migrations
	err = db.AutoMigrate(&models.Instance{}, &models.Job{}, &models.User{})
	require.NoError(s.T(), err, "Failed to run database migrations")

	// Initialize repositories
	s.db = db
	s.jobRepo = NewJobRepository(s.db)
	s.instanceRepo = NewInstanceRepository(s.db)
	s.userRepo = NewUserRepository(s.db)
	s.ctx = context.Background()
}

func (s *DBRepositoryTestSuite) TearDownTest() {
	sqlDB, err := s.db.DB()
	if err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
}

// Helper methods for creating test data

func (s *DBRepositoryTestSuite) createTestInstance() *models.Instance {
	return s.createTestInstanceForOwner(s.randomOwnerID())
}

func (s *DBRepositoryTestSuite) createTestInstanceForOwner(ownerID uint) *models.Instance {
	instance := &models.Instance{
		OwnerID:    ownerID,
		JobID:      1,
		ProviderID: models.ProviderDO,
		Name:       "test-instance",
		PublicIP:   "192.0.2.1",
		Region:     "nyc1",
		Size:       "s-1vcpu-1gb",
		Image:      "ubuntu-20-04-x64",
		Tags:       []string{"test", "dev"},
		Status:     models.InstanceStatusPending,
		CreatedAt:  time.Now(),
	}
	err := s.instanceRepo.Create(s.ctx, instance)
	s.Require().NoError(err)
	return instance
}

func (s *DBRepositoryTestSuite) createTestJob() *models.Job {
	job := &models.Job{
		Name:         "test-job",
		InstanceName: "test-instance",
		ProjectName:  "test-project",
		OwnerID:      1,
		Status:       models.JobStatusPending,
		SSHKeys:      models.SSHKeys{"key1", "key2"},
		WebhookURL:   "https://example.com/webhook",
		WebhookSent:  false,
		CreatedAt:    time.Now(),
	}
	err := s.jobRepo.Create(s.ctx, job)
	s.Require().NoError(err)
	return job
}

func (s *DBRepositoryTestSuite) createTestUser() *models.User {
	user := &models.User{
		Username: "test-user",
		Email:    "test@example.com",
		Role:     models.UserRoleUser,
	}
	err := s.userRepo.CreateUser(s.ctx, user)
	s.Require().NoError(err)
	return user
}

// TestDBRepository runs the test suite for the DBRepository to verify no panic
func TestDBRepository(t *testing.T) {
	suite.Run(t, new(DBRepositoryTestSuite))
}
