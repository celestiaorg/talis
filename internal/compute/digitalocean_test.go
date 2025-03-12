package compute_test

import (
	"context"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/compute/mocks"
)

// Test helper functions

// NewTestProvider creates a DigitalOceanProvider with a mock client for testing
func NewTestProvider() (*compute.DigitalOceanProvider, *mocks.MockDOClient) {
	mockClient := mocks.NewMockDOClient()
	provider := &compute.DigitalOceanProvider{
		DOClient: mockClient,
	}
	return provider, mockClient
}

// SetupMockDropletCreate configures the mock client to return a successful droplet creation
func SetupMockDropletCreate(mockClient *mocks.MockDOClient, dropletID int, dropletName string) {
	mockClient.MockDropletService.CreateFunc = func(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		return &godo.Droplet{
			ID:   dropletID,
			Name: dropletName,
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{
					{
						Type:      "public",
						IPAddress: "192.0.2.1",
					},
				},
			},
			Region: &godo.Region{
				Slug: req.Region,
			},
		}, nil, nil
	}
}

// SetupMockSSHKeyList configures the mock client to return a list of SSH keys
func SetupMockSSHKeyList(mockClient *mocks.MockDOClient, keys []godo.Key) {
	mockClient.MockKeyService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return keys, nil, nil
	}
}

