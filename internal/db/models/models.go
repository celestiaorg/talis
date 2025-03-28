package models

// StatusFilter represents how to filter by status
type StatusFilter string

const (
	StatusFilterEqual    StatusFilter = "equal"
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
