package models

import (
	"fmt"

	"gorm.io/gorm"
)

type UserRole int

const (
	UserRoleUser UserRole = iota
	UserRoleAdmin
)

type User struct {
	gorm.Model
	Username     string   `json:"username" gorm:"not null;unique"`
	Role         UserRole `json:"role" gorm:"index"`
	PublicSshKey string   `json:"public_ssh_key" gorm:""`
}

func (s UserRole) String() string {
	return []string{
		"user",
		"admin",
	}[s]
}

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
