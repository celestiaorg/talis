package models

const (
  // DefaultLimit is the max number of rows that are retrieved from the DB per listing API call
  DefaultLimit = 50
)

// StatusFilter represents how to filter instances by status
type StatusFilter string

const (
	// StatusFilterEqual indicates filtering for instances with matching status
	StatusFilterEqual StatusFilter = "equal"
	// StatusFilterNotEqual indicates filtering for instances with non-matching status
	StatusFilterNotEqual StatusFilter = "not_equal"
)

// ListOptions represents pagination and filtering options for list operations
type ListOptions struct {
	Limit          int             `json:"limit"`  // Number of items to return
	Offset         int             `json:"offset"` // Number of items to skip
	IncludeDeleted bool            `json:"include_deleted"`
	Status         *InstanceStatus `json:"status,omitempty"`        // Filter by instance status
	StatusFilter   StatusFilter    `json:"status_filter,omitempty"` // How to filter by status
}
