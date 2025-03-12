package compute

import (
	"context"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions

// NewTestProvider creates a DigitalOceanProvider with a mock client for testing
func NewTestProvider() (*DigitalOceanProvider, *MockDOClient) {
	mockClient := NewMockDOClient()
	provider := &DigitalOceanProvider{}
	provider.SetClient(mockClient)
	return provider, mockClient
}

// SetupMockDropletCreate configures the mock client to return a successful droplet creation
func SetupMockDropletCreate(mockClient *MockDOClient, dropletID int, dropletName string) {
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
func SetupMockSSHKeyList(mockClient *MockDOClient, keys []godo.Key) {
	mockClient.MockKeyService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return keys, nil, nil
	}
}

// Tests

func TestNewDigitalOceanProvider(t *testing.T) {
	// Save original env var
	originalToken := os.Getenv("DIGITALOCEAN_TOKEN")
	defer func() {
		err := os.Setenv("DIGITALOCEAN_TOKEN", originalToken)
		if err != nil {
			t.Logf("Failed to restore DIGITALOCEAN_TOKEN: %v", err)
		}
	}()

	t.Run("valid_token", func(t *testing.T) {
		// Set env var for test
		err := os.Setenv("DIGITALOCEAN_TOKEN", "test-token")
		require.NoError(t, err)

		provider, err := NewDigitalOceanProvider()
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("empty_token", func(t *testing.T) {
		// Clear env var for test
		err := os.Setenv("DIGITALOCEAN_TOKEN", "")
		require.NoError(t, err)

		provider, err := NewDigitalOceanProvider()
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "DIGITALOCEAN_TOKEN environment variable is not set")
	})
}

func TestDigitalOceanProvider_ValidateCredentials(t *testing.T) {
	t.Run("valid_token", func(t *testing.T) {
		provider, _ := NewTestProvider()
		err := provider.ValidateCredentials()
		assert.NoError(t, err)
	})

	t.Run("empty_token", func(t *testing.T) {
		// Save original env var
		originalToken := os.Getenv("DIGITALOCEAN_TOKEN")
		defer func() {
			err := os.Setenv("DIGITALOCEAN_TOKEN", originalToken)
			if err != nil {
				t.Logf("Failed to restore DIGITALOCEAN_TOKEN: %v", err)
			}
		}()

		// Clear env var for test
		err := os.Setenv("DIGITALOCEAN_TOKEN", "")
		require.NoError(t, err)

		provider, err := NewDigitalOceanProvider()
		if err != nil {
			t.Skipf("Skipping test because provider creation failed: %v", err)
			return
		}

		err = provider.ValidateCredentials()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client not initialized")
	})
}

func TestDigitalOceanProvider_GetEnvironmentVars(t *testing.T) {
	t.Run("valid_token", func(t *testing.T) {
		// Save original env var
		originalToken := os.Getenv("DIGITALOCEAN_TOKEN")
		defer func() {
			err := os.Setenv("DIGITALOCEAN_TOKEN", originalToken)
			if err != nil {
				t.Logf("Failed to restore DIGITALOCEAN_TOKEN: %v", err)
			}
		}()

		// Set env var for test
		err := os.Setenv("DIGITALOCEAN_TOKEN", "test-token")
		require.NoError(t, err)

		provider, _ := NewTestProvider()
		envVars := provider.GetEnvironmentVars()
		assert.Equal(t, "test-token", envVars["DIGITALOCEAN_TOKEN"])
	})

	t.Run("empty_token", func(t *testing.T) {
		// Save original env var
		originalToken := os.Getenv("DIGITALOCEAN_TOKEN")
		defer func() {
			err := os.Setenv("DIGITALOCEAN_TOKEN", originalToken)
			if err != nil {
				t.Logf("Failed to restore DIGITALOCEAN_TOKEN: %v", err)
			}
		}()

		// Clear env var for test
		err := os.Setenv("DIGITALOCEAN_TOKEN", "")
		require.NoError(t, err)

		provider, err := NewDigitalOceanProvider()
		if err != nil {
			t.Skipf("Skipping test because provider creation failed: %v", err)
			return
		}

		envVars := provider.GetEnvironmentVars()
		assert.Equal(t, "", envVars["DIGITALOCEAN_TOKEN"])
	})
}

func TestDigitalOceanProvider_CreateInstance(t *testing.T) {
	// Skip if no token is set
	if os.Getenv("DIGITALOCEAN_TOKEN") == "" {
		t.Skip("Skipping test because DIGITALOCEAN_TOKEN is not set")
	}

	// This test would create actual resources, so we'll just test the request creation
	provider, err := NewDigitalOceanProvider()
	require.NoError(t, err)

	config := InstanceConfig{
		Region:   "nyc1",
		Size:     "s-1vcpu-1gb",
		Image:    "ubuntu-20-04-x64",
		SSHKeyID: "nonexistent-key", // This should cause the test to fail early
	}

	_, err = provider.CreateInstance(context.Background(), "test-instance", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get SSH key ID")
}

func TestDigitalOceanProvider_DeleteInstance(t *testing.T) {
	// Skip if no token is set
	if os.Getenv("DIGITALOCEAN_TOKEN") == "" {
		t.Skip("Skipping test because DIGITALOCEAN_TOKEN is not set")
	}

	// This test would delete actual resources, so we'll just test with a nonexistent instance
	provider, err := NewDigitalOceanProvider()
	require.NoError(t, err)

	err = provider.DeleteInstance(context.Background(), "nonexistent-instance", "nyc1")
	assert.NoError(t, err) // Should not error for nonexistent instances
}

func TestDigitalOceanProvider_WaitForIP(t *testing.T) {
	provider, mockClient := NewTestProvider()

	// Setup mock
	mockClient.MockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		return &godo.Droplet{
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{
					{
						Type:      "public",
						IPAddress: "192.0.2.1",
					},
				},
			},
		}, nil, nil
	}

	ip, err := provider.WaitForIP(context.Background(), 12345, 1)
	assert.NoError(t, err)
	assert.Equal(t, "192.0.2.1", ip)
}

func TestDigitalOceanProvider_GetSSHKeyID(t *testing.T) {
	provider, mockClient := NewTestProvider()

	// Setup mock
	SetupMockSSHKeyList(mockClient, []godo.Key{
		{ID: 12345, Name: "test-key"},
	})

	id, err := provider.GetSSHKeyID(context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, 12345, id)
}

func TestDigitalOceanProvider_WaitForDeletion(t *testing.T) {
	provider, mockClient := NewTestProvider()

	// Setup mock to return no droplets (already deleted)
	mockClient.MockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		return []godo.Droplet{}, nil, nil
	}

	err := provider.WaitForDeletion(context.Background(), "test-instance", "nyc1", 1)
	assert.NoError(t, err)
}

