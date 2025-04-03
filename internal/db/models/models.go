package models

// StatusFilter represents how to filter db items by status
type StatusFilter string

const (
	// StatusFilterEqual indicates filtering for instances with matching status
	StatusFilterEqual StatusFilter = "equal"
	// StatusFilterNotEqual indicates filtering for instances with non-matching status
	StatusFilterNotEqual StatusFilter = "not_equal"
)

// ListOptions represents pagination and filtering options for list operations
type ListOptions struct {
	// Pagination
	Limit  int `json:"limit"`  // Number of items to return
	Offset int `json:"offset"` // Number of items to skip
	// Filtering
	IncludeDeleted bool         `json:"include_deleted"`
	StatusFilter   StatusFilter `json:"status_filter,omitempty"` // How to filter by status
	// Statuses
	InstanceStatus *InstanceStatus `json:"instance_status,omitempty"` // Filter by instance status
	JobStatus      *JobStatus      `json:"job_status,omitempty"`      // Filter by job status
}

// UserQueryOptions represents query params for GetUserByUsername operation
type UserQueryOptions struct {
	Username string `json:"username" gorm:"not null;unique"`
}
