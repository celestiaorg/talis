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

// DOClient defines the interface for interacting with DigitalOcean API
type DOClient interface {
	Droplets() DropletService
	Keys() KeyService
}

// DropletService defines the interface for DigitalOcean droplet operations
type DropletService interface {
	Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	Delete(ctx context.Context, id int) (*godo.Response, error)
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

// KeyService defines the interface for DigitalOcean SSH key operations
type KeyService interface {
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// DefaultDOClient is the standard implementation of DOClient using godo
type DefaultDOClient struct {
	client *godo.Client
}

// NewDOClient creates a new DefaultDOClient
func NewDOClient(token string) DOClient {
	return &DefaultDOClient{
		client: godo.NewFromToken(token),
	}
}

// Droplets returns the droplet service
func (c *DefaultDOClient) Droplets() DropletService {
	return &DefaultDropletService{service: c.client.Droplets}
}

// Keys returns the key service
func (c *DefaultDOClient) Keys() KeyService {
	return &DefaultKeyService{service: c.client.Keys}
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

// Get retrieves a droplet by ID
func (s *DefaultDropletService) Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
	return s.service.Get(ctx, id)
}

// Delete deletes a droplet
func (s *DefaultDropletService) Delete(ctx context.Context, id int) (*godo.Response, error) {
	return s.service.Delete(ctx, id)
}

// List lists all droplets
func (s *DefaultDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	return s.service.List(ctx, opt)
}

// DefaultKeyService implements KeyService using godo
type DefaultKeyService struct {
	service godo.KeysService
}

// List lists all SSH keys
func (s *DefaultKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	return s.service.List(ctx, opt)
}

// DigitalOceanProvider implements the ComputeProvider interface
type DigitalOceanProvider struct {
	DOClient DOClient // Exported for testing
}

// NewDigitalOceanProvider creates a new DigitalOcean provider instance
func NewDigitalOceanProvider() (*DigitalOceanProvider, error) {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}

	// Create DigitalOcean API client
	doClient := NewDOClient(token)

	return &DigitalOceanProvider{
		DOClient: doClient,
	}, nil
}

// ValidateCredentials validates the DigitalOcean credentials
func (p *DigitalOceanProvider) ValidateCredentials() error {
	if p.DOClient == nil {
		return fmt.Errorf("client not initialized")
	}
	return nil
}

// GetEnvironmentVars returns the environment variables needed for DigitalOcean
func (p *DigitalOceanProvider) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"DIGITALOCEAN_TOKEN": os.Getenv("DIGITALOCEAN_TOKEN"),
	}
}

// ConfigureProvider is a no-op since we're not using Pulumi anymore
func (p *DigitalOceanProvider) ConfigureProvider(stack interface{}) error {
	return nil
}

// findSSHKeyByName finds an SSH key by name from a list of keys
func findSSHKeyByName(keys []godo.Key, keyName string) (int, bool) {
	for _, key := range keys {
		if key.Name == keyName {
			return key.ID, true
		}
	}
	return 0, false
}

// printAvailableKeys prints available SSH keys for diagnostic purposes
func printAvailableKeys(keys []godo.Key) {
	if len(keys) > 0 {
		fmt.Println("Available SSH keys:")
		for _, key := range keys {
			fmt.Printf("  - %s (ID: %d)\n", key.Name, key.ID)
		}
	}
}

