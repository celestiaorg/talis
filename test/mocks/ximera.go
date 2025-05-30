// Package mocks provides mock implementations for external services and APIs used in testing
package mocks

import (
	"fmt"
	"sync"
	"time"

	ximeraTypes "github.com/celestiaorg/talis/internal/compute/types"
)

// MockXimeraAPIClient implements a mock version of XimeraAPIClient for testing
type MockXimeraAPIClient struct {
	StandardResponses *XimeraStandardResponses
	// Function mocks for different API calls
	ListServersFunc           func() (*ximeraTypes.XimeraServersListResponse, error)
	CreateServerFunc          func(name string, packageID int, storage, traffic, memory, cpuCores int) (*ximeraTypes.XimeraServerResponse, error)
	GetServerFunc             func(id int) (*ximeraTypes.XimeraServerResponse, error)
	BuildServerFunc           func(id int, osID, name, sshKey string) (*ximeraTypes.XimeraServerResponse, error)
	DeleteServerFunc          func(id int) error
	ServerExistsFunc          func(name string) (bool, int, error)
	ListTemplatesFunc         func(packageID int) (*ximeraTypes.XimeraTemplatesResponse, error)
	WaitForServerCreationFunc func(serverID int, timeoutSeconds int) error
	// State tracking for testing
	mutex        sync.Mutex
	attemptCount int
	nextID       int
}

// XimeraStandardResponses holds standard mock responses for Ximera API calls
type XimeraStandardResponses struct {
	Servers   *XimeraServerResponses
	Templates *XimeraTemplateResponses
	Errors    *XimeraErrorResponses
}

// XimeraServerResponses holds standard server-related responses
type XimeraServerResponses struct {
	DefaultServer     *ximeraTypes.XimeraServerResponse
	DefaultServerList *ximeraTypes.XimeraServersListResponse
	EmptyServerList   *ximeraTypes.XimeraServersListResponse
	ServerWithTag     *ximeraTypes.XimeraServerResponse
	ServerListWithTag *ximeraTypes.XimeraServersListResponse
}

// XimeraTemplateResponses holds standard template-related responses
type XimeraTemplateResponses struct {
	DefaultTemplates *ximeraTypes.XimeraTemplatesResponse
}

// XimeraErrorResponses holds standard error responses
type XimeraErrorResponses struct {
	NotFoundError       error
	AuthenticationError error
	RateLimitError      error
	ServerCreationError error
	NetworkError        error
}

// Test constants for Ximera
const (
	DefaultXimeraServerID1   = 12345
	DefaultXimeraServerID2   = 12346
	DefaultXimeraServerIP1   = "203.0.113.10"
	DefaultXimeraServerIP2   = "203.0.113.11"
	DefaultXimeraServerName1 = "test-server-1"
	DefaultXimeraServerName2 = "test-server-2"
	TestTagName              = "talis-123-456"
	DefaultPackageID         = 1
	DefaultHypervisorID      = 10
	DefaultUserID            = 100
	DefaultTemplateID        = 200
)

// Common errors
var (
	ErrXimeraServerNotFound = fmt.Errorf("server not found")
	ErrXimeraUnauthorized   = fmt.Errorf("unauthorized access")
	ErrXimeraRateLimit      = fmt.Errorf("rate limit exceeded")
	ErrXimeraNetwork        = fmt.Errorf("network error")
)

// NewXimeraStandardResponses creates standard mock responses for Ximera
func NewXimeraStandardResponses() *XimeraStandardResponses {
	return &XimeraStandardResponses{
		Servers:   newXimeraServerResponses(),
		Templates: newXimeraTemplateResponses(),
		Errors:    newXimeraErrorResponses(),
	}
}

