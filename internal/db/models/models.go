package models

import "math"

// AdminID represents the special ID for admin-level access
const AdminID uint = math.MaxUint32

// ListOptions represents pagination and filtering options for list operations
type ListOptions struct {
	Limit          int  `json:"limit"`  // Number of items to return
	Offset         int  `json:"offset"` // Number of items to skip
	IncludeDeleted bool `json:"include_deleted"`
}
