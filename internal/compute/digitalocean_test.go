package compute

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions

// newTestProvider creates a DigitalOceanProvider with a mock client for testing
func newTestProvider() (*DigitalOceanProvider, *mockDOClient) {
	mockClient := newMockDOClient()
	provider := &DigitalOceanProvider{}
	provider.SetClient(mockClient)
	return provider, mockClient
}

// setupMockDropletCreate configures the mock client to return a successful droplet creation
func setupMockDropletCreate(mockClient *mockDOClient, dropletID int, dropletName string) {
	mockClient.mockDropletService.CreateFunc = func(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
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

	// Also set up the Get function to return the same droplet with IP
	mockClient.mockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		if id == dropletID {
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
					Slug: "nyc1",
				},
			}, nil, nil
		}
		return nil, nil, fmt.Errorf("droplet not found")
	}
}

// setupMockSSHKeyList configures the mock client to return a list of SSH keys
func setupMockSSHKeyList(mockClient *mockDOClient, keys []godo.Key) {
	mockClient.mockKeyService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return keys, nil, nil
	}
}

// Tests grouped by interface/struct implementation

// DOClient interface and implementations tests

// TestDOClient tests the DOClient interface implementation
func TestDOClient(t *testing.T) {
	t.Run("MockDOClient", func(t *testing.T) {
		mockClient := newMockDOClient()

		dropletService := mockClient.Droplets()
		assert.NotNil(t, dropletService)

		keyService := mockClient.Keys()
		assert.NotNil(t, keyService)
	})

	t.Run("NewDOClient", func(t *testing.T) {
		client := NewDOClient("test-token")
		assert.NotNil(t, client)
	})
}

// DropletService interface and implementations tests

// TestDropletService tests the DropletService interface implementation
func TestDropletService(t *testing.T) {
	t.Run("Create_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mocks
		mockClient.mockDropletService.CreateFunc = func(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
			return &godo.Droplet{
				ID:   12345,
				Name: req.Name,
				Region: &godo.Region{
					Slug: req.Region,
				},
			}, nil, nil
		}

		// Call the method
		droplet, _, err := provider.doClient.Droplets().Create(context.Background(), &godo.DropletCreateRequest{
			Name:   "test-droplet",
			Region: "nyc1",
		})

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, droplet)
		assert.Equal(t, 12345, droplet.ID)
		assert.Equal(t, "test-droplet", droplet.Name)
	})

	t.Run("Create_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockDropletService.CreateFunc = func(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
			return nil, nil, errors.New("API error")
		}

		// Call the method
		droplet, _, err := provider.doClient.Droplets().Create(context.Background(), &godo.DropletCreateRequest{
			Name:   "test-droplet",
			Region: "nyc1",
		})

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, droplet)
	})

	t.Run("CreateMultiple_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mocks
		mockClient.mockDropletService.CreateMultipleFunc = func(ctx context.Context, req *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
			droplets := make([]godo.Droplet, len(req.Names))
			for i, name := range req.Names {
				droplets[i] = godo.Droplet{
					ID:   10000 + i,
					Name: name,
					Region: &godo.Region{
						Slug: req.Region,
					},
				}
			}
			return droplets, nil, nil
		}

		// Call the method
		droplets, _, err := provider.doClient.Droplets().CreateMultiple(context.Background(), &godo.DropletMultiCreateRequest{
			Names:  []string{"test-1", "test-2"},
			Region: "nyc1",
		})

		// Verify results
		assert.NoError(t, err)
		assert.Len(t, droplets, 2)
		assert.Equal(t, "test-1", droplets[0].Name)
		assert.Equal(t, "test-2", droplets[1].Name)
	})

	t.Run("CreateMultiple_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockDropletService.CreateMultipleFunc = func(ctx context.Context, req *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
			return nil, nil, errors.New("API error")
		}

		// Call the method
		droplets, _, err := provider.doClient.Droplets().CreateMultiple(context.Background(), &godo.DropletMultiCreateRequest{
			Names:  []string{"test-1", "test-2"},
			Region: "nyc1",
		})

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, droplets)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mocks
		mockClient.mockDropletService.DeleteFunc = func(ctx context.Context, id int) (*godo.Response, error) {
			return nil, nil
		}

		// Call the method
		_, err := provider.doClient.Droplets().Delete(context.Background(), 12345)

		// Verify results
		assert.NoError(t, err)
	})

	t.Run("Delete_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockDropletService.DeleteFunc = func(ctx context.Context, id int) (*godo.Response, error) {
			return nil, errors.New("API error")
		}

		// Call the method
		_, err := provider.doClient.Droplets().Delete(context.Background(), 12345)

		// Verify results
		assert.Error(t, err)
	})

	t.Run("Get_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mocks
		mockClient.mockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
			return &godo.Droplet{
				ID:   id,
				Name: "test-droplet",
				Region: &godo.Region{
					Slug: "nyc1",
				},
			}, nil, nil
		}

		// Call the method
		droplet, _, err := provider.doClient.Droplets().Get(context.Background(), 12345)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, droplet)
		assert.Equal(t, 12345, droplet.ID)
		assert.Equal(t, "test-droplet", droplet.Name)
	})

	t.Run("Get_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
			return nil, nil, errors.New("API error")
		}

		// Call the method
		droplet, _, err := provider.doClient.Droplets().Get(context.Background(), 12345)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, droplet)
	})

	t.Run("List_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mocks
		mockClient.mockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
			return []godo.Droplet{
				{ID: 12345, Name: "test-1"},
				{ID: 67890, Name: "test-2"},
			}, nil, nil
		}

		// Call the method
		droplets, _, err := provider.doClient.Droplets().List(context.Background(), &godo.ListOptions{})

		// Verify results
		assert.NoError(t, err)
		assert.Len(t, droplets, 2)
		assert.Equal(t, "test-1", droplets[0].Name)
		assert.Equal(t, "test-2", droplets[1].Name)
	})

	t.Run("List_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
			return nil, nil, errors.New("API error")
		}

		// Call the method
		droplets, _, err := provider.doClient.Droplets().List(context.Background(), &godo.ListOptions{})

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, droplets)
	})
}