// newXimeraServerResponses creates standard server responses
func newXimeraServerResponses() *XimeraServerResponses {
	// Create default server response
	defaultServer := &ximeraTypes.XimeraServerResponse{
		Data: struct {
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
			Network          struct {
				Interfaces []struct {
					IPv4 []struct {
						Address string `json:"address"`
					} `json:"ipv4"`
				} `json:"interfaces"`
			} `json:"network"`
		}{
			ID:               DefaultXimeraServerID1,
			OwnerID:          DefaultUserID,
			HypervisorID:     DefaultHypervisorID,
			Name:             DefaultXimeraServerName1,
			Hostname:         "test-hostname",
			UUID:             "test-uuid-123",
			State:            "complete",
			CommissionStatus: 1,
			PublicIP:         DefaultXimeraServerIP1,
			Created:          time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			Updated:          time.Now().Format(time.RFC3339),
		},
	}

	// Add network interface data
	defaultServer.Data.Network.Interfaces = []struct {
		IPv4 []struct {
			Address string `json:"address"`
		} `json:"ipv4"`
	}{
		{
			IPv4: []struct {
				Address string `json:"address"`
			}{
				{Address: DefaultXimeraServerIP1},
			},
		},
	}

	// Create server with tag
	serverWithTag := &ximeraTypes.XimeraServerResponse{
		Data: defaultServer.Data,
	}
	serverWithTag.Data.ID = DefaultXimeraServerID2
	serverWithTag.Data.Name = TestTagName
	serverWithTag.Data.PublicIP = DefaultXimeraServerIP2

	// Create default server list
	defaultServerList := &ximeraTypes.XimeraServersListResponse{
		Data: []struct {
			ID           int    `json:"id"`
			OwnerID      int    `json:"ownerId"`
			HypervisorID int    `json:"hypervisorId"`
			Name         string `json:"name"`
			Hostname     string `json:"hostname"`
			UUID         string `json:"uuid"`
			State        string `json:"state"`
		}{
			{
				ID:           DefaultXimeraServerID1,
				OwnerID:      DefaultUserID,
				HypervisorID: DefaultHypervisorID,
				Name:         DefaultXimeraServerName1,
				Hostname:     "test-hostname",
				UUID:         "test-uuid-123",
				State:        "complete",
			},
			{
				ID:           DefaultXimeraServerID2,
				OwnerID:      DefaultUserID,
				HypervisorID: DefaultHypervisorID,
				Name:         DefaultXimeraServerName2,
				Hostname:     "test-hostname-2",
				UUID:         "test-uuid-456",
				State:        "complete",
			},
		},
	}

	// Create server list with tag
	serverListWithTag := &ximeraTypes.XimeraServersListResponse{
		Data: []struct {
			ID           int    `json:"id"`
			OwnerID      int    `json:"ownerId"`
			HypervisorID int    `json:"hypervisorId"`
			Name         string `json:"name"`
			Hostname     string `json:"hostname"`
			UUID         string `json:"uuid"`
			State        string `json:"state"`
		}{
			{
				ID:           DefaultXimeraServerID1,
				OwnerID:      DefaultUserID,
				HypervisorID: DefaultHypervisorID,
				Name:         DefaultXimeraServerName1,
				Hostname:     "test-hostname",
				UUID:         "test-uuid-123",
				State:        "complete",
			},
			{
				ID:           DefaultXimeraServerID2,
				OwnerID:      DefaultUserID,
				HypervisorID: DefaultHypervisorID,
				Name:         TestTagName, // Server with the tag as name
				Hostname:     "test-hostname-tag",
				UUID:         "test-uuid-tag",
				State:        "complete",
			},
		},
	}

	// Create empty server list
	emptyServerList := &ximeraTypes.XimeraServersListResponse{
		Data: []struct {
			ID           int    `json:"id"`
			OwnerID      int    `json:"ownerId"`
			HypervisorID int    `json:"hypervisorId"`
			Name         string `json:"name"`
			Hostname     string `json:"hostname"`
			UUID         string `json:"uuid"`
			State        string `json:"state"`
		}{},
	}

	return &XimeraServerResponses{
		DefaultServer:     defaultServer,
		DefaultServerList: defaultServerList,
		EmptyServerList:   emptyServerList,
		ServerWithTag:     serverWithTag,
		ServerListWithTag: serverListWithTag,
	}
}

