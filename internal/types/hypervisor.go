package types

// Server represents a VirtFusion server
type Server struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Hostname    string `json:"hostname"`
	Description string `json:"description"`
	Memory      int    `json:"memory"`
	CPU         int    `json:"cpu"`
	Disk        int    `json:"disk"`
	IPAddress   string `json:"ip_address"`
}

// ServerCreateRequest represents a request to create a new server
type ServerCreateRequest struct {
	Name                       string   `json:"name"`
	Description                string   `json:"description,omitempty"`
	Memory                     int      `json:"memory"`
	CPU                        int      `json:"cpuCores"`
	Disk                       int      `json:"storage"`
	Image                      string   `json:"image,omitempty"`
	HypervisorID               int      `json:"hypervisorId"`
	HypervisorGroup            string   `json:"hypervisorGroup,omitempty"` // Name of the hypervisor group (e.g. "Prague")
	SSHKeys                    []string `json:"sshKeys,omitempty"`
	UserData                   string   `json:"userData,omitempty"`
	PackageID                  int      `json:"packageId"`
	UserID                     int      `json:"userId,omitempty"`
	IPv4                       int      `json:"ipv4"`
	Traffic                    int      `json:"traffic"`
	NetworkSpeedInbound        int      `json:"networkSpeedInbound"`
	NetworkSpeedOutbound       int      `json:"networkSpeedOutbound"`
	StorageProfile             int      `json:"storageProfile"`
	NetworkProfile             int      `json:"networkProfile"`
	FirewallRulesets           []int    `json:"firewallRulesets"`
	HypervisorAssetGroups      []int    `json:"hypervisorAssetGroups,omitempty"`
	AdditionalStorage1Enable   bool     `json:"additionalStorage1Enable,omitempty"`
	AdditionalStorage2Enable   bool     `json:"additionalStorage2Enable,omitempty"`
	AdditionalStorage1Profile  int      `json:"additionalStorage1Profile,omitempty"`
	AdditionalStorage2Profile  int      `json:"additionalStorage2Profile,omitempty"`
	AdditionalStorage1Capacity int      `json:"additionalStorage1Capacity,omitempty"`
	AdditionalStorage2Capacity int      `json:"additionalStorage2Capacity,omitempty"`
}

// APIResponse represents a VirtFusion API response
type APIResponse struct {
	StatusCode int
	Headers    map[string][]string
	Error      string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
}

// Network represents a VirtFusion network
type Network struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Bridge  string `json:"bridge"`
	Primary bool   `json:"primary"`
	Default bool   `json:"default"`
}

// HypervisorGroup represents a VirtFusion hypervisor group
type HypervisorGroup struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Default          bool   `json:"default"`
	Enabled          bool   `json:"enabled"`
	DistributionType int    `json:"distributionType"`
}

// HypervisorNetwork represents a network configuration for a hypervisor
type HypervisorNetwork struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Bridge  string `json:"bridge"`
	Primary bool   `json:"primary"`
	Default bool   `json:"default"`
}

// Hypervisor represents a VirtFusion hypervisor
type Hypervisor struct {
	ID           int                 `json:"id"`
	Name         string              `json:"name"`
	Status       string              `json:"status"`
	IP           string              `json:"ip"`
	IPAlt        string              `json:"ipAlt"`
	Hostname     string              `json:"hostname"`
	Port         int                 `json:"port"`
	SSHPort      int                 `json:"sshPort"`
	Maintenance  bool                `json:"maintenance"`
	Enabled      bool                `json:"enabled"`
	NFType       int                 `json:"nfType"`
	MaxServers   int                 `json:"maxServers"`
	MaxCPU       int                 `json:"maxCpu"`
	MaxMemory    int                 `json:"maxMemory"`
	Commissioned int                 `json:"commissioned"`
	Group        *HypervisorGroup    `json:"group"`
	Networks     []HypervisorNetwork `json:"networks"`
}
