package compute

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/celestiaorg/talis/internal/config"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// MockClient implements a mock VirtFusion client for testing
type MockClient struct {
	servers     map[int64]*types.Server
	hypervisors map[int]*types.Hypervisor
	serverIDSeq int64
	mu          sync.RWMutex
	config      *config.VirtFusionConfig
}

// NewMockClient creates a new mock client
func NewMockClient() *Client {
	mock := &MockClient{
		servers:     make(map[int64]*types.Server),
		hypervisors: make(map[int]*types.Hypervisor),
		config:      &config.VirtFusionConfig{},
	}

	// Add a default hypervisor
	mock.hypervisors[1] = &types.Hypervisor{
		ID:           1,
		Name:         "mock-hypervisor-1",
		IP:           "192.168.1.1",
		IPAlt:        "",
		Hostname:     "mock-hypervisor-1",
		Port:         8892,
		SSHPort:      22,
		Maintenance:  false,
		Enabled:      true,
		NFType:       4,
		MaxServers:   0,
		MaxCPU:       64,
		MaxMemory:    32768,
		Commissioned: 3,
		Group: &types.HypervisorGroup{
			ID:               2,
			Name:             "Prague",
			Description:      "Prague Datacenter",
			Default:          true,
			Enabled:          true,
			DistributionType: 5,
		},
	}

	return &Client{
		httpClient: &http.Client{},
		config:     mock.config,
		baseURL:    "http://mock-virtfusion",
	}
}

// MockServerService implements a mock server service
type MockServerService struct {
	client *MockClient
}

// MockHypervisorService implements a mock hypervisor service
type MockHypervisorService struct {
	client *MockClient
}

func (m *MockClient) Servers() *MockServerService {
	return &MockServerService{client: m}
}

func (m *MockClient) Hypervisors() *MockHypervisorService {
	return &MockHypervisorService{client: m}
}

// List returns a list of all hypervisors
func (s *MockHypervisorService) List(ctx context.Context) ([]*types.Hypervisor, *types.APIResponse, error) {
	s.client.mu.RLock()
	defer s.client.mu.RUnlock()

	hypervisors := make([]*types.Hypervisor, 0, len(s.client.hypervisors))
	for _, h := range s.client.hypervisors {
		hypervisors = append(hypervisors, h)
	}

	logger.Debugf("Mock: Listed hypervisors: count=%d", len(hypervisors))
	return hypervisors, &types.APIResponse{StatusCode: 200}, nil
}

// Get returns details of a specific hypervisor
func (s *MockHypervisorService) Get(ctx context.Context, id int) (*types.Hypervisor, *types.APIResponse, error) {
	s.client.mu.RLock()
	defer s.client.mu.RUnlock()

	hypervisor, exists := s.client.hypervisors[id]
	if !exists {
		logger.Errorf("Mock: Hypervisor not found: id=%d", id)
		return nil, &types.APIResponse{StatusCode: 404}, fmt.Errorf("hypervisor not found")
	}

	logger.Debugf("Mock: Retrieved hypervisor: id=%d, name=%s", hypervisor.ID, hypervisor.Name)
	return hypervisor, &types.APIResponse{StatusCode: 200}, nil
}

// Create mocks server creation
func (s *MockServerService) Create(ctx context.Context, req *types.ServerCreateRequest) (*types.Server, *types.APIResponse, error) {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	// Validate required fields
	if req.PackageID == 0 {
		logger.Errorf("Mock: Missing required field: packageId")
		return nil, &types.APIResponse{StatusCode: 422}, fmt.Errorf("missing required field: packageId")
	}

	if len(req.FirewallRulesets) == 0 {
		logger.Errorf("Mock: Missing required field: firewallRulesets")
		return nil, &types.APIResponse{StatusCode: 422}, fmt.Errorf("missing required field: firewallRulesets")
	}

	// Validate hypervisor exists and belongs to the specified group
	hypervisor, exists := s.client.hypervisors[req.HypervisorID]
	if !exists {
		logger.Errorf("Mock: Invalid hypervisor ID: %d", req.HypervisorID)
		return nil, &types.APIResponse{StatusCode: 422}, fmt.Errorf("invalid hypervisor ID")
	}

	// If hypervisor group is specified, validate it matches
	if req.HypervisorGroup != "" && hypervisor.Group != nil && hypervisor.Group.Name != req.HypervisorGroup {
		logger.Errorf("Mock: Invalid hypervisor group: got %s, expected %s", req.HypervisorGroup, hypervisor.Group.Name)
		return nil, &types.APIResponse{StatusCode: 422}, fmt.Errorf("invalid hypervisor group")
	}

	s.client.serverIDSeq++
	server := &types.Server{
		ID:          s.client.serverIDSeq,
		Name:        req.Name,
		Status:      "pending",
		Memory:      req.Memory,
		CPU:         req.CPU,
		Disk:        req.Disk,
		Description: req.Description,
	}

	s.client.servers[server.ID] = server
	logger.Debugf("Mock: Created server: id=%d, name=%s", server.ID, server.Name)

	return server, &types.APIResponse{StatusCode: 201}, nil
}

