package types

import "github.com/celestiaorg/talis/internal/db/models"

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username     string          `json:"username" gorm:"not null;unique"`
	Email        string          `json:"email" gorm:""`
	Role         models.UserRole `json:"role" gorm:"index"`
	PublicSSHKey string          `json:"public_ssh_key" gorm:""`
}

// CreateUserResponse represents the response from the create user endpoint
type CreateUserResponse struct {
	UserID uint `json:"id"`
}

// UserResponse is a flexible response type for both single and multiple user scenarios
type UserResponse struct {
	// This can be a single user or null when returning multiple users
	User models.User `json:"user,omitempty"`

	// This can be an array of users or null when returning a single user
	Users []models.User `json:"users,omitempty"`

	// Pagination info included only when returning multiple users
	Pagination PaginationResponse `json:"pagination,omitempty"`
}