// KeyService interface and implementations tests

// TestKeyService tests the KeyService interface implementation
func TestKeyService(t *testing.T) {
	t.Run("List_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock
		expectedKeys := []godo.Key{
			{ID: 12345, Name: "test-key-1", PublicKey: "ssh-rsa AAAAB..."},
			{ID: 67890, Name: "test-key-2", PublicKey: "ssh-rsa BBBBB..."},
		}

		setupMockSSHKeyList(mockClient, expectedKeys)

		// Call the method
		keys, _, err := provider.doClient.Keys().List(context.Background(), &godo.ListOptions{})

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, expectedKeys, keys)
	})

	t.Run("List_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockKeyService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
			return nil, nil, errors.New("API error")
		}

		// Call the method
		keys, _, err := provider.doClient.Keys().List(context.Background(), &godo.ListOptions{})

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, keys)
	})
}

// DigitalOceanProvider struct and methods tests

// TestDigitalOceanProvider tests the basic provider functionality
func TestDigitalOceanProvider(t *testing.T) {
	// Save original env var
	originalToken := os.Getenv("DIGITALOCEAN_TOKEN")
	defer func() {
		err := os.Setenv("DIGITALOCEAN_TOKEN", originalToken)
		if err != nil {
			t.Logf("Failed to restore DIGITALOCEAN_TOKEN: %v", err)
		}
	}()

	t.Run("ValidToken", func(t *testing.T) {
		// Set env var for test
		err := os.Setenv("DIGITALOCEAN_TOKEN", "test-token")
		require.NoError(t, err)

		provider, err := NewDigitalOceanProvider()
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		// Clear env var for test
		err := os.Setenv("DIGITALOCEAN_TOKEN", "")
		require.NoError(t, err)

		provider, err := NewDigitalOceanProvider()
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "DIGITALOCEAN_TOKEN environment variable is not set")
	})

	t.Run("ConfigureProvider", func(t *testing.T) {
		provider, _ := newTestProvider()
		err := provider.ConfigureProvider(nil)
		assert.NoError(t, err)
	})

	t.Run("CreateDropletRequest", func(t *testing.T) {
		provider, _ := newTestProvider()
		config := InstanceConfig{
			Region:   "nyc1",
			Size:     "s-1vcpu-1gb",
			Image:    "ubuntu-20-04-x64",
			SSHKeyID: "test-key",
		}

		request := provider.CreateDropletRequest("test-instance", config, 12345)
		assert.Equal(t, "test-instance", request.Name)
		assert.Equal(t, "nyc1", request.Region)
		assert.Equal(t, "s-1vcpu-1gb", request.Size)
		assert.Equal(t, "ubuntu-20-04-x64", request.Image.Slug)
		assert.Equal(t, 12345, request.SSHKeys[0].ID)
	})

	t.Run("CreateInstance_SingleInstance", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mocks
		setupMockSSHKeyList(mockClient, []godo.Key{
			{ID: 12345, Name: "test-key"},
		})

		setupMockDropletCreate(mockClient, 54321, "test-instance")

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
	})

	t.Run("CreateInstance_MultipleInstances", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mocks
		setupMockSSHKeyList(mockClient, []godo.Key{
			{ID: 12345, Name: "test-key"},
		})

		mockClient.mockDropletService.CreateMultipleFunc = func(ctx context.Context, req *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
			droplets := make([]godo.Droplet, len(req.Names))
			for i, name := range req.Names {
				droplets[i] = godo.Droplet{
					ID:   10000 + i,
					Name: name,
					Networks: &godo.Networks{
						V4: []godo.NetworkV4{
							{
								Type:      "public",
								IPAddress: fmt.Sprintf("192.0.2.%d", i+1),
							},
						},
					},
					Region: &godo.Region{
						Slug: req.Region,
					},
				}
			}
			return droplets, nil, nil
		}

		// Setup the Get function to return droplets with IPs
		mockClient.mockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
			// Calculate the index based on the ID (10000, 10001, 10002)
			index := id - 10000
			if index >= 0 && index < 3 {
				return &godo.Droplet{
					ID:   id,
					Name: fmt.Sprintf("test-instance-%d", index),
					Networks: &godo.Networks{
						V4: []godo.NetworkV4{
							{
								Type:      "public",
								IPAddress: fmt.Sprintf("192.0.2.%d", index+1),
							},
						},
					},
					Region: &godo.Region{
						Slug: "nyc1",
					},
				}, nil, nil
			}
			return nil, nil, fmt.Errorf("droplet not found")
		}

		// Create multiple instances
		config := InstanceConfig{
			Region:            "nyc1",
			Size:              "s-1vcpu-1gb",
			Image:             "ubuntu-20-04-x64",
			SSHKeyID:          "test-key",
			NumberOfInstances: 3,
		}

		instances, err := provider.CreateInstance(context.Background(), "test-instance", config)
		assert.NoError(t, err)
		assert.Len(t, instances, 3)
		assert.Equal(t, "test-instance-0", instances[0].Name)
		assert.Equal(t, "test-instance-1", instances[1].Name)
		assert.Equal(t, "test-instance-2", instances[2].Name)
	})

	t.Run("DeleteInstance", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mocks
		var listCallCount int
		mockClient.mockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
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

		mockClient.mockDropletService.DeleteFunc = func(ctx context.Context, id int) (*godo.Response, error) {
			return nil, nil
		}

		// Delete instance
		err := provider.DeleteInstance(context.Background(), "test-instance", "nyc1")
		assert.NoError(t, err)
	})

	t.Run("GetEnvironmentVars", func(t *testing.T) {
		// Set env var for test
		err := os.Setenv("DIGITALOCEAN_TOKEN", "test-token")
		require.NoError(t, err)

		provider, _ := newTestProvider()
		envVars := provider.GetEnvironmentVars()
		assert.Equal(t, "test-token", envVars["DIGITALOCEAN_TOKEN"])
	})

	t.Run("GetSSHKeyID_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock
		setupMockSSHKeyList(mockClient, []godo.Key{
			{ID: 12345, Name: "test-key"},
		})

		// Call the method
		id, err := provider.GetSSHKeyID(context.Background(), "test-key")

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, 12345, id)
	})

	t.Run("GetSSHKeyID_KeyNotFound", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock
		setupMockSSHKeyList(mockClient, []godo.Key{
			{ID: 12345, Name: "other-key"},
		})

		// Call the method
		_, err := provider.GetSSHKeyID(context.Background(), "test-key")

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SSH key 'test-key' not found")
	})

	t.Run("GetSSHKeyID_NoKeys", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock
		setupMockSSHKeyList(mockClient, []godo.Key{})

		// Call the method
		_, err := provider.GetSSHKeyID(context.Background(), "test-key")

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no SSH keys found")
	})

	t.Run("GetSSHKeyID_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockKeyService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
			return nil, nil, errors.New("API error")
		}

		// Call the method
		_, err := provider.GetSSHKeyID(context.Background(), "test-key")

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list SSH keys")
	})

	t.Run("SetClient", func(t *testing.T) {
		provider := &DigitalOceanProvider{}
		mockClient := newMockDOClient()

		// Initially the client should be nil
		assert.Nil(t, provider.doClient)

		// Set the client
		provider.SetClient(mockClient)

		// Verify the client was set
		assert.NotNil(t, provider.doClient)
		assert.Equal(t, mockClient, provider.doClient)
	})

	t.Run("ValidateCredentials", func(t *testing.T) {
		// Base case should be valid
		provider, _ := newTestProvider()
		err := provider.ValidateCredentials()
		assert.NoError(t, err)

		// Case where client is not initialized
		provider = &DigitalOceanProvider{}
		err = provider.ValidateCredentials()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client not initialized")
	})

	t.Run("WaitForDeletion_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return no droplets (already deleted)
		mockClient.mockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
			return []godo.Droplet{}, nil, nil
		}

		// Call the unexported method directly with a short interval
		err := provider.waitForDeletion(context.Background(), "test-instance", "nyc1", 1, 100*time.Millisecond)

		// Verify results
		assert.NoError(t, err)
	})

	t.Run("WaitForDeletion_StillExists", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return the droplet (not deleted yet)
		mockClient.mockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
			return []godo.Droplet{
				{
					ID:   12345,
					Name: "test-instance",
					Region: &godo.Region{
						Slug: "nyc1",
					},
				},
			}, nil, nil
		}

		// Call the unexported method directly with a short interval
		err := provider.waitForDeletion(context.Background(), "test-instance", "nyc1", 1, 100*time.Millisecond)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "still exists after")
	})

	t.Run("WaitForDeletion_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockDropletService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
			return nil, nil, errors.New("API error")
		}

		// Call the unexported method directly with a short interval
		err := provider.waitForDeletion(context.Background(), "test-instance", "nyc1", 1, 100*time.Millisecond)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list droplets")
	})

	t.Run("WaitForIP_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock
		mockClient.mockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
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

		// Call the unexported method directly with a short interval
		ip, err := provider.waitForIP(context.Background(), 12345, 1, 100*time.Millisecond)

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, "192.0.2.1", ip)
	})

	t.Run("WaitForIP_NoPublicIP", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return a droplet with no public IP
		mockClient.mockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
			return &godo.Droplet{
				Networks: &godo.Networks{
					V4: []godo.NetworkV4{
						{
							Type:      "private",
							IPAddress: "10.0.0.1",
						},
					},
				},
			}, nil, nil
		}

		// Call the unexported method directly with a short interval
		_, err := provider.waitForIP(context.Background(), 12345, 1, 100*time.Millisecond)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no public IP found after")
	})

	t.Run("WaitForIP_Error", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
			return nil, nil, errors.New("API error")
		}

		// Call the unexported method directly with a short interval
		_, err := provider.waitForIP(context.Background(), 12345, 1, 100*time.Millisecond)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get droplet details")
	})
}
