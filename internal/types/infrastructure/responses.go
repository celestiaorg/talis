package infrastructure

import (
	"github.com/celestiaorg/talis/internal/db/models"
)

// Response represents the response from the infrastructure API
type Response struct {
	ID     uint   `json:"id"`     // ID of the job
	Status string `json:"status"` // Status of the job
}

// CreateRequest represents a request to create infrastructure
type CreateRequest struct {
	Name        string            `json:"name"`         // Name of the job
	ProjectName string            `json:"project_name"` // Project name of the job
	WebhookURL  string            `json:"webhook_url"`  // Webhook URL of the job
	Instances   []InstanceRequest `json:"instances"`    // Instances to create
}

// ListInstancesResponse represents the response from the list instances endpoint
type ListInstancesResponse struct {
	Instances  []models.Instance  `json:"instances"`  // List of instances
	Pagination PaginationResponse `json:"pagination"` // Pagination information
}

// ListJobsResponse represents the response from the list jobs endpoint
type ListJobsResponse struct {
	Slug       Slug               `json:"slug"`       // Slug of the job
	Jobs       []models.Job       `json:"jobs"`       // List of jobs
	Pagination PaginationResponse `json:"pagination"` // Pagination information
}

// JobInstancesResponse represents the response from getting instances for a specific job
type JobInstancesResponse struct {
	Instances []models.Instance `json:"instances"` // List of instances for the job
	Total     int               `json:"total"`     // Total number of instances
	JobID     uint              `json:"job_id"`    // ID of the job
}

// InstanceMetadataResponse represents the metadata response for instances
type InstanceMetadataResponse struct {
	Instances  []models.Instance  `json:"instances"`  // List of instances
	Pagination PaginationResponse `json:"pagination"` // Pagination information
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

// CreateUserResponse represents the response from the create user endpoint
type CreateUserResponse struct {
	UserID uint `json:"id"`
}

// UserResponse is a flexible response type for both single and multiple user scenarios
type UserResponse struct {
	// This can be a single user or null when returning multiple users
	User models.User `json:"user,omitempty"`

	// This can be an array of users or null when returning a single user
	Users []models.User `json:"users,omitempty"`

	// Pagination info included only when returning multiple users
	Pagination PaginationResponse `json:"pagination,omitempty"`
}
