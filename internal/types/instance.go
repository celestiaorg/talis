// Package types provides type definitions for the application
package types

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/celestiaorg/talis/internal/db/models"
)

const maxPayloadSize = 2 * 1024 * 1024 // 2MB

// InstanceConfig represents the configuration for creating a new instance
type InstanceConfig struct {
	Name            string         `json:"name"`
	Template        string         `json:"template"`
	Package         string         `json:"package"`
	Size            string         `json:"size"`
	Image           string         `json:"image"`
	Region          string         `json:"region"`
	Memory          int            `json:"memory"`
	Disk            int            `json:"disk"`
	CPU             int            `json:"cpu"`
	HypervisorID    int            `json:"hypervisor_id,omitempty"`
	HypervisorGroup string         `json:"hypervisor_group,omitempty"`
	SSHKeys         []string       `json:"ssh_keys"`
	Volumes         []VolumeConfig `json:"volumes"`
	Network         *NetworkConfig `json:"network,omitempty"`
	UserData        string         `json:"user_data,omitempty"`
}

// InstancesRequest represents an RPC request multiple instances
// NOTE: These should be cleaned up and replaced with specific RPC request types
type InstancesRequest struct {
	InstanceName    string            `json:"instance_name"`
	ProjectName     string            `json:"project_name"`
	TaskName        string            `json:"-"` // used internally by the Infra API
	Instances       []InstanceRequest `json:"instances"`
	WebhookURL      string            `json:"webhook_url"`
	Action          string            `json:"action"`
	Provider        models.ProviderID `json:"provider"`
	Volumes         []VolumeConfig    `json:"volumes"`
	HypervisorID    int               `json:"hypervisor_id,omitempty"`
	HypervisorGroup string            `json:"hypervisor_group,omitempty"`
}

// NetworkConfig represents network configuration for an instance
type NetworkConfig struct {
	ProfileID        int   `json:"profile_id,omitempty"`
	Bandwidth        int   `json:"bandwidth,omitempty"`
	FirewallRulesets []int `json:"firewall_rulesets,omitempty"`
}

// InstanceRequest represents an RPC request for a single instance
// NOTE: These should be cleaned up and replaced with specific RPC request types
type InstanceRequest struct {
	Provider          models.ProviderID `json:"provider"`                   // Cloud provider (e.g., "do")
	Region            string            `json:"region"`                     // Region where instances will be created
	Size              string            `json:"size"`                       // Instance size/type
	Image             string            `json:"image"`                      // OS image to use
	SSHKeyName        string            `json:"ssh_key_name"`               // Name of the SSH key to use
	Tags              []string          `json:"tags"`                       // Tags to apply to instances
	NumberOfInstances int               `json:"number_of_instances"`        // Number of instances to create
	Name              string            `json:"name"`                       // Optional custom name for instances
	Provision         bool              `json:"provision"`                  // Whether to run Ansible provisioning
	Volumes           []VolumeConfig    `json:"volumes"`                    // Optional volumes to attach
	OwnerID           uint              `json:"owner_id"`                   // Owner ID of the instance
	PayloadPath       string            `json:"payload_path,omitempty"`     // Local path to the payload script on the API server
	ExecutePayload    bool              `json:"execute_payload,omitempty"`  // Whether to execute the payload after copying
	SSHKeyType        string            `json:"ssh_key_type,omitempty"`     // Type of the private SSH key for Ansible (e.g., "rsa", "ed25519"). Defaults to "rsa".
	SSHKeyPath        string            `json:"ssh_key_path,omitempty"`     // Custom path to the private SSH key file for Ansible. Overrides defaults.
	HypervisorID      int               `json:"hypervisor_id,omitempty"`    // ID of the hypervisor to use
	HypervisorGroup   string            `json:"hypervisor_group,omitempty"` // Name of the hypervisor group to use
	Network           *NetworkConfig    `json:"network,omitempty"`          // Network configuration
}

// DeleteInstanceRequest represents the request body for deleting instances
type DeleteInstanceRequest struct {
	ProjectName   string   `json:"project_name" validate:"required"`         // Project name
	InstanceNames []string `json:"instance_names" validate:"required,min=1"` // Instances to delete
}

