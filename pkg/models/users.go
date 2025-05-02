package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// UserRole represents the role of a user in the system
type UserRole = internalmodels.UserRole

// Constants for user roles
const (
	// UserRoleAdmin represents an administrator user role.
	UserRoleAdmin UserRole = internalmodels.UserRoleAdmin
	// UserRoleUser represents a standard user role.
	UserRoleUser UserRole = internalmodels.UserRoleUser
)

// User represents a user in the system
type User = internalmodels.User

// AdminID represents the special ID for admin-level access
const AdminID uint = internalmodels.AdminID

// ValidateOwnerID ensures the ownerID is valid
var (
	ValidateOwnerID = internalmodels.ValidateOwnerID
	ParseUserRole   = internalmodels.ParseUserRole
)
