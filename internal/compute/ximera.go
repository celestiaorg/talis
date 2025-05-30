package compute

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	ximeraTypes "github.com/celestiaorg/talis/internal/compute/types"
	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/types"
)

// XimeraServerCache holds cached server information
type XimeraServerCache struct {
	Servers   *ximeraTypes.XimeraServersListResponse
	Timestamp time.Time
	mutex     sync.RWMutex
}

// XimeraProvider implements the Provider interface for Ximera
type XimeraProvider struct {
	client   interface{}
	cache    *XimeraServerCache
	cacheTTL time.Duration
}

// NewXimeraProvider creates a new Ximera provider instance
func NewXimeraProvider() (*XimeraProvider, error) {
	cfg, err := InitXimeraConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ximera config: %w", err)
	}
	client := NewXimeraAPIClient(cfg)
	return &XimeraProvider{
		client: client,
		cache: &XimeraServerCache{
			mutex: sync.RWMutex{},
		},
		cacheTTL: 5 * time.Minute, // 5 minute cache TTL
	}, nil
}

// SetClient sets the client for testing purposes
func (p *XimeraProvider) SetClient(client interface{}) {
	p.client = client
}

// getClient returns the client as XimeraAPIClient interface
func (p *XimeraProvider) getClient() interface {
	ListServers() (*ximeraTypes.XimeraServersListResponse, error)
	CreateServer(name string, packageID int, storage, traffic, memory, cpuCores int) (*ximeraTypes.XimeraServerResponse, error)
	GetServer(id int) (*ximeraTypes.XimeraServerResponse, error)
	BuildServer(id int, osID, name, sshKey string) (*ximeraTypes.XimeraServerResponse, error)
	DeleteServer(id int) error
	WaitForServerCreation(serverID int, timeoutSeconds int) error
	GetPackageID() int
} {
	return p.client.(interface {
		ListServers() (*ximeraTypes.XimeraServersListResponse, error)
		CreateServer(name string, packageID int, storage, traffic, memory, cpuCores int) (*ximeraTypes.XimeraServerResponse, error)
		GetServer(id int) (*ximeraTypes.XimeraServerResponse, error)
		BuildServer(id int, osID, name, sshKey string) (*ximeraTypes.XimeraServerResponse, error)
		DeleteServer(id int) error
		WaitForServerCreation(serverID int, timeoutSeconds int) error
		GetPackageID() int
	})
}

// ValidateCredentials validates the Ximera credentials
func (p *XimeraProvider) ValidateCredentials() error {
	// Try to list servers as a credential check
	_, err := p.getClient().ListServers()
	if err != nil {
		return fmt.Errorf("ximera credential validation failed: %w", err)
	}
	return nil
}

// GetEnvironmentVars returns the environment variables needed for the provider
func (p *XimeraProvider) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"XIMERA_API_URL":             os.Getenv("XIMERA_API_URL"),
		"XIMERA_API_TOKEN":           os.Getenv("XIMERA_API_TOKEN"),
		"XIMERA_USER_ID":             os.Getenv("XIMERA_USER_ID"),
		"XIMERA_HYPERVISOR_GROUP_ID": os.Getenv("XIMERA_HYPERVISOR_GROUP_ID"),
		"XIMERA_PACKAGE_ID":          os.Getenv("XIMERA_PACKAGE_ID"),
	}
}

// ConfigureProvider configures the provider with the given stack (no-op for ximera)
func (p *XimeraProvider) ConfigureProvider(_ interface{}) error {
	return nil
}

// CreateInstance creates a new instance using Ximera
func (p *XimeraProvider) CreateInstance(_ context.Context, req *types.InstanceRequest) error {
	machineName := fmt.Sprintf("%s-%s", req.ProjectName, generateRandomSuffix())

	if len(req.Volumes) == 0 {
		return fmt.Errorf("no volume details provided")
	}

	if len(req.Volumes) > 1 {
		return fmt.Errorf("only one volume is supported")
	}

	if req.Size != "" {
		fmt.Printf("Warning: 'size' parameter '%s' is not supported for Ximera provider and will be ignored. Use 'memory' and 'cpu' parameters instead.\n", req.Size)
	}

	if req.Memory == 0 {
		return fmt.Errorf("memory is required for Ximera")
	}

	if req.CPU == 0 {
		return fmt.Errorf("cpu is required for Ximera")
	}

	// we want each ximera server to have unlimited traffic
	traffic := 0
	// Use the configured package ID from the client's configuration
	packageID := p.getClient().GetPackageID()

	// Map InstanceRequest to ximera's CreateServer
	resp, err := p.getClient().CreateServer(
		machineName,
		packageID,
		req.Volumes[0].SizeGB,
		traffic,
		req.Memory,
		req.CPU,
	)
	if err != nil {
		return fmt.Errorf("failed to create ximera server: %w", err)
	}

	// Get SSH key name from environment variable
	sshKeyName := os.Getenv(constants.EnvTalisSSHKeyName)
	if sshKeyName != "" {
		logger.Debugf("🔑 Using SSH key name from environment variable: %s", sshKeyName)
	} else {
		return fmt.Errorf("environment variable %s is not set but required for Ximera", constants.EnvTalisSSHKeyName)
	}

	// Build the server after creation
	buildResp, err := p.getClient().BuildServer(resp.Data.ID, req.Image, machineName, sshKeyName)
	if err != nil {
		return fmt.Errorf("failed to build ximera server: %w", err)
	}
	if buildResp == nil {
		return fmt.Errorf("build server response is nil")
	}

	// Wait for the server to be fully created (polling with timeout)
	err = p.getClient().WaitForServerCreation(buildResp.Data.ID, 120) // 120s timeout
	if err != nil {
		return fmt.Errorf("failed to wait for ximera server to be fully created: %w", err)
	}

	// Get the server details (extract IP here)
	server, err := p.getClient().GetServer(buildResp.Data.ID)
	if err != nil {
		return fmt.Errorf("failed to get ximera server details: %w", err)
	}

	fmt.Printf("Ximera server created with ID %d and public IP %s\n", server.Data.ID, server.Data.PublicIP)

	req.ProviderInstanceID = server.Data.ID
	req.PublicIP = server.Data.PublicIP
	return nil
}

