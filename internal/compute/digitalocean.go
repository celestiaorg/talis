// DigitalOcean provider for creating and managing DigitalOcean droplets
// https://github.com/digitalocean/godo/blob/main/droplets.go#L18
package compute

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

// DOClient interface and implementations

// DOClient defines the interface for interacting with DigitalOcean API
type DOClient interface {
	Droplets() DropletService
	Keys() KeyService
}

// DefaultDOClient is the standard implementation of DOClient using godo
type DefaultDOClient struct {
	client *godo.Client
}

// Droplets returns the droplet service
func (c *DefaultDOClient) Droplets() DropletService {
	return &DefaultDropletService{service: c.client.Droplets}
}

// Keys returns the key service
func (c *DefaultDOClient) Keys() KeyService {
	return &DefaultKeyService{service: c.client.Keys}
}

// NewDOClient creates a new DefaultDOClient
func NewDOClient(token string) DOClient {
	return &DefaultDOClient{
		client: godo.NewFromToken(token),
	}
}

// DropletService interface and implementations

// DropletService defines the interface for DigitalOcean droplet operations
type DropletService interface {
	Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	Delete(ctx context.Context, id int) (*godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

// DefaultDropletService implements DropletService using godo
type DefaultDropletService struct {
	service godo.DropletsService
}

// Create creates a new droplet
func (s *DefaultDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	return s.service.Create(ctx, createRequest)
}

// CreateMultiple creates multiple droplets
func (s *DefaultDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	return s.service.CreateMultiple(ctx, createRequest)
}

// Delete deletes a droplet
func (s *DefaultDropletService) Delete(ctx context.Context, id int) (*godo.Response, error) {
	return s.service.Delete(ctx, id)
}

// Get retrieves a droplet by ID
func (s *DefaultDropletService) Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
	return s.service.Get(ctx, id)
}

// List lists all droplets
func (s *DefaultDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	return s.service.List(ctx, opt)
}

// KeyService interface and implementations

// KeyService defines the interface for DigitalOcean SSH key operations
type KeyService interface {
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// DefaultKeyService implements KeyService using godo
type DefaultKeyService struct {
	service godo.KeysService
}

// List lists all SSH keys
func (s *DefaultKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	return s.service.List(ctx, opt)
}

// DigitalOceanProvider struct and methods

// DigitalOceanProvider implements the ComputeProvider interface
type DigitalOceanProvider struct {
	doClient DOClient
}

// Exported methods with their unexported helpers

// ConfigureProvider is a no-op since we're not using Pulumi anymore
func (p *DigitalOceanProvider) ConfigureProvider(stack interface{}) error {
	return nil
}

// CreateDropletRequest creates a DropletCreateRequest with common configuration (exported for testing)
func (p *DigitalOceanProvider) CreateDropletRequest(name string, config InstanceConfig, sshKeyID int) *godo.DropletCreateRequest {
	return &godo.DropletCreateRequest{
		Name:   name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: config.Image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: sshKeyID},
		},
		Tags: append([]string{name}, config.Tags...),
		UserData: `#!/bin/bash
apt-get update
apt-get install -y python3`,
	}
}