// getSSHKeyID gets the ID of an SSH key by its name
func (p *DigitalOceanProvider) getSSHKeyID(ctx context.Context, keyName string) (int, error) {
	if p.DOClient == nil {
		return 0, fmt.Errorf("client not initialized")
	}

	fmt.Printf("🔑 Looking up SSH key: %s\n", keyName)

	// List all SSH keys
	keys, _, err := p.DOClient.Keys().List(ctx, &godo.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	// Find the key by name
	keyID, found := findSSHKeyByName(keys, keyName)
	if found {
		fmt.Printf("✅ Found SSH key '%s' with ID: %d\n", keyName, keyID)
		return keyID, nil
	}

	// If we get here, print available keys to help with diagnosis
	printAvailableKeys(keys)

	return 0, fmt.Errorf("SSH key '%s' not found", keyName)
}

// getPublicIPFromDroplet extracts the public IP address from droplet networks
func getPublicIPFromDroplet(droplet *godo.Droplet) (string, bool) {
	if droplet == nil || droplet.Networks == nil {
		return "", false
	}

	for _, network := range droplet.Networks.V4 {
		if network.Type == "public" {
			return network.IPAddress, true
		}
	}
	return "", false
}

// waitForIP waits for a droplet to get an IP address
func (p *DigitalOceanProvider) waitForIP(
	ctx context.Context,
	dropletID int,
	maxRetries int,
) (string, error) {
	if p.DOClient == nil {
		return "", fmt.Errorf("client not initialized")
	}

	fmt.Println("⏳ Waiting for droplet to get an IP address...")
	for i := 0; i < maxRetries; i++ {
		d, _, err := p.DOClient.Droplets().Get(ctx, dropletID)
		if err != nil {
			return "", fmt.Errorf("failed to get droplet details: %w", err)
		}

		// Get the public IPv4 address
		ip, found := getPublicIPFromDroplet(d)
		if found {
			fmt.Printf("📍 Found public IP for droplet: %s\n", ip)
			return ip, nil
		}

		fmt.Printf("⏳ IP not assigned yet, retrying in 10 seconds (attempt %d/%d)...\n", i+1, maxRetries)
		time.Sleep(10 * time.Second)
	}

	return "", fmt.Errorf("droplet created but no public IP found after %d retries", maxRetries)
}

// createDropletRequest creates a DropletCreateRequest with common configuration
func (p *DigitalOceanProvider) createDropletRequest(
	name string,
	config InstanceConfig,
	sshKeyID int,
) *godo.DropletCreateRequest {
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

// createMultiDropletRequest creates a DropletMultiCreateRequest with common configuration
func (p *DigitalOceanProvider) createMultiDropletRequest(
	names []string,
	config InstanceConfig,
	sshKeyID int,
) *godo.DropletMultiCreateRequest {
	return &godo.DropletMultiCreateRequest{
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
}

// generateBatchNames generates names for a batch of droplets
func generateBatchNames(baseName string, startIndex int, count int) []string {
	names := make([]string, count)
	for i := 0; i < count; i++ {
		names[i] = fmt.Sprintf("%s-%d", baseName, startIndex+i)
	}
	return names
}

// createDropletInfo creates an InstanceInfo from a droplet and IP
func createDropletInfo(droplet godo.Droplet, ip string, region string, size string) InstanceInfo {
	return InstanceInfo{
		ID:       fmt.Sprintf("%d", droplet.ID),
		Name:     droplet.Name,
		PublicIP: ip,
		Provider: "digitalocean",
		Region:   region,
		Size:     size,
	}
}

// createMultipleDroplets creates multiple droplets using the CreateMultiple API
func (p *DigitalOceanProvider) createMultipleDroplets(
	ctx context.Context,
	name string,
	config InstanceConfig,
	sshKeyID int,
) ([]InstanceInfo, error) {
	if p.DOClient == nil {
		return nil, fmt.Errorf("client not initialized")
	}

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
		names := generateBatchNames(name, startIndex, batchSize)

		// Create the request
		createRequest := p.createMultiDropletRequest(names, config, sshKeyID)

		fmt.Printf("🚀 Creating batch %d of droplets (%d instances)...\n", batchNumber+1, batchSize)
		droplets, _, err := p.DOClient.Droplets().CreateMultiple(ctx, createRequest)
		if err != nil {
			fmt.Printf("❌ Failed to create droplets in batch %d: %v\n", batchNumber+1, err)
			return nil, fmt.Errorf("failed to create droplets: %w", err)
		}

		// Wait for all droplets in this batch to get their IPs and collect information
		batchInstances, err := p.processCreatedDroplets(ctx, droplets, config)
		if err != nil {
			fmt.Printf("⚠️ Warning: Some droplets in batch %d may not have been processed correctly: %v\n", batchNumber+1, err)
		}

		allInstances = append(allInstances, batchInstances...)
		remainingInstances -= batchSize
		batchNumber++
	}

	return allInstances, nil
}

// processCreatedDroplets waits for IPs and creates instance info for created droplets
func (p *DigitalOceanProvider) processCreatedDroplets(
	ctx context.Context,
	droplets []godo.Droplet,
	config InstanceConfig,
) ([]InstanceInfo, error) {
	var instances []InstanceInfo
	var errors []error

	for _, droplet := range droplets {
		fmt.Printf("⏳ Waiting for droplet %s to get an IP address...\n", droplet.Name)
		ip, err := p.waitForIP(ctx, droplet.ID, 10)
		if err != nil {
			fmt.Printf("⚠️ Warning: Failed to get IP for droplet %s: %v\n", droplet.Name, err)
			errors = append(errors, err)
			continue
		}

		instance := createDropletInfo(droplet, ip, config.Region, config.Size)
		instances = append(instances, instance)
	}

	if len(errors) > 0 {
		return instances, fmt.Errorf("failed to process %d droplets", len(errors))
	}

	return instances, nil
}

// createSingleDroplet creates a single droplet
func (p *DigitalOceanProvider) createSingleDroplet(
	ctx context.Context,
	name string,
	config InstanceConfig,
	sshKeyID int,
) (InstanceInfo, error) {
	// Create the request
	createRequest := p.createDropletRequest(name, config, sshKeyID)

	// Create the droplet
	droplet, _, err := p.DOClient.Droplets().Create(ctx, createRequest)
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("failed to create droplet: %w", err)
	}

	fmt.Printf("✅ Droplet created with ID: %d\n", droplet.ID)

	// Wait for the droplet to get an IP address
	ip, err := p.waitForIP(ctx, droplet.ID, 10)
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("failed to get droplet IP: %w", err)
	}

	return createDropletInfo(*droplet, ip, config.Region, config.Size), nil
}

