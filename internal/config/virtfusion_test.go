package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVirtFusionConfig(t *testing.T) {
	// Save original env vars
	originalToken := os.Getenv("VIRTFUSION_API_TOKEN")
	originalHost := os.Getenv("VIRTFUSION_HOST")
	defer func() {
		os.Setenv("VIRTFUSION_API_TOKEN", originalToken)
		os.Setenv("VIRTFUSION_HOST", originalHost)
	}()

	tests := []struct {
		name      string
		setEnv    map[string]string
		wantError bool
	}{
		{
			name: "valid configuration",
			setEnv: map[string]string{
				"VIRTFUSION_API_TOKEN": "test-token",
				"VIRTFUSION_HOST":      "https://api.virtfusion.com",
			},
			wantError: false,
		},
		{
			name: "missing token",
			setEnv: map[string]string{
				"VIRTFUSION_API_TOKEN": "",
				"VIRTFUSION_HOST":      "https://api.virtfusion.com",
			},
			wantError: true,
		},
		{
			name: "missing host",
			setEnv: map[string]string{
				"VIRTFUSION_API_TOKEN": "test-token",
				"VIRTFUSION_HOST":      "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.setEnv {
				err := os.Setenv(k, v)
				require.NoError(t, err)
			}

			// Test configuration creation
			config, err := NewVirtFusionConfig()
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.setEnv["VIRTFUSION_API_TOKEN"], config.APIToken)
				assert.Equal(t, tt.setEnv["VIRTFUSION_HOST"], config.Host)
			}
		})
	}
}

func TestVirtFusionConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *VirtFusionConfig
		wantError bool
	}{
		{
			name: "valid configuration",
			config: &VirtFusionConfig{
				APIToken: "test-token",
				Host:     "https://api.virtfusion.com",
			},
			wantError: false,
		},
		{
			name: "missing token",
			config: &VirtFusionConfig{
				APIToken: "",
				Host:     "https://api.virtfusion.com",
			},
			wantError: true,
		},
		{
			name: "missing host",
			config: &VirtFusionConfig{
				APIToken: "test-token",
				Host:     "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVirtFusionConfig_GetEnvironmentVars(t *testing.T) {
	config := &VirtFusionConfig{
		APIToken: "test-token",
		Host:     "https://api.virtfusion.com",
	}

	envVars := config.GetEnvironmentVars()
	assert.Equal(t, config.APIToken, envVars["VIRTFUSION_API_TOKEN"])
	assert.Equal(t, config.Host, envVars["VIRTFUSION_HOST"])
}
