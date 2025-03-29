package services

import (
	"context"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

type User struct {
	repo *repos.UserRepository
}

func NewUserService(repo *repos.UserRepository) *User {
	return &User{
		repo: repo,
	}
}

func (s User) CreateUser(ctx context.Context, userReq *infrastructure.CreateUserRequest) (v uint, err error) {
	user := &models.User{
		Username: userReq.Username,
		Email:    userReq.Email,
		Role:     userReq.Role,
	}
	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return v, err
	}
	return user.ID, nil
}

func (s User) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s User) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
