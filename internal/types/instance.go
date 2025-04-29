// Package types provides type definitions for the application
package types

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/celestiaorg/talis/internal/db/models"
)

// maxUploadSize is the maximum size of a payload file in bytes. It applies to both the payload and tar archive.
const maxUploadSize = 2 * 1024 * 1024 // 2MB

// InstanceRequest represents an RPC request for a single instance
// NOTE: These should be cleaned up and replaced with specific RPC request types
type InstanceRequest struct {
	// DB Model Data - User Defined
	Name     string            `json:"name"`     // Name for instance, or instances if number_of_instances > 1
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
	SSHKeyName        string         `json:"ssh_key_name"`               // Name of the SSH key to use
	NumberOfInstances int            `json:"number_of_instances"`        // Number of instances to create
	Provision         bool           `json:"provision"`                  // Whether to run Ansible provisioning
	PayloadPath       string         `json:"payload_path,omitempty"`     // Local path to the payload script on the API server
	ExecutePayload    bool           `json:"execute_payload,omitempty"`  // Whether to execute the payload after copying
	TarArchivePath    string         `json:"tar_archive_path,omitempty"` // Local path to the tar archive on the API server
	Volumes           []VolumeConfig `json:"volumes"`                    // Optional volumes to attach

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
type DeleteInstancesRequest struct {
	OwnerID       uint     `json:"owner_id" validate:"required"`             // Owner ID
	ProjectName   string   `json:"project_name" validate:"required"`         // Project name
	InstanceNames []string `json:"instance_names" validate:"required,min=1"` // Instances to delete
}

// Validate validates the instance configuration
func (i *InstanceRequest) Validate() error {
	// Validate Metadata
	if i.Name == "" {
		return fmt.Errorf("instance name is required")
	}
	if err := validateHostname(i.Name); err != nil {
		return fmt.Errorf("invalid instance name: %w", err)
	}
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
		if err := validateUpload(i.PayloadPath); err != nil {
			return fmt.Errorf("invalid payload_path: %w", err)
		}
		// If payload_path is provided, provision must be true
		if !i.Provision {
			return fmt.Errorf("provision must be true when payload_path is provided")
		}
	}

	// If execute_payload is true, payload_path must be provided
	if i.ExecutePayload && i.PayloadPath == "" {
		return fmt.Errorf("payload_path is required when execute_payload is true")
	}

	// Validate tar archive path if provided
	if i.TarArchivePath != "" {
		if err := validateUpload(i.TarArchivePath); err != nil {
			return fmt.Errorf("invalid tar_archive_path: %w", err)
		}
		// If tar_archive_path is provided, provision must be true
		if !i.Provision {
			return fmt.Errorf("provision must be true when tar_archive_path is provided")
		}
	}

	// Confirm an action is provided
	if i.Action == "" {
		return fmt.Errorf("action is required")
	}

	return nil
}

func validateUpload(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("must be an absolute path")
	}

	// Clean the path
	path = filepath.Clean(path)

	// Check file existence and size
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("error accessing file: %w", err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("cannot be a directory")
	}

	if fileInfo.Size() > maxUploadSize {
		return fmt.Errorf("file size exceeds the limit of 2MB")
	}
	return nil
}