func TestDigitalOceanProvider_CreateDropletRequest(t *testing.T) {
	provider, _ := NewTestProvider()

	config := InstanceConfig{
		Region:   "nyc1",
		Size:     "s-1vcpu-1gb",
		Image:    "ubuntu-20-04-x64",
		SSHKeyID: "test-key",
		Tags:     []string{"test", "dev"},
	}

	request := provider.CreateDropletRequest("test-instance", config, 12345)
	assert.Equal(t, "test-instance", request.Name)
	assert.Equal(t, "nyc1", request.Region)
	assert.Equal(t, "s-1vcpu-1gb", request.Size)
	assert.Equal(t, "ubuntu-20-04-x64", request.Image.Slug)
	assert.Equal(t, 12345, request.SSHKeys[0].ID)
	assert.Contains(t, request.Tags, "test-instance")
	assert.Contains(t, request.Tags, "test")
	assert.Contains(t, request.Tags, "dev")
}

func TestDigitalOceanProvider_ConfigureProvider(t *testing.T) {
	provider, _ := NewTestProvider()
	err := provider.ConfigureProvider(nil)
	assert.NoError(t, err)
}

func TestDigitalOceanProvider_CreateMultipleDroplets(t *testing.T) {
	t.Skip("This test requires access to unexported methods and should be rewritten using the mock client")
}

func TestCreateInstanceWithMock(t *testing.T) {
	provider, mockClient := NewTestProvider()

	// Setup mocks
	SetupMockSSHKeyList(mockClient, []godo.Key{
		{ID: 12345, Name: "test-key"},
	})

	SetupMockDropletCreate(mockClient, 54321, "test-instance")

	// Create instance
	config := InstanceConfig{
		Region:   "nyc1",
		Size:     "s-1vcpu-1gb",
		Image:    "ubuntu-20-04-x64",
		SSHKeyID: "test-key",
	}

	instances, err := provider.CreateInstance(context.Background(), "test-instance", config)
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "test-instance", instances[0].Name)
	assert.Equal(t, "192.0.2.1", instances[0].PublicIP)
	assert.Equal(t, "54321", instances[0].ID)
}

func TestDeleteInstanceWithMock(t *testing.T) {
	provider, mockClient := NewTestProvider()

	// Setup mocks
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

	mockClient.MockDropletService.DeleteFunc = func(ctx context.Context, id int) (*godo.Response, error) {
		return nil, nil
	}

	// First list call returns the droplet, second call (in waitForDeletion) returns empty list
	var listCallCount int
	mockClient.MockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		listCallCount++
		if listCallCount == 1 {
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
		return []godo.Droplet{}, nil, nil
	}

	// Delete instance
	err := provider.DeleteInstance(context.Background(), "test-instance", "nyc1")
	assert.NoError(t, err)
}
