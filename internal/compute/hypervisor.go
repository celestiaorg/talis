package compute

import (
	"context"
	"fmt"
	"net/http"

	"github.com/celestiaorg/talis/internal/types"
)

// HypervisorService handles hypervisor-related operations
type HypervisorService struct {
	client *Client
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

// Hypervisor represents a VirtFusion hypervisor
type Hypervisor struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	Status       string           `json:"status,omitempty"`
	IP           string           `json:"ip"`
	IPAlt        string           `json:"ipAlt"`
	Hostname     string           `json:"hostname"`
	Port         int              `json:"port"`
	SSHPort      int              `json:"sshPort"`
	Maintenance  bool             `json:"maintenance"`
	Enabled      bool             `json:"enabled"`
	NFType       int              `json:"nfType"`
	MaxServers   int              `json:"maxServers"`
	MaxCPU       int              `json:"maxCpu"`
	MaxMemory    int              `json:"maxMemory"`
	Commissioned int              `json:"commissioned"`
	Group        *HypervisorGroup `json:"group"`
}

// newHypervisorService creates a new hypervisor service
func newHypervisorService(client *Client) *HypervisorService {
	return &HypervisorService{client: client}
}

// List returns a list of all hypervisors
func (s *HypervisorService) List(ctx context.Context) ([]*types.Hypervisor, *types.APIResponse, error) {
	resp, err := s.client.doRequest(ctx, http.MethodGet, "compute/hypervisors", nil)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Data []*types.Hypervisor `json:"data"`
	}
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, nil, err
	}

	return result.Data, &types.APIResponse{StatusCode: resp.StatusCode}, nil
}

// Get returns details of a specific hypervisor
func (s *HypervisorService) Get(ctx context.Context, id int) (*types.Hypervisor, *types.APIResponse, error) {
	resp, err := s.client.doRequest(ctx, http.MethodGet, fmt.Sprintf("compute/hypervisors/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Data *types.Hypervisor `json:"data"`
	}
	if err := s.client.parseResponse(resp, &result); err != nil {
		return nil, nil, err
	}

	return result.Data, &types.APIResponse{StatusCode: resp.StatusCode}, nil
}
