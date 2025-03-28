package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/celestiaorg/talis/internal/db/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create create a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if _, err := r.GetUserByUsername(ctx, user.Username); err == nil {
		return fmt.Errorf("usernmae %w, already exists", err)
	}
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if err != nil {

		return nil, fmt.Errorf("failed to get user: %w", err) // why error with using default loggerr??
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userId uint) (*models.User, error) {
	var user *models.User
	err := r.db.WithContext(ctx).Find(&user, userId).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user %w", err)
	}
	return user, nil
}
