package types

import (
	"fmt"
)

// XimeraConfiguration holds the API configuration
type XimeraConfiguration struct {
	APIURL       string
	APIToken     string
	UserID       int
	HypervisorID int
	PackageID    int // Default package ID to use when creating servers
}

// Validate validates the XimeraConfiguration struct
func (c *XimeraConfiguration) Validate() error {
	if c.APIURL == "" {
		return fmt.Errorf("API URL cannot be empty")
	}
	if c.APIToken == "" {
		return fmt.Errorf("API token cannot be empty")
	}
	return nil
}

// XimeraServerCreateRequest represents the request to create a server
type XimeraServerCreateRequest struct {
	PackageID    int    `json:"packageId"`
	UserID       int    `json:"userId"`
	HypervisorID int    `json:"hypervisorId"`
	IPv4         int    `json:"ipv4"`
	Name         string `json:"name,omitempty"`
	Storage      int    `json:"storage,omitempty"`
	Traffic      int    `json:"traffic,omitempty"`
	Memory       int    `json:"memory,omitempty"`
	CPUCores     int    `json:"cpuCores,omitempty"`
}

// Validate validates the XimeraServerCreateRequest struct
func (s *XimeraServerCreateRequest) Validate() error {
	if s.PackageID <= 0 {
		return fmt.Errorf("package ID must be positive")
	}
	if s.UserID <= 0 {
		return fmt.Errorf("user ID must be positive")
	}
	if s.HypervisorID <= 0 {
		return fmt.Errorf("hypervisor ID must be positive")
	}
	return nil
}

// XimeraServerConfig represents a server configuration in the batch file
type XimeraServerConfig struct {
	Name       string `json:"name"`
	Package    int    `json:"package"`
	Hypervisor int    `json:"hypervisor"`
	Memory     int    `json:"memory"`
	CPU        int    `json:"cpu"`
	Storage    int    `json:"storage"`
	Traffic    int    `json:"traffic"`
	Build      bool   `json:"build"`
	OS         int    `json:"os"`
	Hostname   string `json:"hostname"`
	Delete     bool   `json:"delete,omitempty"`
}

// Validate validates the XimeraServerConfig struct
func (s *XimeraServerConfig) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	if s.Package <= 0 {
		return fmt.Errorf("package ID must be positive")
	}
	if s.Hypervisor <= 0 {
		return fmt.Errorf("hypervisor ID must be positive")
	}
	if s.Memory <= 0 {
		return fmt.Errorf("memory must be positive")
	}
	if s.CPU <= 0 {
		return fmt.Errorf("CPU must be positive")
	}
	if s.Storage <= 0 {
		return fmt.Errorf("storage must be positive")
	}
	if s.Build && s.OS <= 0 {
		return fmt.Errorf("OS ID must be positive when build is true")
	}
	return nil
}

// XimeraServerBuildRequest represents the request to build a server
type XimeraServerBuildRequest struct {
	OperatingSystemID int    `json:"operatingSystemId"`
	Name              string `json:"name,omitempty"`
	Hostname          string `json:"hostname,omitempty"`
	SSHKeys           []int  `json:"sshKeys,omitempty"`
}

// Validate validates the XimeraServerBuildRequest struct
func (s *XimeraServerBuildRequest) Validate() error {
	if s.OperatingSystemID <= 0 {
		return fmt.Errorf("operating system ID must be positive")
	}
	return nil
}

// XimeraServerResponse represents the response from the API
// Refactored to include nested structs for network and interfaces
// so that the public IP can be unmarshalled directly.
type XimeraServerResponse struct {
	Data struct {
		ID               int    `json:"id"`
		OwnerID          int    `json:"ownerId"`
		HypervisorID     int    `json:"hypervisorId"`
		Name             string `json:"name"`
		Hostname         string `json:"hostname"`
		UUID             string `json:"uuid"`
		State            string `json:"state"`
		CommissionStatus int    `json:"commissionStatus"`
		PublicIP         string `json:"publicIp,omitempty"`
		Created          string `json:"created"`
		Updated          string `json:"updated"`

		Network struct {
			Interfaces []struct {
				IPv4 []struct {
					Address string `json:"address"`
				} `json:"ipv4"`
			} `json:"interfaces"`
		} `json:"network"`
	} `json:"data"`
}

// XimeraTemplateGroup represents a group of templates in the API response
type XimeraTemplateGroup struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Icon        string           `json:"icon"`
	Templates   []XimeraTemplate `json:"templates"`
	ID          int              `json:"id"`
}

// XimeraTemplate represents an OS template in the API response
type XimeraTemplate struct {
	ID          int    `json:"id"`          // Unique identifier for the template
	Name        string `json:"name"`        // Name of the OS template
	Version     string `json:"version"`     // Version of the OS
	Variant     string `json:"variant"`     // Variant of the OS (e.g., Desktop, Server)
	Arch        int    `json:"arch"`        // Architecture type (e.g., 1 for x86_64)
	Description string `json:"description"` // Description of the template
	Icon        string `json:"icon"`        // Icon identifier for the template
	EOL         bool   `json:"eol"`         // End of Life flag
	EOLDate     string `json:"eol_date"`    // End of Life date in ISO format
	EOLWarning  bool   `json:"eol_warning"` // Whether to show EOL warning
	DeployType  int    `json:"deploy_type"` // Deployment type identifier
	VNC         bool   `json:"vnc"`         // Whether VNC is supported
	Type        string `json:"type"`        // Template type identifier
}

// Validate validates the XimeraTemplate struct
func (t *XimeraTemplate) Validate() error {
	if t.ID <= 0 {
		return fmt.Errorf("template ID must be positive")
	}
	if t.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	return nil
}

// XimeraTemplatesResponse represents the response from the API for OS templates
type XimeraTemplatesResponse struct {
	Data []XimeraTemplateGroup `json:"data"`
}

// XimeraServersListResponse represents the response from listing servers
type XimeraServersListResponse struct {
	Data []struct {
		ID           int    `json:"id"`
		OwnerID      int    `json:"ownerId"`
		HypervisorID int    `json:"hypervisorId"`
		Name         string `json:"name"`
		Hostname     string `json:"hostname"`
		UUID         string `json:"uuid"`
		State        string `json:"state"`
	} `json:"data"`
}
