package models

const (
	DefaultLimit = 50
)

// ListOptions represents pagination and filtering options for list operations
type ListOptions struct {
	Limit          int  // Number of items to return
	Offset         int  // Number of items to skip
	IncludeDeleted bool // Include deleted items
}
