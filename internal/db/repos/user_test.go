package repos

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/celestiaorg/talis/pkg/models"
)

type UserRepositoryTestSuite struct {
	DBRepositoryTestSuite
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) TestCreateUser() {
	// Test successful user creation
	user := s.createTestUser()
	s.NotZero(user.ID)

	// Test duplicate username
	duplicateUser := &models.User{
		Username: user.Username,
		Email:    "another@example.com",
		Role:     models.UserRoleUser,
	}
	err := s.userRepo.CreateUser(s.ctx, duplicateUser)
	s.Error(err)
	s.Contains(err.Error(), "username already exists")

	// Test batch creation
	users := []*models.User{
		s.randomUser(),
		s.randomUser(),
		s.randomUser(),
	}
	err = s.userRepo.CreateBatch(s.ctx, users)
	s.NoError(err)
	foundUsers, err := s.userRepo.GetUsers(s.ctx, nil)
	s.NoError(err)
	s.Equal(len(foundUsers), 4)
}

func (s *UserRepositoryTestSuite) TestGetUserByUsername() {
	original := s.createTestUser()

	// Create a second user
	user2 := &models.User{
		Username: "test-user-2",
		Email:    "test2@example.com",
		Role:     models.UserRoleAdmin,
	}
	err := s.userRepo.CreateUser(s.ctx, user2)
	s.NoError(err)
	s.NotZero(user2.ID)
	s.NotEqual(original.ID, user2.ID)

	// Test getting first user
	found, err := s.userRepo.GetUserByUsername(s.ctx, original.Username)
	s.NoError(err)
	s.Equal(original.ID, found.ID)
	s.Equal(original.Username, found.Username)
	s.Equal(original.Email, found.Email)
	s.Equal(original.Role, found.Role)

	// Test getting second user
	found2, err := s.userRepo.GetUserByUsername(s.ctx, user2.Username)
	s.NoError(err)
	s.Equal(user2.ID, found2.ID)
	s.Equal(user2.Username, found2.Username)
	s.Equal(user2.Email, found2.Email)
	s.Equal(user2.Role, found2.Role)

	// Test getting non-existent user
	_, err = s.userRepo.GetUserByUsername(s.ctx, "non-existent")
	s.Error(err)
	s.Contains(err.Error(), "user not found")
}

func (s *UserRepositoryTestSuite) TestGetUserByID() {
	original := s.createTestUser()

	// Create a second user
	user2 := &models.User{
		Username: "test-user-2",
		Email:    "test2@example.com",
		Role:     models.UserRoleAdmin,
	}
	err := s.userRepo.CreateUser(s.ctx, user2)
	s.NoError(err)
	s.NotZero(user2.ID)
	s.NotEqual(original.ID, user2.ID)

	// Test getting first user
	found, err := s.userRepo.GetUserByID(s.ctx, original.ID)
	s.NoError(err)
	s.Equal(original.ID, found.ID)
	s.Equal(original.Username, found.Username)
	s.Equal(original.Email, found.Email)
	s.Equal(original.Role, found.Role)

	// Test getting second user
	found2, err := s.userRepo.GetUserByID(s.ctx, user2.ID)
	s.NoError(err)
	s.Equal(user2.ID, found2.ID)
	s.Equal(user2.Username, found2.Username)
	s.Equal(user2.Email, found2.Email)
	s.Equal(user2.Role, found2.Role)

	// Test getting non-existent user
	_, err = s.userRepo.GetUserByID(s.ctx, 999)
	s.Error(err)
	s.Contains(err.Error(), "user not found")
}
