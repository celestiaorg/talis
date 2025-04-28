// Package handlers provides HTTP request handling
package handlers

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/celestiaorg/talis/pkg/models"
)

// UserGetParams defines the parameters for retrieving a user by id
type UserGetParams struct {
	Username string `json:"username,omitempty"`
	Page     int    `json:"page,omitempty"`
}

// Validate validates the parameters for retrieving a user by username or retrieving all users
func (p *UserGetParams) Validate() error {
	// Ensure page is positive if provided
	if p.Page < 0 {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgNegativePagination))
	}
	return nil
}

// UserGetByIDParams defines the parameters for retrieving a user by id
type UserGetByIDParams struct {
	ID uint `json:"id"`
}

// Validate validates the parameters for retrieving a user by id
func (p UserGetByIDParams) Validate() error {
	if p.ID <= 0 {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgUserIDRequired))
	}
	return nil
}

// CreateUserParams defines the parameters for creating a user
type CreateUserParams struct {
	Username     string          `json:"username" gorm:"not null;unique"`
	Email        string          `json:"email" gorm:""`
	Role         models.UserRole `json:"role" gorm:"index"`
	PublicSSHKey string          `json:"public_ssh_key" gorm:""`
}

// Validate validates the parameters for creating a user
func (p CreateUserParams) Validate() error {
	if p.Username == "" {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgUsernameRequired))
	}
	if p.Email != "" {
		if _, err := mail.ParseAddress(p.Email); err != nil {
			return fmt.Errorf("%s", strings.ToLower(ErrMsgInvalidUserEmail))
		}
	}
	return nil
}

// DeleteUserParams defines the parameters for deleting a user with a given user id
type DeleteUserParams struct {
	ID uint `json:"id"`
}

// Validate validates the parameters for deleting a user
func (p DeleteUserParams) Validate() error {
	if p.ID <= 0 {
		return fmt.Errorf("%s", strings.ToLower(ErrMsgUserIDRequired))
	}
	return nil
}
