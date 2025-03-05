package compute

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDigitalOceanProvider(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "valid-token",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.token != "" {
				os.Setenv("DIGITALOCEAN_TOKEN", tt.token)
			} else {
				os.Unsetenv("DIGITALOCEAN_TOKEN")
			}

			// Create provider
			provider, err := NewDigitalOceanProvider()

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
				return
			}

			// Check provider
			assert.NoError(t, err)
			assert.NotNil(t, provider)
			assert.NotNil(t, provider.doClient)
		})
	}
}

func TestDigitalOceanProvider_ValidateCredentials(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   "valid-token",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.token != "" {
				os.Setenv("DIGITALOCEAN_TOKEN", tt.token)
			} else {
				os.Unsetenv("DIGITALOCEAN_TOKEN")
			}

			// Create provider
			provider, err := NewDigitalOceanProvider()
			if err != nil {
				t.Skipf("Skipping test because provider creation failed: %v", err)
			}

			// Validate credentials
			err = provider.ValidateCredentials()

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestDigitalOceanProvider_GetEnvironmentVars(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		wantVars map[string]string
	}{
		{
			name:  "valid token",
			token: "valid-token",
			wantVars: map[string]string{
				"DIGITALOCEAN_TOKEN": "valid-token",
			},
		},
		{
			name:  "empty token",
			token: "",
			wantVars: map[string]string{
				"DIGITALOCEAN_TOKEN": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.token != "" {
				os.Setenv("DIGITALOCEAN_TOKEN", tt.token)
			} else {
				os.Unsetenv("DIGITALOCEAN_TOKEN")
			}

			// Create provider
			provider, err := NewDigitalOceanProvider()
			if err != nil {
				t.Skipf("Skipping test because provider creation failed: %v", err)
			}

			// Get environment variables
			vars := provider.GetEnvironmentVars()

			// Check variables
			assert.Equal(t, tt.wantVars, vars)
		})
	}
}

func TestDigitalOceanProvider_CreateInstance(t *testing.T) {
	// Skip if no token is provided
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		t.Skip("Skipping test because DIGITALOCEAN_TOKEN is not set")
	}

	// Create provider
	provider, err := NewDigitalOceanProvider()
	require.NoError(t, err)
	require.NotNil(t, provider)

	tests := []struct {
		name           string
		config         InstanceConfig
		wantErr        bool
		validateResult func(*testing.T, InstanceInfo, error)
	}{
		{
			name: "single instance with invalid key",
			config: InstanceConfig{
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 1,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info.PublicIP)
			},
		},
		{
			name: "multiple instances with invalid key",
			config: InstanceConfig{
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 2,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info.PublicIP)
			},
		},
		{
			name: "invalid region",
			config: InstanceConfig{
				Region:            "invalid-region",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 1,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info.PublicIP)
			},
		},
		{
			name: "invalid size",
			config: InstanceConfig{
				Region:            "nyc3",
				Size:              "invalid-size",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 1,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info.PublicIP)
			},
		},
		{
			name: "invalid image",
			config: InstanceConfig{
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             "invalid-image",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 1,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info.PublicIP)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			info, err := provider.CreateInstance(ctx, "test-instance", tt.config)
			tt.validateResult(t, info, err)
		})
	}
}

func TestDigitalOceanProvider_DeleteInstance(t *testing.T) {
	// Skip if no token is provided
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		t.Skip("Skipping test because DIGITALOCEAN_TOKEN is not set")
	}

	// Create provider
	provider, err := NewDigitalOceanProvider()
	require.NoError(t, err)
	require.NotNil(t, provider)

	tests := []struct {
		name         string
		instanceName string
		wantErr      bool
	}{
		{
			name:         "non-existent instance",
			instanceName: "test-instance",
			wantErr:      true,
		},
		{
			name:         "empty instance name",
			instanceName: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := provider.DeleteInstance(ctx, tt.instanceName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestDigitalOceanProvider_ConfigureProvider(t *testing.T) {
	provider := &DigitalOceanProvider{}
	err := provider.ConfigureProvider(nil)
	assert.NoError(t, err)
}
