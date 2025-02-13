package compute

import (
	"context"
	"fmt"
	"os"

	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DigitalOceanProvider struct{}

// ValidateCredentials verifies that the required DigitalOcean credentials are present
func (d *DigitalOceanProvider) ValidateCredentials() error {
	doToken := os.Getenv("DIGITALOCEAN_TOKEN")
	if doToken == "" {
		return fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}
	return nil
}

// GetEnvironmentVars returns the environment variables needed for DigitalOcean
func (d *DigitalOceanProvider) GetEnvironmentVars() map[string]string {
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
	err := stack.SetConfig(context.Background(), "digitalocean:token", auto.ConfigValue{
		Value:  doToken,
		Secret: true,
	})
	if err != nil {
		fmt.Printf("❌ Failed to configure DigitalOcean: %v\n", err)
		return err
	}

	fmt.Println("✅ DigitalOcean credentials configured successfully")
	return nil
}

// getSSHKeyID obtiene el ID de una clave SSH por su nombre
func (p *DigitalOceanProvider) getSSHKeyID(ctx *pulumi.Context, keyName string) (string, error) {
	fmt.Printf("🔑 Looking up SSH key: %s\n", keyName)

	// Verificar que tenemos el token configurado
	if os.Getenv("DIGITALOCEAN_TOKEN") == "" {
		return "", fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}

	sshKey, err := digitalocean.LookupSshKey(ctx, &digitalocean.LookupSshKeyArgs{
		Name: keyName,
	})
	if err != nil {
		// Intentar listar las claves disponibles para ayudar en el diagnóstico
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
func (p *DigitalOceanProvider) CreateInstance(
	ctx *pulumi.Context,
	name string,
	config InstanceConfig,
) (pulumi.Resource, error) {
	fmt.Printf("🚀 Creating DigitalOcean droplet: %s\n", name)
	fmt.Printf("  Region: %s\n", config.Region)
	fmt.Printf("  Size: %s\n", config.Size)
	fmt.Printf("  Image: %s\n", config.Image)

	// Get SSH key ID first
	sshKeyID, err := p.getSSHKeyID(ctx, config.SSHKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key: %v", err)
	}

	droplet, err := digitalocean.NewDroplet(ctx, name, &digitalocean.DropletArgs{
		Image:  pulumi.String(config.Image),
		Region: pulumi.String(config.Region),
		Size:   pulumi.String(config.Size),
		SshKeys: pulumi.StringArray{
			pulumi.String(sshKeyID),
		},
		Tags: pulumi.ToStringArray(config.Tags),
	})
	if err != nil {
		fmt.Printf("❌ Failed to create droplet: %v\n", err)
		return nil, fmt.Errorf("failed to create droplet: %v", err)
	}

	// Export the IP address and ensure it's ready
	ip := droplet.Ipv4Address.ApplyT(func(ip string) string {
		fmt.Printf("✅ Instance %s is ready with IP: %s\n", name, ip)
		return ip
	}).(pulumi.StringOutput)

	ctx.Export(fmt.Sprintf("%s_ip", name), ip)

	fmt.Printf("✅ Droplet creation initiated: %s\n", name)
	return droplet, nil
}

// GetNixOSConfig returns the provider-specific NixOS configuration
func (d *DigitalOceanProvider) GetNixOSConfig() string {
	return "cloud/digitalocean.nix"
}
