package models

import (
	"encoding/json"
	"fmt"
	"math"

	"gorm.io/gorm"

	"github.com/celestiaorg/talis/internal/logger"
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
	PublicSSHKey string   `json:"public_ssh_key" gorm:""`
}

// MarshalJSON implements the json.Marshaler interface for User
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User // Create an alias to avoid infinite recursion
	return json.Marshal(Alias(u))
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

// AdminID represents the special ID for admin-level access
const AdminID uint = math.MaxUint32

// ValidateOwnerID ensures the ownerID is valid
func ValidateOwnerID(ownerID uint) error {
	if ownerID == 0 {
		// TODO: remove this once we have a proper logging system
		logger.Warn("owner_id cannot be 0")
		// return fmt.Errorf("owner_id cannot be 0")
	}
	return nil
}
