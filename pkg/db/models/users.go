// Package models contains PUBLIC aliases for database models and related types.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// UserRole represents the role of a user
type UserRole = internalmodels.UserRole

// User role constants
const (
	UserRoleUser  UserRole = internalmodels.UserRoleUser
	UserRoleAdmin UserRole = internalmodels.UserRoleAdmin
)

// User represents a user in the system (public alias).
type User = internalmodels.User

// Constants related to users
const (
	AdminID = internalmodels.AdminID
)

// Functions related to users (public aliases)
var (
	ValidateOwnerID = internalmodels.ValidateOwnerID
)

// NOTE: Methods are defined on the original internal types.
