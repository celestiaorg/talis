package infrastructure

import "github.com/celestiaorg/talis/internal/db/models"

// --------------------------------------------------
// Instance
// --------------------------------------------------

// InstanceRequest represents a request to create or modify a compute instance
type InstanceRequest struct {
	Name              string            `json:"name"`                // Name of the instance
	Provider          models.ProviderID `json:"provider"`            // Provider of the compute service
	NumberOfInstances int               `json:"number_of_instances"` // Number of instances to create
	Provision         bool              `json:"provision"`           // Whether to provision the instance
	Region            string            `json:"region"`              // Region of the instance
	Size              string            `json:"size"`                // Size of the instance
	Image             string            `json:"image"`               // Image of the instance
	Tags              []string          `json:"tags"`                // Tags of the instance
	SSHKeyName        string            `json:"ssh_key_name"`        // SSH key name of the instance
}

// DeleteInstanceRequest represents the request body for deleting instances
type DeleteInstanceRequest struct {
	ID          uint              `json:"id" validate:"required"`              // ID of the job
	Name        string            `json:"name" validate:"required"`            // Name of the job
	ProjectName string            `json:"project_name" validate:"required"`    // Project name of the job
	Instances   []InstanceRequest `json:"instances" validate:"required,min=1"` // Instances to delete
}

// DeleteRequest represents a request to delete infrastructure
type DeleteRequest struct {
	Name        string           `json:"name"`         // Name of the job
	ProjectName string           `json:"project_name"` // Project name of the job
	WebhookURL  string           `json:"webhook_url"`  // Webhook URL of the job
	Provider    string           `json:"provider"`     // Provider of the compute service
	Instances   []DeleteInstance `json:"instances"`    // Instances to delete
}

// DeleteInstance represents the configuration for deleting an instance
type DeleteInstance struct {
	Provider          string   `json:"provider"`            // Provider of the compute service
	NumberOfInstances int      `json:"number_of_instances"` // Number of instances to delete
	Region            string   `json:"region"`              // Region of the instance
	Size              string   `json:"size"`                // Size of the instance
	Image             string   `json:"image"`               // Image of the instance
	Tags              []string `json:"tags"`                // Tags of the instance
	SSHKeyName        string   `json:"ssh_key_name"`        // SSH key name of the instance
}

// InstanceInfo represents information about a created instance
type InstanceInfo struct {
	Name     string            `json:"name"`     // Name of the instance
	IP       string            `json:"ip"`       // IP address of the instance
	Provider models.ProviderID `json:"provider"` // Provider of the compute service
	Region   string            `json:"region"`   // Region of the instance
	Size     string            `json:"size"`     // Size of the instance
}

// --------------------------------------------------
// Job
// --------------------------------------------------

// JobRequest represents the infrastructure request
type JobRequest struct {
	Name        string            `json:"name"`
	ProjectName string            `json:"project_name"`
	Provider    models.ProviderID `json:"provider"`
	Instances   []InstanceRequest `json:"instances"`
	Action      string            `json:"action"` // "create" or "delete"
	WebhookURL  string            `json:"webhook_url"`
}

// JobStatus represents the status of an infrastructure job
type JobStatus struct {
	JobID     string `json:"job_id"`     // ID of the job
	Status    string `json:"status"`     // Status of the job
	CreatedAt string `json:"created_at"` // Timestamp when the job was created
}
