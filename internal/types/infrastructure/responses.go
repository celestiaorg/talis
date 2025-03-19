package infrastructure

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

// ListJobsResponse represents the response from the list jobs endpoint
type ListJobsResponse struct {
	Jobs   []JobStatus `json:"jobs"`   // List of jobs
	Total  int         `json:"total"`  // Total number of jobs
	Page   int         `json:"page"`   // Current page number
	Limit  int         `json:"limit"`  // Number of items per page
	Offset int         `json:"offset"` // Offset from start of results
}
