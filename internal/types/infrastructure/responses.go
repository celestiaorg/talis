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
