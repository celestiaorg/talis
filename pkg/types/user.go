// Package types contains PUBLIC aliases for internal request/response structs.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package types

import (
	internaltypes "github.com/celestiaorg/talis/internal/types"
)

// UserResponse defines the structure for the response containing user details (public alias).
type UserResponse = internaltypes.UserResponse

// CreateUserResponse defines the structure for the response after creating a user (public alias).
type CreateUserResponse = internaltypes.CreateUserResponse
