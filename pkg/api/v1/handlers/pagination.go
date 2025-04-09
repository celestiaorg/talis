package handlers

import "github.com/celestiaorg/talis/internal/db/models"

const (
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 100
	// MinPageSize is the minimum allowed page size
	MinPageSize = 1
	// MaxPageSize is the maximum allowed page size
	MaxPageSize = 1000
)

// getPaginationOptions returns a ListOptions struct with validated pagination parameters
func getPaginationOptions(page int, includeDeleted ...bool) *models.ListOptions {
	// Validate and set defaults for page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * models.DefaultLimit
	options := &models.ListOptions{
		Limit:  models.DefaultLimit,
		Offset: offset,
	}

	// Set include_deleted if provided
	if len(includeDeleted) > 0 {
		options.IncludeDeleted = includeDeleted[0]
	}

	return options
}
