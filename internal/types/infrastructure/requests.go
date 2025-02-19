package infrastructure

// JobRequest represents the infrastructure request
type JobRequest struct {
	Name        string            `json:"name"`
	ProjectName string            `json:"project_name"`
	Provider    string            `json:"provider"`
	Instances   []InstanceRequest `json:"instances"`
	Action      string            `json:"action"` // "create" or "delete"
}

// Instance represents a compute instance configuration
type InstanceRequest struct {
	Provider          string   `json:"provider"`
	NumberOfInstances int      `json:"number_of_instances"`
	Provision         bool     `json:"provision"`
	Region            string   `json:"region"`
	Size              string   `json:"size"`
	Image             string   `json:"image"`
	Tags              []string `json:"tags"`
	SSHKeyName        string   `json:"ssh_key_name"`
}

// DeleteRequest represents a request to delete infrastructure
type DeleteRequest struct {
	Name        string           `json:"name"`
	ProjectName string           `json:"project_name"`
	WebhookURL  string           `json:"webhook_url"`
	Provider    string           `json:"provider"`
	Instances   []DeleteInstance `json:"instances"`
}

// DeleteInstance represents the configuration for deleting an instance
type DeleteInstance struct {
	Provider          string   `json:"provider"`
	NumberOfInstances int      `json:"number_of_instances"`
	Region            string   `json:"region"`
	Size              string   `json:"size"`
	Image             string   `json:"image"`
	Tags              []string `json:"tags"`
	SSHKeyName        string   `json:"ssh_key_name"`
}

// InstanceInfo represents information about a created instance
type InstanceInfo struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// JobStatus represents the status of an infrastructure job
type JobStatus struct {
	JobID     string `json:"job_id"`     // ID of the job
	Status    string `json:"status"`     // Status of the job
	CreatedAt string `json:"created_at"` // Timestamp when the job was created
}
