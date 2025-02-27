package models

import (
	"fmt"

	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
type UserRole int

// User role constants
const (
	// UserRoleUser represents a standard user
	UserRoleUser UserRole = iota
	// UserRoleAdmin represents an administrator user
	UserRoleAdmin
)

// User represents a user in the system
type User struct {
	gorm.Model
	Username     string   `json:"username" gorm:"not null;unique"`
	Email        string   `json:"email" gorm:""`
	Role         UserRole `json:"role" gorm:"index"`
	PublicSshKey string   `json:"public_ssh_key" gorm:""`
	CreatedAt    string   `json:"created_at" gorm:""`
	UpdatedAt    string   `json:"updated_at" gorm:""`
}

func (s UserRole) String() string {
	return []string{
		"user",
		"admin",
	}[s]
}

// ParseUserRole converts a string representation of a user role to UserRole type
func ParseUserRole(str string) (UserRole, error) {
	for i, role := range []string{
		"user",
		"admin",
	} {
		if role == str {
			return UserRole(i), nil
		}
	}
	return UserRoleUser, fmt.Errorf("invalid user role: %s", str)
}