// DeleteInstance deletes an instance using Ximera
func (p *XimeraProvider) DeleteInstance(_ context.Context, providerInstanceID int) error {
	// Invalidate cache after deletion
	defer p.invalidateCache()
	return p.getClient().DeleteServer(providerInstanceID)
}

// invalidateCache invalidates the server cache
func (p *XimeraProvider) invalidateCache() {
	p.cache.mutex.Lock()
	defer p.cache.mutex.Unlock()
	p.cache.Servers = nil
	p.cache.Timestamp = time.Time{}
}

// isCacheValid checks if the cache is still valid
func (p *XimeraProvider) isCacheValid() bool {
	p.cache.mutex.RLock()
	defer p.cache.mutex.RUnlock()

	if p.cache.Servers == nil {
		return false
	}

	return time.Since(p.cache.Timestamp) < p.cacheTTL
}

// getCachedServers retrieves servers from cache with pagination handling
func (p *XimeraProvider) getCachedServers() (*ximeraTypes.XimeraServersListResponse, error) {
	// Check if cache is still valid
	if p.isCacheValid() {
		p.cache.mutex.RLock()
		defer p.cache.mutex.RUnlock()
		logger.Debugf("🗄️ Using cached Ximera servers (%d servers)", len(p.cache.Servers.Data))
		return p.cache.Servers, nil
	}

	// Cache is invalid or empty, fetch all servers with pagination
	logger.Debugf("📥 Fetching all Ximera servers (cache miss or expired)")

	allServers := &ximeraTypes.XimeraServersListResponse{
		Data: make([]struct {
			ID           int    `json:"id"`
			OwnerID      int    `json:"ownerId"`
			HypervisorID int    `json:"hypervisorId"`
			Name         string `json:"name"`
			Hostname     string `json:"hostname"`
			UUID         string `json:"uuid"`
			State        string `json:"state"`
		}, 0),
	}

	// Note: Ximera's ListServers method should handle pagination internally
	// If it doesn't, we would need to implement pagination here
	servers, err := p.getClient().ListServers()
	if err != nil {
		return nil, fmt.Errorf("failed to list ximera servers: %w", err)
	}

	allServers.Data = append(allServers.Data, servers.Data...)

	// Update cache with thread safety
	p.cache.mutex.Lock()
	p.cache.Servers = allServers
	p.cache.Timestamp = time.Now()
	p.cache.mutex.Unlock()

	logger.Debugf("📦 Cached %d Ximera servers", len(allServers.Data))
	return allServers, nil
}

// GetInstanceByTag retrieves an instance by its Talis tag from Ximera using caching
// Since Ximera doesn't support tags like DigitalOcean, we search by server name
func (p *XimeraProvider) GetInstanceByTag(_ context.Context, tag string) (*types.InstanceRequest, error) {
	logger.Debugf("🔍 Searching for Ximera server with tag: %s (using cached data)", tag)

	// Get servers from cache (will fetch if cache is invalid/empty)
	servers, err := p.getCachedServers()
	if err != nil {
		return nil, fmt.Errorf("failed to get ximera servers: %w", err)
	}

	// Search for a server with the tag in its name
	// For Ximera, we'll look for the tag as part of the server name
	for _, server := range servers.Data {
		if server.Name == tag {
			logger.Debugf("✅ Found Ximera server with tag %s: %s (ID: %d)", tag, server.Name, server.ID)

			// Get full server details to access PublicIP and other fields
			fullServer, err := p.getClient().GetServer(server.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get full server details for server %d: %w", server.ID, err)
			}

			// Convert server to InstanceRequest format
			instanceRequest := &types.InstanceRequest{
				ProviderInstanceID: fullServer.Data.ID,
				Name:               fullServer.Data.Name,
				PublicIP:           fullServer.Data.PublicIP,
			}

			return instanceRequest, nil
		}
	}

	// Instance not found
	logger.Debugf("❌ No Ximera server found with tag: %s", tag)
	return nil, fmt.Errorf("instance with tag '%s' not found", tag)
}
