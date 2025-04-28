package config

import (
	"fmt"
	"os"
)

// VirtFusionConfig represents the configuration for VirtFusion
type VirtFusionConfig struct {
	APIToken string
	Host     string
}

// NewVirtFusionConfig creates a new VirtFusion configuration
func NewVirtFusionConfig() (*VirtFusionConfig, error) {
	apiToken := os.Getenv("VIRTFUSION_API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("VIRTFUSION_API_TOKEN environment variable is not set")
	}

	host := os.Getenv("VIRTFUSION_HOST")
	if host == "" {
		return nil, fmt.Errorf("VIRTFUSION_HOST environment variable is not set")
	}

	return &VirtFusionConfig{
		APIToken: apiToken,
		Host:     host,
	}, nil
}

// Validate validates the VirtFusion configuration
func (c *VirtFusionConfig) Validate() error {
	if c.APIToken == "" {
		return fmt.Errorf("API token is required")
	}
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	return nil
}

// GetEnvironmentVars returns the environment variables needed for VirtFusion
func (c *VirtFusionConfig) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"VIRTFUSION_API_TOKEN": c.APIToken,
		"VIRTFUSION_HOST":      c.Host,
	}
}
