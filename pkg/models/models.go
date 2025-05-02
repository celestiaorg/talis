package models

import (
	internalmodels "github.com/celestiaorg/talis/internal/db/models"
)

const (
	// DefaultLimit is the max number of rows that are retrieved from the DB per listing API call
	DefaultLimit = internalmodels.DefaultLimit
	// DBBatchSize is the standard batch size for DB CreateBatch operations
	DBBatchSize = internalmodels.DBBatchSize
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

// UserQueryOptions represents query params for GetUserByUsername operation
type UserQueryOptions = internalmodels.UserQueryOptions