// Get mocks getting server details
func (s *MockServerService) Get(ctx context.Context, id int) (*types.Server, *types.APIResponse, error) {
	s.client.mu.RLock()
	defer s.client.mu.RUnlock()

	server, exists := s.client.servers[int64(id)]
	if !exists {
		logger.Errorf("Mock: Server not found: id=%d", id)
		return nil, &types.APIResponse{StatusCode: 404}, fmt.Errorf("server not found")
	}

	logger.Debugf("Mock: Retrieved server: id=%d, name=%s", server.ID, server.Name)
	return server, &types.APIResponse{StatusCode: 200}, nil
}

// Delete mocks server deletion
func (s *MockServerService) Delete(ctx context.Context, id int) (*types.APIResponse, error) {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	if _, exists := s.client.servers[int64(id)]; !exists {
		logger.Errorf("Mock: Server not found for deletion: id=%d", id)
		return &types.APIResponse{StatusCode: 404}, fmt.Errorf("server not found")
	}

	delete(s.client.servers, int64(id))
	logger.Debugf("Mock: Deleted server: id=%d", id)

	return &types.APIResponse{StatusCode: 204}, nil
}

// List mocks listing servers
func (s *MockServerService) List(ctx context.Context) ([]*types.Server, *types.APIResponse, error) {
	s.client.mu.RLock()
	defer s.client.mu.RUnlock()

	servers := make([]*types.Server, 0, len(s.client.servers))
	for _, server := range s.client.servers {
		servers = append(servers, server)
	}

	logger.Debugf("Mock: Listed servers: count=%d", len(servers))
	return servers, &types.APIResponse{StatusCode: 200}, nil
}

// Build mocks building a server
func (s *MockServerService) Build(ctx context.Context, id int) (*types.APIResponse, error) {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	server, exists := s.client.servers[int64(id)]
	if !exists {
		logger.Errorf("Mock: Server not found for build: id=%d", id)
		return &types.APIResponse{StatusCode: 404}, fmt.Errorf("server not found")
	}

	server.Status = "building"
	logger.Debugf("Mock: Building server: id=%d", id)

	return &types.APIResponse{StatusCode: 202}, nil
}

// GetStatus mocks getting server status
func (s *MockServerService) GetStatus(ctx context.Context, id int) (string, *types.APIResponse, error) {
	s.client.mu.RLock()
	defer s.client.mu.RUnlock()

	server, exists := s.client.servers[int64(id)]
	if !exists {
		logger.Errorf("Mock: Server not found for status check: id=%d", id)
		return "", &types.APIResponse{StatusCode: 404}, fmt.Errorf("server not found")
	}

	logger.Debugf("Mock: Retrieved server status: id=%d, status=%s", id, server.Status)
	return server.Status, &types.APIResponse{StatusCode: 200}, nil
}

// Suspend mocks suspending a server
func (s *MockServerService) Suspend(ctx context.Context, id int) (*types.APIResponse, error) {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	server, exists := s.client.servers[int64(id)]
	if !exists {
		logger.Errorf("Mock: Server not found for suspension: id=%d", id)
		return &types.APIResponse{StatusCode: 404}, fmt.Errorf("server not found")
	}

	server.Status = "suspended"
	logger.Debugf("Mock: Suspended server: id=%d", id)

	return &types.APIResponse{StatusCode: 202}, nil
}

// Unsuspend mocks unsuspending a server
func (s *MockServerService) Unsuspend(ctx context.Context, id int) (*types.APIResponse, error) {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	server, exists := s.client.servers[int64(id)]
	if !exists {
		logger.Errorf("Mock: Server not found for unsuspension: id=%d", id)
		return &types.APIResponse{StatusCode: 404}, fmt.Errorf("server not found")
	}

	server.Status = "running"
	logger.Debugf("Mock: Unsuspended server: id=%d", id)

	return &types.APIResponse{StatusCode: 202}, nil
}
