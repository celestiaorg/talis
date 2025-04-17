package repos

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/db/models"
)

// UserRepository handles database operations for user entities
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository handles database operations for user entities
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
// Returns an error if the username already exists
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	_, err := r.GetUserByUsername(ctx, user.Username)
	if err == nil {
		return fmt.Errorf("username already exists")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("error checking username existence: %w", err)
	}
	return r.db.WithContext(ctx).Create(user).Error
}

// CreateBatch creates a batch of users in the database
func (r *UserRepository) CreateBatch(ctx context.Context, users []*models.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(users, 100).Error
	})
}

// GetUserByUsername retrieves a user by their username
// Returns ErrRecordNotFound if the user doesn't exist
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by their ID
// Returns ErrRecordNotFound if the user doesn't exist
func (r *UserRepository) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, userID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUsers retrieves all users
func (r *UserRepository) GetUsers(ctx context.Context, opts *models.ListOptions) ([]models.User, error) {
	var users []models.User
	db := r.db.WithContext(ctx)

	query := db.Model(&models.User{})
	if opts != nil {
		query = query.Limit(opts.Limit).Offset(opts.Offset)
		if !opts.IncludeDeleted {
			query = query.Unscoped().Where("deleted_at IS NULL")
		}
	}
	err := query.Find(&users).Error

	return users, err
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, userID uint) error {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&models.User{}, userID).Error
}