// newXimeraTemplateResponses creates standard template responses
func newXimeraTemplateResponses() *XimeraTemplateResponses {
	defaultTemplates := &ximeraTypes.XimeraTemplatesResponse{
		Data: []ximeraTypes.XimeraTemplateGroup{
			{
				ID:          1,
				Name:        "Ubuntu",
				Description: "Ubuntu Operating Systems",
				Icon:        "ubuntu",
				Templates: []ximeraTypes.XimeraTemplate{
					{
						ID:          DefaultTemplateID,
						Name:        "Ubuntu",
						Version:     "22.04",
						Variant:     "Server",
						Arch:        1,
						Description: "Ubuntu 22.04 LTS Server",
						Icon:        "ubuntu",
						EOL:         false,
						DeployType:  1,
						VNC:         false,
						Type:        "linux",
					},
				},
			},
		},
	}

	return &XimeraTemplateResponses{
		DefaultTemplates: defaultTemplates,
	}
}

// newXimeraErrorResponses creates standard error responses
func newXimeraErrorResponses() *XimeraErrorResponses {
	return &XimeraErrorResponses{
		NotFoundError:       ErrXimeraServerNotFound,
		AuthenticationError: ErrXimeraUnauthorized,
		RateLimitError:      ErrXimeraRateLimit,
		ServerCreationError: fmt.Errorf("failed to create server"),
		NetworkError:        ErrXimeraNetwork,
	}
}

// NewMockXimeraAPIClient creates a new mock Ximera API client
func NewMockXimeraAPIClient() *MockXimeraAPIClient {
	client := &MockXimeraAPIClient{
		StandardResponses: NewXimeraStandardResponses(),
		nextID:            100000,
	}

	client.setupStandardResponses()
	return client
}

// setupStandardResponses configures the mock client with standard successful responses
func (c *MockXimeraAPIClient) setupStandardResponses() {
	c.ListServersFunc = func() (*ximeraTypes.XimeraServersListResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return c.StandardResponses.Servers.DefaultServerList, nil
	}

	c.CreateServerFunc = func(name string, packageID int, storage, traffic, memory, cpuCores int) (*ximeraTypes.XimeraServerResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		server := *c.StandardResponses.Servers.DefaultServer
		server.Data.ID = c.nextID
		c.nextID++
		c.mutex.Unlock()
		server.Data.Name = name
		return &server, nil
	}

	c.GetServerFunc = func(id int) (*ximeraTypes.XimeraServerResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		if id == DefaultXimeraServerID2 {
			return c.StandardResponses.Servers.ServerWithTag, nil
		}
		if id == 3 {
			// Handle custom test server for exact match test
			customServer := &ximeraTypes.XimeraServerResponse{
				Data: c.StandardResponses.Servers.DefaultServer.Data,
			}
			customServer.Data.ID = 3
			customServer.Data.Name = TestTagName
			customServer.Data.PublicIP = DefaultXimeraServerIP2
			return customServer, nil
		}
		return c.StandardResponses.Servers.DefaultServer, nil
	}

	c.BuildServerFunc = func(id int, osID, name, sshKey string) (*ximeraTypes.XimeraServerResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		server := *c.StandardResponses.Servers.DefaultServer
		server.Data.ID = id
		server.Data.Name = name
		return &server, nil
	}

	c.DeleteServerFunc = func(id int) error {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil
	}

	c.ServerExistsFunc = func(name string) (bool, int, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		if name == TestTagName {
			return true, DefaultXimeraServerID2, nil
		}
		return false, 0, nil
	}

	c.ListTemplatesFunc = func(packageID int) (*ximeraTypes.XimeraTemplatesResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return c.StandardResponses.Templates.DefaultTemplates, nil
	}

	c.WaitForServerCreationFunc = func(serverID int, timeoutSeconds int) error {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil
	}
}

// ResetToStandard resets all mock functions to their standard responses
func (c *MockXimeraAPIClient) ResetToStandard() {
	c.setupStandardResponses()
	c.mutex.Lock()
	c.attemptCount = 0
	c.mutex.Unlock()
}

// Mock API method implementations
func (c *MockXimeraAPIClient) ListServers() (*ximeraTypes.XimeraServersListResponse, error) {
	return c.ListServersFunc()
}

func (c *MockXimeraAPIClient) CreateServer(name string, packageID int, storage, traffic, memory, cpuCores int) (*ximeraTypes.XimeraServerResponse, error) {
	return c.CreateServerFunc(name, packageID, storage, traffic, memory, cpuCores)
}

func (c *MockXimeraAPIClient) GetServer(id int) (*ximeraTypes.XimeraServerResponse, error) {
	return c.GetServerFunc(id)
}