// Tests

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
				err := os.Setenv("DIGITALOCEAN_TOKEN", tt.token)
				require.NoError(t, err)
			} else {
				err := os.Unsetenv("DIGITALOCEAN_TOKEN")
				require.NoError(t, err)
			}

			// Create provider
			provider, err := compute.NewDigitalOceanProvider()

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
				return
			}

			// Check provider
			assert.NoError(t, err)
			assert.NotNil(t, provider)
			assert.NotNil(t, provider.DOClient)
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
				err := os.Setenv("DIGITALOCEAN_TOKEN", tt.token)
				require.NoError(t, err)
			} else {
				err := os.Unsetenv("DIGITALOCEAN_TOKEN")
				require.NoError(t, err)
			}

			// Create provider
			provider, err := compute.NewDigitalOceanProvider()
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
				err := os.Setenv("DIGITALOCEAN_TOKEN", tt.token)
				require.NoError(t, err)
			} else {
				err := os.Unsetenv("DIGITALOCEAN_TOKEN")
				require.NoError(t, err)
			}

			// Create provider
			provider, err := compute.NewDigitalOceanProvider()
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
	provider, err := compute.NewDigitalOceanProvider()
	require.NoError(t, err)
	require.NotNil(t, provider)

	tests := []struct {
		name           string
		config         compute.InstanceConfig
		wantErr        bool
		validateResult func(*testing.T, []compute.InstanceInfo, error)
	}{
		{
			name: "single instance with invalid key",
			config: compute.InstanceConfig{
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 1,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info []compute.InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info)
			},
		},
		{
			name: "multiple instances with invalid key",
			config: compute.InstanceConfig{
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 2,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info []compute.InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info)
			},
		},
		{
			name: "invalid region",
			config: compute.InstanceConfig{
				Region:            "invalid-region",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 1,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info []compute.InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info)
			},
		},
		{
			name: "invalid size",
			config: compute.InstanceConfig{
				Region:            "nyc3",
				Size:              "invalid-size",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 1,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info []compute.InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info)
			},
		},
		{
			name: "invalid image",
			config: compute.InstanceConfig{
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             "invalid-image",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 1,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info []compute.InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info)
			},
		},
		{
			name: "zero instances",
			config: compute.InstanceConfig{
				Region:            "nyc3",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-22-04-x64",
				SSHKeyID:          "test-key",
				Tags:              []string{"test"},
				NumberOfInstances: 0,
			},
			wantErr: true,
			validateResult: func(t *testing.T, info []compute.InstanceInfo, err error) {
				assert.Error(t, err)
				assert.Empty(t, info)
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
	provider, err := compute.NewDigitalOceanProvider()
	require.NoError(t, err)
	require.NotNil(t, provider)

	tests := []struct {
		name         string
		instanceName string
		region       string
		wantErr      bool
	}{
		{
			name:         "non-existent instance",
			instanceName: "test-instance",
			region:       "nyc3",
			wantErr:      true,
		},
		{
			name:         "empty instance name",
			instanceName: "",
			region:       "nyc3",
			wantErr:      true,
		},
		{
			name:         "empty region",
			instanceName: "test-instance",
			region:       "",
			wantErr:      true,
		},
		{
			name:         "invalid region",
			instanceName: "test-instance",
			region:       "invalid-region",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := provider.DeleteInstance(ctx, tt.instanceName, tt.region)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDigitalOceanProvider_WaitForIP tests the waitForIP functionality
func TestDigitalOceanProvider_WaitForIP(t *testing.T) {
	// Test with nil client
	provider := &compute.DigitalOceanProvider{
		DOClient: nil,
	}

	ctx := context.Background()
	_, err := provider.WaitForIP(ctx, 123, 1)
	assert.Error(t, err)
}

// TestDigitalOceanProvider_GetSSHKeyID tests the getSSHKeyID functionality
func TestDigitalOceanProvider_GetSSHKeyID(t *testing.T) {
	// Test with nil client
	provider := &compute.DigitalOceanProvider{
		DOClient: nil,
	}

	ctx := context.Background()
	_, err := provider.GetSSHKeyID(ctx, "test-key")
	assert.Error(t, err)
}

// TestDigitalOceanProvider_WaitForDeletion tests the waitForDeletion functionality
func TestDigitalOceanProvider_WaitForDeletion(t *testing.T) {
	// Test with nil client
	provider := &compute.DigitalOceanProvider{
		DOClient: nil,
	}

	ctx := context.Background()
	err := provider.WaitForDeletion(ctx, "test-droplet", "nyc3", 1)
	assert.Error(t, err)
}

// TestDigitalOceanProvider_CreateDropletRequest tests the createDropletRequest functionality
func TestDigitalOceanProvider_CreateDropletRequest(t *testing.T) {
	provider := &compute.DigitalOceanProvider{}

	config := compute.InstanceConfig{
		Region:   "nyc3",
		Size:     "s-1vcpu-1gb",
		Image:    "ubuntu-20-04-x64",
		SSHKeyID: "test-key",
		Tags:     []string{"test", "example"},
	}

	request := provider.CreateDropletRequest("test-droplet", config, 12345)

	assert.Equal(t, "test-droplet", request.Name)
	assert.Equal(t, "nyc3", request.Region)
	assert.Equal(t, "s-1vcpu-1gb", request.Size)
	assert.Equal(t, "ubuntu-20-04-x64", request.Image.Slug)
	assert.Equal(t, 12345, request.SSHKeys[0].ID)
	assert.Contains(t, request.Tags, "test-droplet")
	assert.Contains(t, request.Tags, "test")
	assert.Contains(t, request.Tags, "example")
}

func TestDigitalOceanProvider_ConfigureProvider(t *testing.T) {
	provider := &compute.DigitalOceanProvider{}
	err := provider.ConfigureProvider(nil)
	assert.NoError(t, err)
}

// TestDigitalOceanProvider_CreateMultipleDroplets tests the createMultipleDroplets functionality
func TestDigitalOceanProvider_CreateMultipleDroplets(t *testing.T) {
	t.Skip("This test requires access to unexported methods and should be rewritten using the mock client")
}

func TestCreateInstanceWithMock(t *testing.T) {
	// Create a provider with a mock client
	provider, mockClient := NewTestProvider()

	// Configure the mock to return a successful SSH key lookup
	mockClient.MockKeyService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return []godo.Key{
			{ID: 12345, Name: "test-key"},
		}, nil, nil
	}

	// Configure the mock to return a successful droplet creation
	mockClient.MockDropletService.CreateFunc = func(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		return &godo.Droplet{
			ID:   54321,
			Name: req.Name,
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{
					{
						Type:      "public",
						IPAddress: "192.0.2.1",
					},
				},
			},
			Region: &godo.Region{
				Slug: req.Region,
			},
		}, nil, nil
	}

	// Configure the mock to return the droplet with IP when Get is called
	mockClient.MockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		return &godo.Droplet{
			ID:   id,
			Name: "test-instance",
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{
					{
						Type:      "public",
						IPAddress: "192.0.2.1",
					},
				},
			},
			Region: &godo.Region{
				Slug: "nyc1",
			},
		}, nil, nil
	}

	// Test the provider
	instances, err := provider.CreateInstance(context.Background(), "test-instance", compute.InstanceConfig{
		Region:   "nyc1",
		Size:     "s-1vcpu-1gb",
		Image:    "ubuntu-20-04-x64",
		SSHKeyID: "test-key",
	})

	// Assert results
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "test-instance", instances[0].Name)
	assert.Equal(t, "192.0.2.1", instances[0].PublicIP)
	assert.Equal(t, "54321", instances[0].ID)
}

func TestDeleteInstanceWithMock(t *testing.T) {
	// Create a provider with a mock client
	provider, mockClient := NewTestProvider()

	// Configure the mock to return a list of droplets
	mockClient.MockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		return []godo.Droplet{
			{
				ID:   54321,
				Name: "test-instance",
				Region: &godo.Region{
					Slug: "nyc1",
				},
			},
		}, nil, nil
	}

	// Configure the mock to return a successful deletion
	mockClient.MockDropletService.DeleteFunc = func(ctx context.Context, id int) (*godo.Response, error) {
		return nil, nil
	}

	// Configure the mock to return an empty list after deletion
	var listCallCount int
	mockClient.MockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		listCallCount++
		if listCallCount == 1 {
			// First call returns the droplet
			return []godo.Droplet{
				{
					ID:   54321,
					Name: "test-instance",
					Region: &godo.Region{
						Slug: "nyc1",
					},
				},
			}, nil, nil
		}
		// Subsequent calls return empty list (droplet deleted)
		return []godo.Droplet{}, nil, nil
	}

	// Test the provider
	err := provider.DeleteInstance(context.Background(), "test-instance", "nyc1")

	// Assert results
	assert.NoError(t, err)
}
