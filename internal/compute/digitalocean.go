// DigitalOcean provider for creating and managing DigitalOcean droplets
// https://github.com/digitalocean/godo/blob/main/droplets.go#L18
package compute

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

// DigitalOceanProvider implements the ComputeProvider interface
type DigitalOceanProvider struct {
	doClient *godo.Client
}

// NewDigitalOceanProvider creates a new DigitalOcean provider instance
func NewDigitalOceanProvider() (*DigitalOceanProvider, error) {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}

	// Create DigitalOcean API client
	doClient := godo.NewFromToken(token)

	return &DigitalOceanProvider{
		doClient: doClient,
	}, nil
}

// ValidateCredentials validates the DigitalOcean credentials
func (p *DigitalOceanProvider) ValidateCredentials() error {
	if p.doClient == nil {
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

// getSSHKeyID gets the ID of an SSH key by its name
func (p *DigitalOceanProvider) getSSHKeyID(ctx context.Context, keyName string) (int, error) {
	if p.doClient == nil {
		return 0, fmt.Errorf("client not initialized")
	}

	log.Printf("🔑 Looking up SSH key: %s", keyName)

	// List all SSH keys
	keys, _, err := p.doClient.Keys.List(ctx, &godo.ListOptions{})
	if err != nil {
		log.Printf("❌ Failed to list SSH keys: %v", err)
		return 0, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	// Find the key by name
	for _, key := range keys {
		if key.Name == keyName {
			log.Printf("✅ Found SSH key '%s' with ID: %d", keyName, key.ID)
			return key.ID, nil
		}
	}

	// If we get here, print available keys to help with diagnosis
	if len(keys) > 0 {
		log.Printf("Available SSH keys:")
		for _, key := range keys {
			log.Printf("  - %s (ID: %d)", key.Name, key.ID)
		}
	}

	return 0, fmt.Errorf("SSH key '%s' not found", keyName)
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

	log.Printf("⏳ Waiting for droplet to get an IP address...")
	for i := 0; i < maxRetries; i++ {
		d, _, err := p.doClient.Droplets.Get(ctx, dropletID)
		if err != nil {
			log.Printf("❌ Failed to get droplet details: %v", err)
			return "", fmt.Errorf("failed to get droplet details: %w", err)
		}

		// Get the public IPv4 address
		for _, network := range d.Networks.V4 {
			if network.Type == "public" {
				ip := network.IPAddress
				log.Printf("📍 Found public IP for droplet: %s", ip)
				return ip, nil
			}
		}

		log.Printf("⏳ IP not assigned yet, retrying in 10 seconds (attempt %d/%d)...", i+1, maxRetries)
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

// createMultipleDroplets creates multiple droplets using the CreateMultiple API
func (p *DigitalOceanProvider) createMultipleDroplets(
	ctx context.Context,
	name string,
	config InstanceConfig,
	sshKeyID int,
) ([]InstanceInfo, error) {
	if p.doClient == nil {
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
		names := make([]string, batchSize)
		startIndex := batchNumber * maxDropletsPerBatch
		for i := 0; i < batchSize; i++ {
			names[i] = fmt.Sprintf("%s-%d", name, startIndex+i)
		}

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
			Tags: append([]string{name}, config.Tags...),
			UserData: `#!/bin/bash
apt-get update
apt-get install -y python3`,
		}

		log.Printf("🚀 Creating batch %d of droplets (%d instances)...", batchNumber+1, batchSize)
		droplets, _, err := p.doClient.Droplets.CreateMultiple(ctx, createRequest)
		if err != nil {
			log.Printf("❌ Failed to create droplets in batch %d: %v", batchNumber+1, err)
			return nil, fmt.Errorf("failed to create droplets: %w", err)
		}

		// Wait for all droplets in this batch to get their IPs and collect information
		for _, droplet := range droplets {
			log.Printf("⏳ Waiting for droplet %s to get an IP address...", droplet.Name)
			ip, err := p.waitForIP(ctx, droplet.ID, 10)
			if err != nil {
				// Log the error but continue with other droplets
				log.Printf("⚠️ Warning: Failed to get IP for droplet %s: %v", droplet.Name, err)
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
			allInstances = append(allInstances, instance)
			log.Printf("✅ Droplet %s is ready with IP: %s", droplet.Name, ip)
		}

		remainingInstances -= batchSize
		batchNumber++

		// If we have more instances to create, add a small delay between batches
		if remainingInstances > 0 {
			log.Printf("⏳ Waiting before creating next batch... (%d instances remaining)", remainingInstances)
			time.Sleep(5 * time.Second)
		}
	}

	log.Printf("✅ Created all %d droplets with base name: %s", len(allInstances), name)
	return allInstances, nil
}

// createSingleDroplet creates a single droplet
func (p *DigitalOceanProvider) createSingleDroplet(
	ctx context.Context,
	name string,
	config InstanceConfig,
	sshKeyID int,
) (InstanceInfo, error) {
	if p.doClient == nil {
		return InstanceInfo{}, fmt.Errorf("client not initialized")
	}

	// Use consistent naming with index for single droplet
	dropletName := fmt.Sprintf("%s-0", name)
	createRequest := p.createDropletRequest(dropletName, config, sshKeyID)

	// Create the droplet
	droplet, _, err := p.doClient.Droplets.Create(ctx, createRequest)
	if err != nil {
		log.Printf("❌ Failed to create droplet: %v", err)
		return InstanceInfo{}, fmt.Errorf("failed to create droplet: %w", err)
	}

	// Wait for droplet to get an IP
	ip, err := p.waitForIP(ctx, droplet.ID, 10)
	if err != nil {
		return InstanceInfo{}, err
	}

	log.Printf("✅ Droplet creation completed: %s (IP: %s)", dropletName, ip)
	return InstanceInfo{
		ID:       fmt.Sprintf("%d", droplet.ID),
		Name:     dropletName,
		PublicIP: ip,
		Provider: "digitalocean",
		Region:   config.Region,
		Size:     config.Size,
	}, nil
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

	log.Printf("🚀 Creating DigitalOcean droplet(s): %s", name)
	log.Printf("  Region: %s", config.Region)
	log.Printf("  Size: %s", config.Size)
	log.Printf("  Image: %s", config.Image)
	log.Printf("  Number of instances: %d", config.NumberOfInstances)

	// Get SSH key ID
	sshKeyID, err := p.getSSHKeyID(ctx, config.SSHKeyID)
	if err != nil {
		log.Printf("❌ Failed to get SSH key: %v", err)
		return nil, fmt.Errorf("failed to get SSH key: %w", err)
	}

	// Create single or multiple droplets based on configuration
	if config.NumberOfInstances > 1 {
		return p.createMultipleDroplets(ctx, name, config, sshKeyID)
	}

	// For single instance, wrap the result in a slice for consistent interface
	instance, err := p.createSingleDroplet(ctx, name, config, sshKeyID)
	if err != nil {
		return nil, err
	}
	return []InstanceInfo{instance}, nil
}

// waitForDeletion waits for a droplet to be fully deleted
func (p *DigitalOceanProvider) waitForDeletion(ctx context.Context, name string, region string, maxRetries int) error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}

	log.Printf("⏳ Waiting for droplet %s in region %s to be deleted...", name, region)
	for i := 0; i < maxRetries; i++ {
		// Try to list the droplet
		droplets, _, err := p.doClient.Droplets.List(ctx, &godo.ListOptions{})
		if err != nil {
			log.Printf("❌ Failed to list droplets: %v", err)
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
			log.Printf("✅ Confirmed droplet %s in region %s has been deleted", name, region)
			return nil
		}

		log.Printf("⏳ Droplet %s in region %s still exists, retrying in 5 seconds (attempt %d/%d)...",
			name, region, i+1, maxRetries)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("droplet %s in region %s still exists after %d retries", name, region, maxRetries)
}

// DeleteInstance deletes a DigitalOcean droplet
func (p *DigitalOceanProvider) DeleteInstance(ctx context.Context, name string, region string) error {
	if p.doClient == nil {
		return fmt.Errorf("client not initialized")
	}

	log.Printf("🗑️ Deleting DigitalOcean droplet: %s in region %s", name, region)

	// List all droplets to find the one with our name in the specific region
	droplets, _, err := p.doClient.Droplets.List(ctx, &godo.ListOptions{})
	if err != nil {
		log.Printf("❌ Failed to list droplets: %v", err)
		return fmt.Errorf("failed to list droplets: %w", err)
	}

	// Find the droplet by name and region
	var dropletID int
	for _, d := range droplets {
		if d.Name == name && d.Region.Slug == region {
			dropletID = d.ID
			break
		}
	}

	if dropletID == 0 {
		return fmt.Errorf("droplet with name %s in region %s not found", name, region)
	}

	// Delete the droplet using the DO API directly
	_, err = p.doClient.Droplets.Delete(ctx, dropletID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %w", err)
	}

	// Wait for the droplet to be fully deleted
	if err := p.waitForDeletion(ctx, name, region, 12); err != nil { // 1 minute timeout (12 * 5 seconds)
		return fmt.Errorf("failed while waiting for droplet deletion: %w", err)
	}

	log.Printf("✅ Droplet deletion confirmed: %s in region %s", name, region)
	return nil
}
