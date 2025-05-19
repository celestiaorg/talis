// Package types provides type definitions for the application
package types

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/celestiaorg/talis/internal/db/models"
)

const maxPayloadSize = 2 * 1024 * 1024 // 2MB

// InstanceRequest represents an RPC request for a single instance
// NOTE: These should be cleaned up and replaced with specific RPC request types
// swagger:model
// Example: {"owner_id":1,"provider":"do","region":"nyc1","size":"s-1vcpu-1gb","image":"ubuntu-20-04-x64","tags":["webserver","production"],"project_name":"my-web-project","ssh_key_name":"my-ssh-key","number_of_instances":2,"provision":true,"payload_path":"/path/to/your/script.sh","execute_payload":true,"volumes":[{"name":"my-volume-1","size_gb":10,"mount_point":"/mnt/data"}]}
type InstanceRequest struct {
	// DB Model Data - User Defined
	OwnerID  uint              `json:"owner_id"` // Owner ID of the instance
	Provider models.ProviderID `json:"provider"` // Cloud provider (e.g., "do")
	Region   string            `json:"region"`   // Region where instances will be created
	Size     string            `json:"size"`     // Instance size/type
	Image    string            `json:"image"`    // OS image to use
	Tags     []string          `json:"tags"`     // Tags to apply to instances

	// DB Model Data - Internally set during creation
	InstanceID    uint            `json:"instance_id"`              // Instance ID
	PublicIP      string          `json:"public_ip"`                // Public IP address
	VolumeIDs     []string        `json:"volume_ids,omitempty"`     // List of attached volume IDs
	VolumeDetails []VolumeDetails `json:"volume_details,omitempty"` // Detailed information about attached volumes

	// User Defined Configs
	ProjectName       string         `json:"project_name"`
	SSHKeyName        string         `json:"ssh_key_name"`              // Name of the SSH key to use
	NumberOfInstances int            `json:"number_of_instances"`       // Number of instances to create
	Provision         bool           `json:"provision"`                 // Whether to run Ansible provisioning
	PayloadPath       string         `json:"payload_path,omitempty"`    // Local path to the payload script on the API server
	ExecutePayload    bool           `json:"execute_payload,omitempty"` // Whether to execute the payload after copying
	Volumes           []VolumeConfig `json:"volumes"`                   // Optional volumes to attach

	// Talis Server Configs - Optional
	SSHKeyType string `json:"ssh_key_type,omitempty"` // Type of the private SSH key for Ansible (e.g., "rsa", "ed25519"). Defaults to "rsa".
	SSHKeyPath string `json:"ssh_key_path,omitempty"` // Custom path to the private SSH key file for Ansible. Overrides defaults.

	// Internal Configs - Set by the Talis Server
	Action             string `json:"action"`
	ProviderInstanceID int    `json:"provider_instance_id"` // Provider-specific instance ID
	LastTaskID         uint   `json:"last_task_id"`         // ID of the last task
}

// DeleteInstanceRequest represents the request body for deleting a single instance
type DeleteInstanceRequest struct {
	InstanceID uint `json:"instance_id" validate:"required"` // Instance ID to delete
}

// DeleteInstancesRequest represents the request body for deleting instances
// swagger:model
// Example: {"owner_id":1,"project_name":"my-web-project","instance_ids":[123,456]}
type DeleteInstancesRequest struct {
	OwnerID     uint   `json:"owner_id" validate:"required"`           // Owner ID
	ProjectName string `json:"project_name" validate:"required"`       // Project name
	InstanceIDs []uint `json:"instance_ids" validate:"required,min=1"` // Instances to delete
}

// Validate validates the instance configuration
func (i *InstanceRequest) Validate() error {
	// Validate Metadata
	if i.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if i.OwnerID == 0 {
		return fmt.Errorf("owner_id is required")
	}

	// Validate User Defined Configs
	if i.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if !i.Provider.IsValid() {
		return fmt.Errorf("unsupported provider: %s", i.Provider)
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
	if i.NumberOfInstances < 1 {
		return fmt.Errorf("number_of_instances must be greater than 0")
	}
	if len(i.Volumes) == 0 {
		return fmt.Errorf("at least one volume configuration is required")
	}
	// Validate volumes if present
	for j := range i.Volumes {
		if err := ValidateVolume(&i.Volumes[j], i.Region); err != nil {
			return fmt.Errorf("invalid volume configuration at index %d: %w", j, err)
		}
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

	// Confirm an action is provided
	if i.Action == "" {
		return fmt.Errorf("action is required")
	}

	return nil
}
