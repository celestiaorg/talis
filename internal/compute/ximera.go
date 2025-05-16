package compute

import (
	"context"
	"fmt"
	"os"

	"github.com/celestiaorg/talis/internal/types"
)

// XimeraProvider implements the Provider interface for Ximera
type XimeraProvider struct {
	client *XimeraAPIClient
}

// NewXimeraProvider creates a new Ximera provider instance
func NewXimeraProvider() (*XimeraProvider, error) {
	cfg, err := InitXimeraConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ximera config: %w", err)
	}
	client := NewXimeraAPIClient(cfg)
	return &XimeraProvider{client: client}, nil
}

// ValidateCredentials validates the Ximera credentials
func (p *XimeraProvider) ValidateCredentials() error {
	// Try to list servers as a credential check
	_, err := p.client.ListServers()
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
	packageID := p.client.config.PackageID

	// Map InstanceRequest to ximera's CreateServer
	resp, err := p.client.CreateServer(
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

	// Build the server after creation
	buildResp, err := p.client.BuildServer(resp.Data.ID, req.Image, machineName, req.SSHKeyName)
	if err != nil {
		return fmt.Errorf("failed to build ximera server: %w", err)
	}
	if buildResp == nil {
		return fmt.Errorf("build server response is nil")
	}

	// Wait for the server to be fully created (polling with timeout)
	err = p.client.WaitForServerCreation(buildResp.Data.ID, 120) // 120s timeout
	if err != nil {
		return fmt.Errorf("failed to wait for ximera server to be fully created: %w", err)
	}

	// Get the server details (extract IP here)
	server, err := p.client.GetServer(buildResp.Data.ID)
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
	return p.client.DeleteServer(providerInstanceID)
}
