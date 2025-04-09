package types

// Response represents the response from the infrastructure API
type Response struct {
	ID     uint   `json:"id"`     // ID of the job
	Status string `json:"status"` // Status of the job
}

// PaginationResponse represents pagination information
type PaginationResponse struct {
	Total  int `json:"total"`  // Total number of values
	Page   int `json:"page"`   // Current page number
	Limit  int `json:"limit"`  // Number of items per page
	Offset int `json:"offset"` // Offset from start of results
}

// PublicIPs represents the public IPs of the instances
type PublicIPs struct {
	JobID    uint   `json:"job_id"`
	PublicIP string `json:"public_ip"`
}

// PublicIPsResponse represents the response from the public IPs endpoint
type PublicIPsResponse struct {
	PublicIPs  []PublicIPs        `json:"public_ips"` // List of public IPs
	Pagination PaginationResponse `json:"pagination"` // Pagination information
}

// ListResponse defines a generic response structure for listing resources
type ListResponse[T any] struct {
	Rows       []T                `json:"rows"`
	Pagination PaginationResponse `json:"pagination"`
}