// CreateInstance creates a new DigitalOcean droplet
func (p *DigitalOceanProvider) CreateInstance(
	ctx context.Context,
	name string,
	config InstanceConfig,
) ([]InstanceInfo, error) {
	if p.doClient == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	fmt.Printf("🚀 Creating DigitalOcean droplet: %s\n", name)
	fmt.Printf("   Region: %s, Size: %s, Image: %s\n", config.Region, config.Size, config.Image)

	// Get SSH key ID
	sshKeyID, err := p.getSSHKeyID(ctx, config.SSHKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key ID: %w", err)
	}

	// If creating multiple instances, use the batch API
	if config.NumberOfInstances > 1 {
		fmt.Printf("🔢 Creating %d instances...\n", config.NumberOfInstances)

		const maxDropletsPerBatch = 10
		var allInstances []InstanceInfo
		remainingInstances := config.NumberOfInstances
		batchNumber := 0

		for remainingInstances > 0 {
			// Calculate how many instances to create in this batch
			batchSize := remainingInstances
			if batchSize > maxDropletsPerBatch {
				batchSize = maxDropletsPerBatch
			}

			// Create names for this batch
			startIndex := batchNumber * maxDropletsPerBatch
			names := make([]string, batchSize)
			for i := 0; i < batchSize; i++ {
				names[i] = fmt.Sprintf("%s-%d", name, startIndex+i)
			}

			// Create the request
			createRequest := &godo.DropletMultiCreateRequest{
				Names:  names,
				Region: config.Region,
				Size:   config.Size,
				Image: godo.DropletCreateImage{
					Slug: config.Image,
				},
				SSHKeys: []godo.DropletCreateSSHKey{
					{ID: sshKeyID},
				},
				Tags: append([]string{names[0]}, config.Tags...),
				UserData: `#!/bin/bash
apt-get update
apt-get install -y python3`,
			}

			fmt.Printf("🚀 Creating batch %d of droplets (%d instances)...\n", batchNumber+1, batchSize)
			droplets, _, err := p.doClient.Droplets().CreateMultiple(ctx, createRequest)
			if err != nil {
				fmt.Printf("❌ Failed to create droplets in batch %d: %v\n", batchNumber+1, err)
				return nil, fmt.Errorf("failed to create droplets: %w", err)
			}

			// Wait for all droplets in this batch to get their IPs and collect information
			var batchInstances []InstanceInfo

			for _, droplet := range droplets {
				fmt.Printf("⏳ Waiting for droplet %s to get an IP address...\n", droplet.Name)
				ip, err := p.waitForIP(ctx, droplet.ID, 10)
				if err != nil {
					fmt.Printf("⚠️ Warning: Failed to get IP for droplet %s: %v\n", droplet.Name, err)
					continue
				}

				instance := InstanceInfo{
					ID:       fmt.Sprintf("%d", droplet.ID),
					Name:     droplet.Name,
					PublicIP: ip,
					Provider: "digitalocean",
					Region:   config.Region,
					Size:     config.Size,
				}
				batchInstances = append(batchInstances, instance)
			}

			allInstances = append(allInstances, batchInstances...)
			remainingInstances -= batchSize
			batchNumber++
		}

		return allInstances, nil
	}

	// Create a single droplet
	createRequest := &godo.DropletCreateRequest{
		Name:   name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: config.Image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: sshKeyID},
		},
		Tags: append([]string{name}, config.Tags...),
		UserData: `#!/bin/bash
apt-get update
apt-get install -y python3`,
	}

	// Create the droplet
	droplet, _, err := p.doClient.Droplets().Create(ctx, createRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create droplet: %w", err)
	}

	fmt.Printf("✅ Droplet created with ID: %d\n", droplet.ID)

	// Wait for the droplet to get an IP address
	ip, err := p.waitForIP(ctx, droplet.ID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet IP: %w", err)
	}

	instance := InstanceInfo{
		ID:       fmt.Sprintf("%d", droplet.ID),
		Name:     droplet.Name,
		PublicIP: ip,
		Provider: "digitalocean",
		Region:   config.Region,
		Size:     config.Size,
	}

	return []InstanceInfo{instance}, nil
}

// DeleteInstance deletes a DigitalOcean droplet
func (p *DigitalOceanProvider) DeleteInstance(ctx context.Context, name string, region string) error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}

	fmt.Printf("🗑️ Deleting DigitalOcean droplet: %s in region %s\n", name, region)

	// List all droplets to find the one with our name in the specific region
	droplets, _, err := p.doClient.Droplets().List(ctx, &godo.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list droplets: %w", err)
	}

	// Find the droplet by name and region
	var dropletID int
	var found bool
	for _, d := range droplets {
		if d.Name == name && d.Region.Slug == region {
			dropletID = d.ID
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("⚠️ Droplet %s in region %s not found, nothing to delete\n", name, region)
		return nil
	}

	// Delete the droplet
	fmt.Printf("🗑️ Deleting droplet with ID: %d\n", dropletID)
	_, err = p.doClient.Droplets().Delete(ctx, dropletID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %w", err)
	}

	// Wait for the droplet to be fully deleted
	return p.waitForDeletion(ctx, name, region, 10)
}

// GetEnvironmentVars returns the environment variables needed for DigitalOcean
func (p *DigitalOceanProvider) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"DIGITALOCEAN_TOKEN": os.Getenv("DIGITALOCEAN_TOKEN"),
	}
}

// GetSSHKeyID gets the ID of an SSH key by its name (exported for testing)
func (p *DigitalOceanProvider) GetSSHKeyID(ctx context.Context, keyName string) (int, error) {
	return p.getSSHKeyID(ctx, keyName)
}

