package repos

import (
	"context"
	"crypto/rand"
	"fmt"
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
	instanceRepo *InstanceRepository
	userRepo     *UserRepository
	projectRepo  *ProjectRepository
	taskRepo     *TaskRepository
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
	err = db.AutoMigrate(&models.Instance{}, &models.User{}, &models.Project{}, &models.Task{})
	require.NoError(s.T(), err, "Failed to run database migrations")

	// Initialize repositories
	s.db = db
	s.instanceRepo = NewInstanceRepository(s.db)
	s.userRepo = NewUserRepository(s.db)
	s.projectRepo = NewProjectRepository(s.db)
	s.taskRepo = NewTaskRepository(s.db)
	s.ctx = context.Background()
}

func (s *DBRepositoryTestSuite) TearDownTest() {
	sqlDB, err := s.db.DB()
	if err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}
}

// Helper methods for creating test data

func (s *DBRepositoryTestSuite) randomInstance() *models.Instance {
	return s.randomInstanceForOwner(s.randomOwnerID())
}

func (s *DBRepositoryTestSuite) randomInstanceForOwner(ownerID uint) *models.Instance {
	return &models.Instance{
		OwnerID:    ownerID,
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
}

func (s *DBRepositoryTestSuite) createTestInstance() *models.Instance {
	return s.createTestInstanceForOwner(s.randomOwnerID())
}

func (s *DBRepositoryTestSuite) createTestInstanceForOwner(ownerID uint) *models.Instance {
	instance := s.randomInstanceForOwner(ownerID)
	err := s.instanceRepo.Create(s.ctx, instance)
	s.Require().NoError(err)
	return instance
}

func (s *DBRepositoryTestSuite) randomUser() *models.User {
	return &models.User{
		Username: fmt.Sprintf("test-user-%v", s.randomOwnerID()),
		Email:    fmt.Sprintf("test@example.com-%v", s.randomOwnerID()),
		Role:     models.UserRoleUser,
	}
}

func (s *DBRepositoryTestSuite) createTestUser() *models.User {
	user := s.randomUser()
	err := s.userRepo.CreateUser(s.ctx, user)
	s.Require().NoError(err)
	return user
}

func (s *DBRepositoryTestSuite) randomProject(ownerID uint) *models.Project {
	return &models.Project{
		Name:        fmt.Sprintf("test-project-%v", s.randomOwnerID()),
		Description: "Test project description",
		OwnerID:     ownerID,
		CreatedAt:   time.Now(),
	}
}

func (s *DBRepositoryTestSuite) createTestProject() *models.Project {
	return s.createTestProjectForOwner(s.randomOwnerID())
}

func (s *DBRepositoryTestSuite) createTestProjectForOwner(ownerID uint) *models.Project {
	project := s.randomProject(ownerID)
	err := s.projectRepo.Create(s.ctx, project)
	s.Require().NoError(err)
	return project
}

func (s *DBRepositoryTestSuite) createTestTask() *models.Task {
	project := s.createTestProject()
	return s.createTestTaskForProject(project.OwnerID, project.ID)
}

func (s *DBRepositoryTestSuite) randomTask(ownerID, projectID uint) *models.Task {
	return &models.Task{
		Name:      fmt.Sprintf("test-task-%v", s.randomOwnerID()),
		ProjectID: projectID,
		OwnerID:   ownerID,
		Status:    models.TaskStatusPending,
		Action:    models.TaskActionCreateInstances,
	}
}

func (s *DBRepositoryTestSuite) createTestTaskForProject(ownerID, projectID uint) *models.Task {
	task := s.randomTask(ownerID, projectID)
	err := s.taskRepo.Create(s.ctx, task)
	s.Require().NoError(err)
	return task
}

// TestDBRepository runs the test suite for the DBRepository to verify no panic
func TestDBRepository(t *testing.T) {
	suite.Run(t, new(DBRepositoryTestSuite))
}
