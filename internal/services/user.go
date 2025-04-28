package services

import (
	"context"
	"errors"

	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/pkg/models"
)

// User provides business logic for user operations
type User struct {
	repo *repos.UserRepository
}

// User service errors
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserCreateFailed = errors.New("failed to create user")
)

// NewUserService creates a new user service instance
func NewUserService(repo *repos.UserRepository) *User {
	return &User{
		repo: repo,
	}
}

// CreateUser creates a new user
func (s User) CreateUser(ctx context.Context, user *models.User) (uint, error) {
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return 0, errors.Join(ErrUserCreateFailed, err)
	}
	return user.ID, nil
}

// GetUserByUsername retrieves a user by username
func (s User) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.Join(ErrUserNotFound, err)
	}
	return user, nil
}

// GetUserByID retrieves a user by id
func (s User) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.Join(ErrUserNotFound, err)
	}
	return user, nil
}

// GetAllUsers retrieves all users
func (s User) GetAllUsers(ctx context.Context, opts *models.ListOptions) ([]models.User, error) {
	return s.repo.GetUsers(ctx, opts)
}

// DeleteUser deletes a user
func (s User) DeleteUser(ctx context.Context, userID uint) error {
	return s.repo.DeleteUser(ctx, userID)
}