// getSSHKeyID gets the ID of an SSH key by its name
func (p *DigitalOceanProvider) getSSHKeyID(ctx context.Context, keyName string) (int, error) {
	if p.doClient == nil {
		return 0, fmt.Errorf("client not initialized")
	}

	fmt.Printf("🔑 Looking up SSH key: %s\n", keyName)

	// List all SSH keys
	keys, _, err := p.doClient.Keys().List(ctx, &godo.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list SSH keys: %w", err)
	}
	if len(keys) == 0 {
		return 0, fmt.Errorf("no SSH keys found")
	}

	// Find the key by name
	for _, key := range keys {
		if key.Name == keyName {
			fmt.Printf("✅ Found SSH key '%s' with ID: %d\n", keyName, key.ID)
			return key.ID, nil
		}
	}

	// If we get here, print available keys to help with diagnosis
	fmt.Println("Available SSH keys:")
	for _, key := range keys {
		fmt.Printf("  - %s (ID: %d)\n", key.Name, key.ID)
	}

	return 0, fmt.Errorf("SSH key '%s' not found", keyName)
}

// SetClient sets the DOClient for testing purposes
func (p *DigitalOceanProvider) SetClient(client DOClient) {
	p.doClient = client
}

// ValidateCredentials validates the DigitalOcean credentials
func (p *DigitalOceanProvider) ValidateCredentials() error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}
	return nil
}

// WaitForDeletion waits for a droplet to be fully deleted (exported for testing)
func (p *DigitalOceanProvider) WaitForDeletion(ctx context.Context, name string, region string, maxRetries int) error {
	return p.waitForDeletion(ctx, name, region, maxRetries)
}

// waitForDeletion waits for a droplet to be fully deleted
func (p *DigitalOceanProvider) waitForDeletion(ctx context.Context, name string, region string, maxRetries int) error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}

	fmt.Printf("⏳ Waiting for droplet %s in region %s to be deleted...\n", name, region)
	for i := 0; i < maxRetries; i++ {
		// Try to list the droplet
		droplets, _, err := p.doClient.Droplets().List(ctx, &godo.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list droplets: %w", err)
		}

		// Check if the droplet still exists in the specific region
		found := false
		for _, d := range droplets {
			if d.Name == name && d.Region.Slug == region {
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("✅ Confirmed droplet %s in region %s has been deleted\n", name, region)
			return nil
		}

		fmt.Printf("⏳ Droplet %s in region %s still exists, retrying in 5 seconds (attempt %d/%d)...\n",
			name, region, i+1, maxRetries)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("droplet %s in region %s still exists after %d retries", name, region, maxRetries)
}

// WaitForIP waits for a droplet to get an IP address (exported for testing)
func (p *DigitalOceanProvider) WaitForIP(ctx context.Context, dropletID int, maxRetries int) (string, error) {
	return p.waitForIP(ctx, dropletID, maxRetries)
}

// waitForIP waits for a droplet to get an IP address
func (p *DigitalOceanProvider) waitForIP(
	ctx context.Context,
	dropletID int,
	maxRetries int,
) (string, error) {
	if p.doClient == nil {
		return "", fmt.Errorf("client not initialized")
	}

	fmt.Println("⏳ Waiting for droplet to get an IP address...")
	for i := 0; i < maxRetries; i++ {
		d, _, err := p.doClient.Droplets().Get(ctx, dropletID)
		if err != nil {
			return "", fmt.Errorf("failed to get droplet details: %w", err)
		}

		// Get the public IPv4 address
		if d != nil && d.Networks != nil {
			for _, network := range d.Networks.V4 {
				if network.Type == "public" {
					ip := network.IPAddress
					fmt.Printf("📍 Found public IP for droplet: %s\n", ip)
					return ip, nil
				}
			}
		}

		fmt.Printf("⏳ IP not assigned yet, retrying in 10 seconds (attempt %d/%d)...\n", i+1, maxRetries)
		time.Sleep(10 * time.Second)
	}

	return "", fmt.Errorf("droplet created but no public IP found after %d retries", maxRetries)
}

// Functions related to DigitalOceanProvider

// NewDigitalOceanProvider creates a new DigitalOcean provider instance
func NewDigitalOceanProvider() (*DigitalOceanProvider, error) {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}

	// Create DigitalOcean API client
	doClient := NewDOClient(token)

	return &DigitalOceanProvider{
		doClient: doClient,
	}, nil
}
