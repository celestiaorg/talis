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

// ConfigureProvider configures the DigitalOcean provider with necessary credentials
func (p *DigitalOceanProvider) ConfigureProvider(stack auto.Stack) error {
	doToken := os.Getenv("DIGITALOCEAN_TOKEN")
	if doToken == "" {
		return fmt.Errorf("DIGITALOCEAN_TOKEN environment variable is not set")
	}

	fmt.Println("Setting up DigitalOcean credentials...")
	return stack.SetConfig(context.Background(), "digitalocean:token", auto.ConfigValue{
		Value:  doToken,
		Secret: true,
	})
}

// getSSHKeyID obtiene el ID de una clave SSH por su nombre
// In DigitalOcean, the SSH key ID is the fingerprint
func (p *DigitalOceanProvider) getSSHKeyID(ctx *pulumi.Context, keyName string) (string, error) {
	sshKey, err := digitalocean.LookupSshKey(ctx, &digitalocean.LookupSshKeyArgs{
		Name: keyName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get SSH key '%s': %v", keyName, err)
	}

	fmt.Printf("Found SSH key '%s' with ID: %d\n", keyName, sshKey.Id)
	return fmt.Sprintf("%d", sshKey.Id), nil
}

// CreateInstance creates a new DigitalOcean droplet
func (p *DigitalOceanProvider) CreateInstance(ctx *pulumi.Context, name string, config InstanceConfig) (InstanceInfo, error) {
	sshKeyID, err := p.getSSHKeyID(ctx, config.SSHKeyID)
	if err != nil {
		return InstanceInfo{}, err
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
		UserData: pulumi.String(config.UserData),
		SshKeys:  pulumi.StringArray{pulumi.String(sshKeyID)},
		Tags:     tags,
	})
	if err != nil {
		return InstanceInfo{}, fmt.Errorf("failed to create droplet: %v", err)
	}

	return InstanceInfo{
		PublicIP: droplet.Ipv4Address,
	}, nil
}
