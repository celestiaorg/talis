// Package types contains PUBLIC aliases for internal request/response structs.
//
// NOTE: This package uses type aliases to internal definitions
// as a temporary measure. This should be revisited
// during a proper refactoring to define stable public types.
package types

import (
	internaltypes "github.com/celestiaorg/talis/internal/types"
)

// ListResponse is a generic response structure for lists (public alias).
type ListResponse[T any] struct {
	Rows  []*T `json:"rows"`
	Total int  `json:"total"`
}

// SlugResponse represents a response containing a slug and potentially data (public alias).
// NOTE: We alias the internal type directly here as it's generic enough.
type SlugResponse = internaltypes.SlugResponse

// Slug is a type alias for internaltypes.Slug.
type Slug = internaltypes.Slug

// Slug constants (public aliases)
const (
	SuccessSlug      Slug = internaltypes.SuccessSlug
	ErrorSlug        Slug = internaltypes.ErrorSlug
	InvalidInputSlug Slug = internaltypes.InvalidInputSlug
	ServerErrorSlug  Slug = internaltypes.ServerErrorSlug
	NotFoundSlug     Slug = internaltypes.NotFoundSlug
)

// TaskResponse represents the detailed response for a task operation (public alias).
type TaskResponse = internaltypes.TaskResponse
