package infrastructure

import "github.com/celestiaorg/talis/internal/db/models"

// --------------------------------------------------
// Instance
// --------------------------------------------------

// InstancesRequest represents a request to manage instances, including creation and deletion.
type InstancesRequest struct {
	JobName      string            `json:"job_name"`
	InstanceName string            `json:"instance_name"`
	Instances    []InstanceRequest `json:"instances"`
	WebhookURL   string            `json:"webhook_url"`
	Action       string            `json:"action"`
	ProjectName  string            `json:"project_name"`
	Provider     models.ProviderID `json:"provider"`
}

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

// InstanceCreateRequest represents the JSON structure for creating infrastructure
type InstanceCreateRequest struct {
	InstanceName string            `json:"instance_name"`
	ProjectName  string            `json:"project_name"`
	WebhookURL   string            `json:"webhook_url,omitempty"`
	Instances    []InstanceRequest `json:"instances"`
}

// DeleteInstanceRequest represents the request body for deleting instances
type DeleteInstanceRequest struct {
	JobName       string   `json:"job_name" validate:"required"`             // Job name of the job
	InstanceNames []string `json:"instance_names" validate:"required,min=1"` // Instances to delete
}

// DeleteRequest represents a request to delete infrastructure
type DeleteRequest struct {
	InstanceName string           `json:"instance_name"` // Base name for instances
	ProjectName  string           `json:"project_name"`  // Project name of the job
	WebhookURL   string           `json:"webhook_url"`   // Webhook URL of the job
	Provider     string           `json:"provider"`      // Provider of the compute service
	Instances    []DeleteInstance `json:"instances"`     // Instances to delete
}

// DeleteInstance represents the configuration for deleting an instance
type DeleteInstance struct {
	Provider          string   `json:"provider"`            // Provider of the compute service
	Name              string   `json:"name"`                // Optional specific instance name to delete
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
	Name string `json:"name"`
}

// JobStatus represents the status of an infrastructure job
type JobStatus struct {
	JobID     string `json:"job_id"`     // ID of the job
	Status    string `json:"status"`     // Status of the job
	CreatedAt string `json:"created_at"` // Timestamp when the job was created
}

// --------------------------------------------------
// User
// --------------------------------------------------

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username     string          `json:"username" gorm:"not null;unique"`
	Email        string          `json:"email" gorm:""`
	Role         models.UserRole `json:"role" gorm:"index"`
	PublicSshKey string          `json:"public_ssh_key" gorm:""`
}