func (c *MockXimeraAPIClient) BuildServer(id int, osID, name, sshKey string) (*ximeraTypes.XimeraServerResponse, error) {
	return c.BuildServerFunc(id, osID, name, sshKey)
}

func (c *MockXimeraAPIClient) DeleteServer(id int) error {
	return c.DeleteServerFunc(id)
}

func (c *MockXimeraAPIClient) ServerExists(name string) (bool, int, error) {
	return c.ServerExistsFunc(name)
}

func (c *MockXimeraAPIClient) ListTemplates(packageID int) (*ximeraTypes.XimeraTemplatesResponse, error) {
	return c.ListTemplatesFunc(packageID)
}

func (c *MockXimeraAPIClient) WaitForServerCreation(serverID int, timeoutSeconds int) error {
	return c.WaitForServerCreationFunc(serverID, timeoutSeconds)
}

// GetPackageID returns the default package ID for testing
func (c *MockXimeraAPIClient) GetPackageID() int {
	return DefaultPackageID
}

// Simulation methods for testing different scenarios

// SimulateNotFound configures the client to return not found errors
func (c *MockXimeraAPIClient) SimulateNotFound() {
	c.ListServersFunc = func() (*ximeraTypes.XimeraServersListResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.NotFoundError
	}
	c.GetServerFunc = func(id int) (*ximeraTypes.XimeraServerResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.NotFoundError
	}
	c.DeleteServerFunc = func(id int) error {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return c.StandardResponses.Errors.NotFoundError
	}
}

// SimulateAuthenticationFailure configures the client to return authentication errors
func (c *MockXimeraAPIClient) SimulateAuthenticationFailure() {
	c.ListServersFunc = func() (*ximeraTypes.XimeraServersListResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.AuthenticationError
	}
	c.CreateServerFunc = func(name string, packageID int, storage, traffic, memory, cpuCores int) (*ximeraTypes.XimeraServerResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.AuthenticationError
	}
	c.GetServerFunc = func(id int) (*ximeraTypes.XimeraServerResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.AuthenticationError
	}
}

// SimulateRateLimit configures the client to return rate limit errors
func (c *MockXimeraAPIClient) SimulateRateLimit() {
	c.ListServersFunc = func() (*ximeraTypes.XimeraServersListResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.RateLimitError
	}
	c.CreateServerFunc = func(name string, packageID int, storage, traffic, memory, cpuCores int) (*ximeraTypes.XimeraServerResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.RateLimitError
	}
}

// SimulateServerCreationError configures the client to return server creation errors
func (c *MockXimeraAPIClient) SimulateServerCreationError() {
	c.CreateServerFunc = func(name string, packageID int, storage, traffic, memory, cpuCores int) (*ximeraTypes.XimeraServerResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.ServerCreationError
	}
	c.WaitForServerCreationFunc = func(serverID int, timeoutSeconds int) error {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return fmt.Errorf("timeout waiting for server creation")
	}
}

// SimulateEmptyServerList configures the client to return empty server lists
func (c *MockXimeraAPIClient) SimulateEmptyServerList() {
	c.ListServersFunc = func() (*ximeraTypes.XimeraServersListResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return c.StandardResponses.Servers.EmptyServerList, nil
	}
}

// SimulateServerListWithTag configures the client to return a server list containing the tag
func (c *MockXimeraAPIClient) SimulateServerListWithTag() {
	c.ListServersFunc = func() (*ximeraTypes.XimeraServersListResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return c.StandardResponses.Servers.ServerListWithTag, nil
	}
}

// SimulateNetworkError configures the client to return network errors
func (c *MockXimeraAPIClient) SimulateNetworkError() {
	c.ListServersFunc = func() (*ximeraTypes.XimeraServersListResponse, error) {
		c.mutex.Lock()
		c.attemptCount++
		c.mutex.Unlock()
		return nil, c.StandardResponses.Errors.NetworkError
	}
}

// GetAttemptCount returns the current attempt count
func (c *MockXimeraAPIClient) GetAttemptCount() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.attemptCount
}

// ResetAttemptCount resets the attempt counter
func (c *MockXimeraAPIClient) ResetAttemptCount() {
	c.mutex.Lock()
	c.attemptCount = 0
	c.mutex.Unlock()
}