// CreateRequest represents a request to create infrastructure
type CreateRequest struct {
	Name        string            `json:"name"`         // Name of the job
	ProjectName string            `json:"project_name"` // Project name of the job
	WebhookURL  string            `json:"webhook_url"`  // Webhook URL of the job
	Instances   []InstanceRequest `json:"instances"`    // Instances to create
}

// InstanceInfo represents information about a created instance
type InstanceInfo struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Region         string   `json:"region"`
	Size           string   `json:"size"`
	Status         string   `json:"status"`
	IP             string   `json:"ip"`
	PublicIP       string   `json:"public_ip"`
	PayloadPath    string   `json:"payload_path,omitempty"`
	ExecutePayload bool     `json:"execute_payload,omitempty"`
	Volumes        []string `json:"volumes,omitempty"`
}

// Validate validates the infrastructure request
func (r *InstancesRequest) Validate() error {
	if r.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if len(r.Instances) == 0 {
		return fmt.Errorf("at least one instance configuration is required")
	}

	// Validate the instances name
	if r.InstanceName != "" {
		if err := validateHostname(r.InstanceName); err != nil {
			return fmt.Errorf("invalid instance_name: %w", err)
		}
	}

	for i, instance := range r.Instances {
		if instance.Name == "" && r.InstanceName == "" {
			return fmt.Errorf("instance_name or instance.name is required")
		}

		if err := instance.Validate(); err != nil {
			return fmt.Errorf("invalid instance configuration at index %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates the instance configuration
func (i *InstanceRequest) Validate() error {
	if i.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if !i.Provider.IsValid() {
		return fmt.Errorf("unsupported provider: %s", i.Provider)
	}
	if i.NumberOfInstances < 1 {
		return fmt.Errorf("number_of_instances must be greater than 0")
	}
	if i.Region == "" {
		return fmt.Errorf("region is required")
	}
	if i.Size == "" {
		return fmt.Errorf("size is required")
	}
	if i.Image == "" {
		return fmt.Errorf("image is required")
	}
	if i.SSHKeyName == "" {
		return fmt.Errorf("ssh_key_name is required")
	}

	// Validate payload path if provided
	if i.PayloadPath != "" {
		// Check if path is absolute
		if !filepath.IsAbs(i.PayloadPath) {
			return fmt.Errorf("payload_path must be an absolute path")
		}

		// Clean the path
		i.PayloadPath = filepath.Clean(i.PayloadPath)

		// Check file existence and size
		fileInfo, err := os.Stat(i.PayloadPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("payload_path file does not exist: %s", i.PayloadPath)
			}
			return fmt.Errorf("error accessing payload_path file: %w", err)
		}

		if fileInfo.IsDir() {
			return fmt.Errorf("payload_path cannot be a directory")
		}

		if fileInfo.Size() > maxPayloadSize {
			return fmt.Errorf("payload file size exceeds the limit of 2MB")
		}
	}

	// If execute_payload is true, payload_path must be provided
	if i.ExecutePayload && i.PayloadPath == "" {
		return fmt.Errorf("payload_path is required when execute_payload is true")
	}

	// If payload_path is provided, provision must be true
	if i.PayloadPath != "" && !i.Provision {
		return fmt.Errorf("provision must be true when payload_path is provided")
	}

	// Validate the instance name
	if i.Name != "" {
		if err := validateHostname(i.Name); err != nil {
			return fmt.Errorf("invalid instance name: %w", err)
		}
	}

	// Validate network configuration if provided
	if i.Network != nil {
		if i.Network.ProfileID < 0 {
			return fmt.Errorf("network profile_id must be a positive number")
		}
		if i.Network.Bandwidth < 0 {
			return fmt.Errorf("network bandwidth must be a positive number")
		}
		for _, rulesetID := range i.Network.FirewallRulesets {
			if rulesetID < 0 {
				return fmt.Errorf("firewall ruleset ID must be a positive number")
			}
		}
	}

	if len(i.Volumes) == 0 {
		return fmt.Errorf("at least one volume configuration is required")
	}

	// Validate volumes if present
	for j, vol := range i.Volumes {
		if err := ValidateVolume(&vol, i.Region); err != nil {
			return fmt.Errorf("invalid volume configuration at index %d: %w", j, err)
		}
	}

	i.SSHKeyName = strings.ToLower(i.SSHKeyName)
	return nil
}
