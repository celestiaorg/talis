package types

// Response represents the response from the infrastructure API
// swagger:model
// Example: {"id":123,"status":"running"}
type Response struct {
	// Unique identifier for the infrastructure job
	ID uint `json:"id"`

	// Current status of the infrastructure job (e.g., "pending", "running", "completed", "failed")
	Status string `json:"status"`
}

// TaskResponse represents the response when a task is created or acted upon
// swagger:model
// Example: {"task_name":"create_instance"}
type TaskResponse struct {
	// Name of the task that was created or acted upon
	TaskName string `json:"task_name"`
}

// PaginationResponse represents pagination information for list endpoints
// swagger:model
// Example: {"total":42,"page":1,"limit":10,"offset":0}
type PaginationResponse struct {
	// Total number of items available across all pages
	Total int `json:"total"`

	// Current page number (1-based)
	Page int `json:"page"`

	// Maximum number of items per page
	Limit int `json:"limit"`

	// Number of items skipped from the beginning of the result set
	Offset int `json:"offset"`
}

// PublicIPs represents the public IP address of a single instance
// swagger:model
// Example: {"public_ip":"203.0.113.42"}
type PublicIPs struct {
	// The public IPv4 address of the instance
	PublicIP string `json:"public_ip"`
}

// PublicIPsResponse represents the response from the public IPs endpoint
// swagger:model
// Example: {"public_ips":[{"public_ip":"203.0.113.42"},{"public_ip":"198.51.100.7"}],"pagination":{"total":2,"page":1,"limit":10,"offset":0}}
type PublicIPsResponse struct {
	// List of public IP addresses for instances
	PublicIPs []PublicIPs `json:"public_ips"`

	// Pagination information for the result set
	Pagination PaginationResponse `json:"pagination"`
}

// ListResponse defines a generic response structure for listing resources
// swagger:model
// Example: {"rows":[{"id":1,"name":"example"},{"id":2,"name":"example2"}],"pagination":{"total":2,"page":1,"limit":10,"offset":0}}
type ListResponse[T any] struct {
	// Array of resource items
	Rows []T `json:"rows"`

	// Pagination information for the result set
	Pagination PaginationResponse `json:"pagination"`
}

// InstanceListResponse represents a response containing a list of instances
// swagger:model
// Example: {"rows":[{"id":1,"provider_id":"do","status":"ready","public_ip":"203.0.113.42"},{"id":2,"provider_id":"aws","status":"provisioning"}],"pagination":{"total":2,"page":1,"limit":10,"offset":0}}
type InstanceListResponse struct {
	// Array of instance objects
	Rows []interface{} `json:"rows"`

	// Pagination information for the result set
	Pagination PaginationResponse `json:"pagination"`
}

// TaskListResponse represents a response containing a list of tasks
// swagger:model
// Example: {"rows":[{"id":1,"name":"create_instance","status":"completed"},{"id":2,"name":"provision_instance","status":"pending"}],"pagination":{"total":2,"page":1,"limit":10,"offset":0}}
type TaskListResponse struct {
	// Array of task objects
	Rows []interface{} `json:"rows"`

	// Pagination information for the result set
	Pagination PaginationResponse `json:"pagination"`
}

// ErrorResponse represents an error response
// swagger:model
// Example: {"error":"Invalid input parameter","details":{"field":"region","message":"Region is required"}}
type ErrorResponse struct {
	// Error message describing what went wrong
	Error string `json:"error"`

	// Optional additional details about the error, may include field-specific validation errors
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
// swagger:model
// Example: {"data":{"id":123,"status":"created"}}
type SuccessResponse struct {
	// Optional data returned by the operation, may include created resource IDs or other operation results
	Data interface{} `json:"data,omitempty"`
}
