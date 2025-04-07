package types

import "github.com/celestiaorg/talis/internal/db/models"

// JobRequest represents the infrastructure request
type JobRequest struct {
	Name    string `json:"name"`
	OwnerID uint   `json:"owner_id"`
}

// JobStatus represents the status of an infrastructure job
type JobStatus struct {
	JobID     string `json:"job_id"`     // ID of the job
	Status    string `json:"status"`     // Status of the job
	CreatedAt string `json:"created_at"` // Timestamp when the job was created
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