// CreateInstance creates a new DigitalOcean droplet
func (p *DigitalOceanProvider) CreateInstance(
	ctx context.Context,
	name string,
	config InstanceConfig,
) ([]InstanceInfo, error) {
	if p.DOClient == nil {
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
		return p.createMultipleDroplets(ctx, name, config, sshKeyID)
	}

	// Create a single droplet
	instance, err := p.createSingleDroplet(ctx, name, config, sshKeyID)
	if err != nil {
		return nil, err
	}

	return []InstanceInfo{instance}, nil
}

// findDropletByNameAndRegion finds a droplet by name and region
func findDropletByNameAndRegion(droplets []godo.Droplet, name string, region string) (int, bool) {
	for _, d := range droplets {
		if d.Name == name && d.Region.Slug == region {
			return d.ID, true
		}
	}
	return 0, false
}

// waitForDeletion waits for a droplet to be fully deleted
func (p *DigitalOceanProvider) waitForDeletion(ctx context.Context, name string, region string, maxRetries int) error {
	if p.DOClient == nil {
		return fmt.Errorf("client not initialized")
	}

	fmt.Printf("⏳ Waiting for droplet %s in region %s to be deleted...\n", name, region)
	for i := 0; i < maxRetries; i++ {
		// Try to list the droplet
		droplets, _, err := p.DOClient.Droplets().List(ctx, &godo.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list droplets: %w", err)
		}

		// Check if the droplet still exists in the specific region
		_, found := findDropletByNameAndRegion(droplets, name, region)
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

// DeleteInstance deletes a DigitalOcean droplet
func (p *DigitalOceanProvider) DeleteInstance(ctx context.Context, name string, region string) error {
	if p.DOClient == nil {
		return fmt.Errorf("client not initialized")
	}

	fmt.Printf("🗑️ Deleting DigitalOcean droplet: %s in region %s\n", name, region)

	// List all droplets to find the one with our name in the specific region
	droplets, _, err := p.DOClient.Droplets().List(ctx, &godo.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list droplets: %w", err)
	}

	// Find the droplet by name and region
	dropletID, found := findDropletByNameAndRegion(droplets, name, region)
	if !found {
		fmt.Printf("⚠️ Droplet %s in region %s not found, nothing to delete\n", name, region)
		return nil
	}

	// Delete the droplet
	fmt.Printf("🗑️ Deleting droplet with ID: %d\n", dropletID)
	_, err = p.DOClient.Droplets().Delete(ctx, dropletID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %w", err)
	}

	// Wait for the droplet to be fully deleted
	return p.waitForDeletion(ctx, name, region, 10)
}

// GetSSHKeyID gets the ID of an SSH key by its name (exported for testing)
func (p *DigitalOceanProvider) GetSSHKeyID(ctx context.Context, keyName string) (int, error) {
	return p.getSSHKeyID(ctx, keyName)
}

// WaitForIP waits for a droplet to get an IP address (exported for testing)
func (p *DigitalOceanProvider) WaitForIP(ctx context.Context, dropletID int, maxRetries int) (string, error) {
	return p.waitForIP(ctx, dropletID, maxRetries)
}

// WaitForDeletion waits for a droplet to be fully deleted (exported for testing)
func (p *DigitalOceanProvider) WaitForDeletion(ctx context.Context, name string, region string, maxRetries int) error {
	return p.waitForDeletion(ctx, name, region, maxRetries)
}

// CreateDropletRequest creates a DropletCreateRequest with common configuration (exported for testing)
func (p *DigitalOceanProvider) CreateDropletRequest(name string, config InstanceConfig, sshKeyID int) *godo.DropletCreateRequest {
	return p.createDropletRequest(name, config, sshKeyID)
}
