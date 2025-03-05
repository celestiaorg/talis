package compute

import (
	"context"
	"fmt"
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
	fmt.Printf("🔑 Looking up SSH key: %s\n", keyName)

	// List all SSH keys
	keys, _, err := p.doClient.Keys.List(ctx, &godo.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list SSH keys: %w", err)
	}

	// Find the key by name
	for _, key := range keys {
		if key.Name == keyName {
			fmt.Printf("✅ Found SSH key '%s' with ID: %d\n", keyName, key.ID)
			return key.ID, nil
		}
	}

	// If we get here, print available keys to help with diagnosis
	if len(keys) > 0 {
		fmt.Println("Available SSH keys:")
		for _, key := range keys {
			fmt.Printf("  - %s (ID: %d)\n", key.Name, key.ID)
		}
	}

	return 0, fmt.Errorf("SSH key '%s' not found", keyName)
}

// CreateInstance creates a new DigitalOcean droplet
func (p *DigitalOceanProvider) CreateInstance(ctx context.Context, name string, config InstanceConfig) (InstanceInfo, error) {
	fmt.Printf("🚀 Creating DigitalOcean droplet(s): %s\n", name)
	fmt.Printf("  Region: %s\n", config.Region)
	fmt.Printf("  Size: %s\n", config.Size)
	fmt.Printf("  Image: %s\n", config.Image)
	fmt.Printf("  Number of instances: %d\n", config.NumberOfInstances)

	// Get SSH key ID
	sshKeyID, err := p.getSSHKeyID(ctx, config.SSHKeyID)
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("failed to get SSH key: %w", err)
	}

	// If we're creating multiple instances, use the CreateMultiple API
	if config.NumberOfInstances > 1 {
		names := make([]string, config.NumberOfInstances)
		for i := 0; i < config.NumberOfInstances; i++ {
			names[i] = fmt.Sprintf("%s-%d", name, i+1)
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

		droplets, _, err := p.doClient.Droplets.CreateMultiple(ctx, createRequest)
		if err != nil {
			fmt.Printf("❌ Failed to create droplets: %v\n", err)
			return InstanceInfo{}, fmt.Errorf("failed to create droplets: %w", err)
		}

		// Wait for the first droplet to get an IP (we'll return this one as the primary)
		fmt.Println("⏳ Waiting for droplets to get IP addresses...")
		maxRetries := 10
		var ip string
		for i := 0; i < maxRetries; i++ {
			d, _, err := p.doClient.Droplets.Get(ctx, droplets[0].ID)
			if err != nil {
				return InstanceInfo{}, fmt.Errorf("failed to get droplet details: %w", err)
			}

			// Get the public IPv4 address
			for _, network := range d.Networks.V4 {
				if network.Type == "public" {
					ip = network.IPAddress
					fmt.Printf("📍 Found public IP for primary droplet: %s\n", ip)
					break
				}
			}

			if ip != "" {
				break
			}

			fmt.Printf("⏳ IP not assigned yet, retrying in 10 seconds (attempt %d/%d)...\n", i+1, maxRetries)
			time.Sleep(10 * time.Second)
		}

		if ip == "" {
			return InstanceInfo{}, fmt.Errorf("droplets created but no public IP found after %d retries", maxRetries)
		}

		fmt.Printf("✅ Created %d droplets with base name: %s\n", len(droplets), name)
		return InstanceInfo{
			ID:       fmt.Sprintf("%d", droplets[0].ID),
			Name:     name,
			PublicIP: ip,
			Provider: "digitalocean",
			Region:   config.Region,
			Size:     config.Size,
		}, nil
	}

	// Single instance creation (existing code)
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
	droplet, _, err := p.doClient.Droplets.Create(ctx, createRequest)
	if err != nil {
		fmt.Printf("❌ Failed to create droplet: %v\n", err)
		return InstanceInfo{}, fmt.Errorf("failed to create droplet: %w", err)
	}

	// Wait for droplet to get an IP
	fmt.Println("⏳ Waiting for droplet to get an IP address...")
	maxRetries := 10
	var ip string
	for i := 0; i < maxRetries; i++ {
		d, _, err := p.doClient.Droplets.Get(ctx, droplet.ID)
		if err != nil {
			return InstanceInfo{}, fmt.Errorf("failed to get droplet details: %w", err)
		}

		// Get the public IPv4 address
		for _, network := range d.Networks.V4 {
			if network.Type == "public" {
				ip = network.IPAddress
				fmt.Printf("📍 Found public IP for droplet: %s\n", ip)
				break
			}
		}

		if ip != "" {
			break
		}

		fmt.Printf("⏳ IP not assigned yet, retrying in 10 seconds (attempt %d/%d)...\n", i+1, maxRetries)
		time.Sleep(10 * time.Second)
	}

	if ip == "" {
		return InstanceInfo{}, fmt.Errorf("droplet created but no public IP found after %d retries", maxRetries)
	}

	fmt.Printf("✅ Droplet creation completed: %s (IP: %s)\n", name, ip)
	return InstanceInfo{
		ID:       fmt.Sprintf("%d", droplet.ID),
		Name:     name,
		PublicIP: ip,
		Provider: "digitalocean",
		Region:   config.Region,
		Size:     config.Size,
	}, nil
}

// DeleteInstance deletes a DigitalOcean droplet
func (p *DigitalOceanProvider) DeleteInstance(ctx context.Context, name string) error {
	fmt.Printf("🗑️ Deleting DigitalOcean droplet: %s\n", name)

	// List all droplets to find the one with our name
	droplets, _, err := p.doClient.Droplets.List(ctx, &godo.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list droplets: %w", err)
	}

	// Find the droplet by name
	var dropletID int
	for _, d := range droplets {
		if d.Name == name {
			dropletID = d.ID
			break
		}
	}

	if dropletID == 0 {
		return fmt.Errorf("droplet with name %s not found", name)
	}

	// Delete the droplet using the DO API directly
	_, err = p.doClient.Droplets.Delete(ctx, dropletID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %w", err)
	}

	fmt.Printf("✅ Droplet deletion initiated: %s\n", name)
	return nil
}

// GetNixOSConfig returns the provider-specific NixOS configuration
func (d *DigitalOceanProvider) GetNixOSConfig() string {
	return "cloud/digitalocean.nix"
}
