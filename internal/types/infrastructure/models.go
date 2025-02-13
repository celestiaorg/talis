package infrastructure

// Constants for configuration
const (
	nixTimeout = 120 // Timeout for Nix configuration (seconds)
	maxRetries = 5   // Maximum number of connection attempts
	retryDelay = 10  // Delay between retry attempts (seconds)
)

// InstanceRequest represents the API request structure
type InstanceRequest struct {
	Name        string     `json:"name"`                  // Name of the infrastructure
	ProjectName string     `json:"project_name"`          // Name of the Pulumi project
	Action      string     `json:"action"`                // Action to perform (create/delete)
	Instances   []Instance `json:"instances"`             // List of instances to create
	WebhookURL  string     `json:"webhook_url,omitempty"` // URL to notify when job completes
}

// Instance represents a single instance configuration
type Instance struct {
	Provider          string   `json:"provider"`            // Provider to use (digitalocean/aws)
	NumberOfInstances int      `json:"number_of_instances"` // Number of instances to create
	Region            string   `json:"region"`              // Region to create the instances in
	Size              string   `json:"size"`                // Size of the instances to create
	Image             string   `json:"image"`               // Image to use for the instances
	Tags              []string `json:"tags"`                // Tags to add to the instances
	SSHKeyName        string   `json:"ssh_key_name"`        // SSH key name to use for the instances
	Provision         bool     `json:"provision"`           // Control Nix provisioning
}

// InstanceInfo represents a created instance
type InstanceInfo struct {
	Name string `json:"name"` // Name of the instance
	IP   string `json:"ip"`   // IP address of the instance
}

// JobStatus represents the status of an infrastructure job
type JobStatus struct {
	JobID     string `json:"job_id"`     // ID of the job
	Status    string `json:"status"`     // Status of the job
	CreatedAt string `json:"created_at"` // Timestamp when the job was created
}

// DeleteRequest represents the API request structure for deletion
type DeleteRequest struct {
	Name        string           `json:"name"`                  // Name of the infrastructure
	ProjectName string           `json:"project_name"`          // Name of the Pulumi project
	Instances   []DeleteInstance `json:"instances"`             // List of instances to delete
	WebhookURL  string           `json:"webhook_url,omitempty"` // URL to notify when job completes
}

// DeleteInstance represents the minimal instance configuration needed for deletion
type DeleteInstance struct {
	Provider          string `json:"provider"`            // Provider to use (digitalocean/aws)
	NumberOfInstances int    `json:"number_of_instances"` // Number of instances to delete
	Region            string `json:"region"`              // Region to delete the instances in
	Size              string `json:"size"`                // Size of the instances to delete
}
