package models

// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

// StatusFilter represents how to filter db items by status
type StatusFilter = internalmodels.StatusFilter

const (
	// StatusFilterEqual indicates filtering for instances with matching status
	StatusFilterEqual StatusFilter = internalmodels.StatusFilterEqual
	// StatusFilterNotEqual indicates filtering for instances with non-matching status
	StatusFilterNotEqual StatusFilter = internalmodels.StatusFilterNotEqual
)

// ListOptions represents pagination and filtering options for list operations
type ListOptions = internalmodels.ListOptions

// NOTE: Constants like DefaultLimit are not aliased here,
// as they are often specific to internal usage (like DB batching)
// or defined contextually (like DefaultPageSize in handlers).

// UserQueryOptions represents query params for GetUserByUsername operation
type UserQueryOptions = internalmodels.UserQueryOptions
