package compute

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DigitalOceanProvider implements the ComputeProvider interface
type DigitalOceanProvider struct{}

// ValidateCredentials validates the DigitalOcean credentials
func (p *DigitalOceanProvider) ValidateCredentials() error {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		return fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}
	return nil
}

// GetEnvironmentVars returns the environment variables needed for DigitalOcean
func (p *DigitalOceanProvider) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"DIGITALOCEAN_TOKEN": os.Getenv("DIGITALOCEAN_TOKEN"),
	}
}

// ConfigureProvider configures the DigitalOcean provider with necessary credentials
func (p *DigitalOceanProvider) ConfigureProvider(stack auto.Stack) error {
	doToken := os.Getenv("DIGITALOCEAN_TOKEN")
	if doToken == "" {
		return fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}

	fmt.Println("🔑 Setting up DigitalOcean credentials...")
	return stack.SetConfig(context.Background(), "digitalocean:token", auto.ConfigValue{
		Value:  doToken,
		Secret: true,
	})
}

// getSSHKeyID gets the ID of an SSH key by its name
func (p *DigitalOceanProvider) getSSHKeyID(ctx *pulumi.Context, keyName string) (string, error) {
	fmt.Printf("🔑 Looking up SSH key: %s\n", keyName)

	// Verify that we have the token configured
	if os.Getenv("DIGITALOCEAN_TOKEN") == "" {
		return "", fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}

	sshKey, err := digitalocean.LookupSshKey(ctx, &digitalocean.LookupSshKeyArgs{
		Name: keyName,
	})
	if err != nil {
		// Try listing available keys to help with diagnosis
		keys, listErr := digitalocean.GetSshKeys(ctx, nil)
		if listErr == nil && len(keys.SshKeys) > 0 {
			fmt.Println("Available SSH keys:")
			for _, key := range keys.SshKeys {
				fmt.Printf("  - %s (ID: %s)\n", key.Name, key.Fingerprint)
			}
		}
		return "", fmt.Errorf("failed to get SSH key '%s': %v", keyName, err)
	}

	fmt.Printf("✅ Found SSH key '%s' with ID: %s\n", keyName, sshKey.Fingerprint)
	return sshKey.Fingerprint, nil
}

// CreateInstance creates a new DigitalOcean droplet
func (p *DigitalOceanProvider) CreateInstance(ctx *pulumi.Context, name string, config InstanceConfig) (InstanceInfo, error) {
	fmt.Printf("🚀 Creating DigitalOcean droplet: %s\n", name)
	fmt.Printf("  Region: %s\n", config.Region)
	fmt.Printf("  Size: %s\n", config.Size)
	fmt.Printf("  Image: %s\n", config.Image)

	// Si no se proporciona user_data, usar un script básico
	userData := config.UserData
	if userData == "" {
		userData = `#!/bin/bash
apt-get update
apt-get install -y python3`
	}

	sshKeyID, err := p.getSSHKeyID(ctx, config.SSHKeyID)
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("failed to get SSH key: %v", err)
	}

	// Convert slice of tags to pulumi.StringArray
	tags := make(pulumi.StringArray, len(config.Tags)+1)
	tags[0] = pulumi.String(name)
	for i, tag := range config.Tags {
		tags[i+1] = pulumi.String(tag)
	}

	droplet, err := digitalocean.NewDroplet(ctx, name, &digitalocean.DropletArgs{
		Region:   pulumi.String(config.Region),
		Size:     pulumi.String(config.Size),
		Image:    pulumi.String(config.Image),
		UserData: pulumi.String(userData),
		SshKeys:  pulumi.StringArray{pulumi.String(sshKeyID)},
		Tags:     tags,
	})
	if err != nil {
		fmt.Printf("❌ Failed to create droplet: %v\n", err)
		return InstanceInfo{}, fmt.Errorf("failed to create droplet: %v", err)
	}

	// Export the IP address and ensure it's ready
	ip := droplet.Ipv4Address.ApplyT(func(ip string) string {
		fmt.Printf("✅ Instance %s is ready with IP: %s\n", name, ip)
		return ip
	}).(pulumi.StringOutput)

	ctx.Export(fmt.Sprintf("%s_ip", name), ip)

	fmt.Printf("✅ Droplet creation initiated: %s\n", name)
	return InstanceInfo{
		PublicIP: droplet.Ipv4Address,
	}, nil
}

// GetNixOSConfig returns the provider-specific NixOS configuration
func (d *DigitalOceanProvider) GetNixOSConfig() string {
	return "cloud/digitalocean.nix"
}
