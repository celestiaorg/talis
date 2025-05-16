package types

// Configuration holds the API configuration
type Configuration struct {
	APIURL       string
	APIToken     string
	UserID       int
	HypervisorID int
}

// ServerCreateRequest represents the request to create a server
type ServerCreateRequest struct {
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

// ServerConfig represents a server configuration in the batch file
type ServerConfig struct {
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

// ServerBuildRequest represents the request to build a server
type ServerBuildRequest struct {
	OperatingSystemID int    `json:"operatingSystemId"`
	Name              string `json:"name,omitempty"`
	Hostname          string `json:"hostname,omitempty"`
	SSHKeys           []int  `json:"sshKeys,omitempty"`
}

// ServerResponse represents the response from the API
// Refactored to include nested structs for network and interfaces
// so that the public IP can be unmarshalled directly.
type ServerResponse struct {
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

// TemplateGroup represents a group of templates in the API response
type TemplateGroup struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Templates   []Template `json:"templates"`
	ID          int        `json:"id"`
}

// Template represents an OS template in the API response
type Template struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Variant     string `json:"variant"`
	Arch        int    `json:"arch"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	EOL         bool   `json:"eol"`
	EOLDate     string `json:"eol_date"`
	EOLWarning  bool   `json:"eol_warning"`
	DeployType  int    `json:"deploy_type"`
	VNC         bool   `json:"vnc"`
	Type        string `json:"type"`
}

// TemplatesResponse represents the response from the API for OS templates
type TemplatesResponse struct {
	Data []TemplateGroup `json:"data"`
}

// ServersListResponse represents the response from listing servers
type ServersListResponse struct {
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
