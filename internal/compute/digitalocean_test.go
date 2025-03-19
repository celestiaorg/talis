package compute

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions

// newTestProvider creates a DigitalOceanProvider with a mock client for testing
func newTestProvider() (*DigitalOceanProvider, *mockDOClient) {
	mockClient := NewMockDOClient().(*mockDOClient)
	provider := &DigitalOceanProvider{
		doClient: &godo.Client{
			Droplets: mockClient.mockDropletService,
			Keys:     mockClient.mockKeyService,
		},
	}
	return provider, mockClient
}

// Tests grouped by interface/struct implementation

// TestDOClient tests the DOClient interface implementation
func TestDOClient(t *testing.T) {
	t.Run("MockDOClient", func(t *testing.T) {
		mockClient := NewMockDOClient()

		dropletService := mockClient.Droplets()
		assert.NotNil(t, dropletService)

		keyService := mockClient.Keys()
		assert.NotNil(t, keyService)
	})

	t.Run("NewProvider", func(t *testing.T) {
		os.Setenv("DIGITALOCEAN_TOKEN", "test-token")
		provider, err := NewDigitalOceanProvider()
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		os.Unsetenv("DIGITALOCEAN_TOKEN")
	})
}

// DropletService interface and implementations tests

// TestDropletService tests the DropletService interface implementation
func TestDropletService(t *testing.T) {
	t.Run("Create_Success", func(t *testing.T) {
		_, mockClient := newTestProvider()

		droplet := &godo.Droplet{
			ID:   12345,
			Name: "test-droplet",
			Region: &godo.Region{
				Slug: "nyc1",
			},
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{
					{
						Type:      "public",
						IPAddress: "192.0.2.1",
					},
				},
			},
		}

		mockClient.mockDropletService.droplets[droplet.ID] = droplet

		// Call the method
		result, _, err := mockClient.mockDropletService.Create(context.Background(), &godo.DropletCreateRequest{
			Name:   "test-droplet",
			Region: "nyc1",
		})

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, droplet.ID, result.ID)
		assert.Equal(t, droplet.Name, result.Name)
	})

	t.Run("Create_Error", func(t *testing.T) {
		_, mockClient := newTestProvider()

		// Setup mock to return an error
		mockClient.mockDropletService.droplets = nil

		// Call the method
		droplet, _, err := mockClient.mockDropletService.Create(context.Background(), &godo.DropletCreateRequest{
			Name:   "test-droplet",
			Region: "nyc1",
		})

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, droplet)
	})

	t.Run("CreateMultiple_Success", func(t *testing.T) {
		_, mockClient := newTestProvider()

		droplets := []godo.Droplet{
			{
				ID:   10000,
				Name: "test-1",
				Region: &godo.Region{
					Slug: "nyc1",
				},
			},
			{
				ID:   10001,
				Name: "test-2",
				Region: &godo.Region{
					Slug: "nyc1",
				},
			},
		}

		for _, d := range droplets {
			mockClient.mockDropletService.droplets[d.ID] = &d
		}

		// Call the method
		result, _, err := mockClient.mockDropletService.CreateMultiple(context.Background(), &godo.DropletMultiCreateRequest{
			Names:  []string{"test-1", "test-2"},
			Region: "nyc1",
		})

		// Verify results
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, droplets[0].Name, result[0].Name)
		assert.Equal(t, droplets[1].Name, result[1].Name)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		_, mockClient := newTestProvider()

		droplet := &godo.Droplet{
			ID:   12345,
			Name: "test-droplet",
		}
		mockClient.mockDropletService.droplets[droplet.ID] = droplet

		// Call the method
		_, err := mockClient.mockDropletService.Delete(context.Background(), droplet.ID)

		// Verify results
		assert.NoError(t, err)
		_, exists := mockClient.mockDropletService.droplets[droplet.ID]
		assert.False(t, exists)
	})

	t.Run("Get_Success", func(t *testing.T) {
		_, mockClient := newTestProvider()

		droplet := &godo.Droplet{
			ID:   12345,
			Name: "test-droplet",
			Region: &godo.Region{
				Slug: "nyc1",
			},
		}
		mockClient.mockDropletService.droplets[droplet.ID] = droplet

		// Call the method
		result, _, err := mockClient.mockDropletService.Get(context.Background(), droplet.ID)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, droplet.ID, result.ID)
		assert.Equal(t, droplet.Name, result.Name)
	})

	t.Run("Get_Error", func(t *testing.T) {
		_, mockClient := newTestProvider()

		// Call the method with non-existent ID
		droplet, _, err := mockClient.mockDropletService.Get(context.Background(), 99999)

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, droplet)
	})

	t.Run("GetBackupPolicy_Success", func(t *testing.T) {
		_, mockClient := newTestProvider()

		expectedPolicy := &godo.DropletBackupPolicy{
			DropletID:     1,
			BackupEnabled: true,
			BackupPolicy: &godo.DropletBackupPolicyConfig{
				Plan:                "weekly",
				Weekday:             "monday",
				Hour:                2,
				WindowLengthHours:   4,
				RetentionPeriodDays: 30,
			},
			NextBackupWindow: &godo.BackupWindow{},
		}

		mockClient.mockDropletService.GetBackupPolicyFunc = func(ctx context.Context, dropletID int) (*godo.DropletBackupPolicy, *godo.Response, error) {
			return expectedPolicy, &godo.Response{}, nil
		}

		policy, _, err := mockClient.mockDropletService.GetBackupPolicy(context.Background(), 1)

		assert.NoError(t, err)
		assert.NotNil(t, policy)
		assert.Equal(t, expectedPolicy.DropletID, policy.DropletID)
		assert.Equal(t, expectedPolicy.BackupEnabled, policy.BackupEnabled)
		assert.Equal(t, expectedPolicy.BackupPolicy.Plan, policy.BackupPolicy.Plan)
		assert.Equal(t, expectedPolicy.BackupPolicy.Weekday, policy.BackupPolicy.Weekday)
		assert.Equal(t, expectedPolicy.BackupPolicy.Hour, policy.BackupPolicy.Hour)
	})

	t.Run("GetBackupPolicy_Error", func(t *testing.T) {
		_, mockClient := newTestProvider()

		mockClient.mockDropletService.GetBackupPolicyFunc = func(ctx context.Context, dropletID int) (*godo.DropletBackupPolicy, *godo.Response, error) {
			return nil, &godo.Response{}, fmt.Errorf("failed to get backup policy")
		}

		policy, _, err := mockClient.mockDropletService.GetBackupPolicy(context.Background(), 1)

		assert.Error(t, err)
		assert.Nil(t, policy)
		assert.Contains(t, err.Error(), "failed to get backup policy")
	})
}

// KeyService interface and implementations tests

// TestKeyService tests the KeyService interface implementation
func TestKeyService(t *testing.T) {
	t.Run("List_Success", func(t *testing.T) {
		_, mockClient := newTestProvider()

		expectedKeys := []godo.Key{
			{ID: 12345, Name: "test-key-1", PublicKey: "ssh-rsa AAAAB..."},
			{ID: 67890, Name: "test-key-2", PublicKey: "ssh-rsa BBBBB..."},
		}

		for _, key := range expectedKeys {
			mockClient.mockKeyService.keys[key.ID] = &key
		}

		// Call the method
		keys, _, err := mockClient.mockKeyService.List(context.Background(), &godo.ListOptions{})

		// Verify results
		assert.NoError(t, err)
		assert.Len(t, keys, len(expectedKeys))
		for i, key := range keys {
			assert.Equal(t, expectedKeys[i].ID, key.ID)
			assert.Equal(t, expectedKeys[i].Name, key.Name)
		}
	})

	t.Run("List_Error", func(t *testing.T) {
		_, mockClient := newTestProvider()

		// Clear the keys to simulate error
		mockClient.mockKeyService.keys = nil

		// Call the method
		keys, _, err := mockClient.mockKeyService.List(context.Background(), &godo.ListOptions{})

		// Verify results
		assert.Error(t, err)
		assert.Nil(t, keys)
	})
}

// DigitalOceanProvider struct and methods tests

// TestDigitalOceanProvider tests the provider functionality
func TestDigitalOceanProvider(t *testing.T) {
	t.Run("CreateInstance", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Setup test data
		sshKey := &godo.Key{
			ID:   12345,
			Name: "test-key",
		}
		mockClient.mockKeyService.keys[sshKey.ID] = sshKey

		droplets := []*godo.Droplet{
			{
				ID:   10000,
				Name: "test-instance-0",
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
			},
			{
				ID:   10001,
				Name: "test-instance-1",
				Networks: &godo.Networks{
					V4: []godo.NetworkV4{
						{
							Type:      "public",
							IPAddress: "192.0.2.2",
						},
					},
				},
				Region: &godo.Region{
					Slug: "nyc1",
				},
			},
		}

		mockClient.mockDropletService.CreateFunc = func(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
			for _, d := range droplets {
				if d.Name == req.Name {
					mockClient.mockDropletService.droplets[d.ID] = d
					return d, &godo.Response{}, nil
				}
			}
			return nil, nil, fmt.Errorf("droplet not found")
		}

		// Create multiple instances
		config := InstanceConfig{
			Region:            "nyc1",
			Size:              "s-1vcpu-1gb",
			Image:             "ubuntu-20-04-x64",
			SSHKeyID:          "test-key",
			NumberOfInstances: 2,
		}

		instances, err := provider.CreateInstance(context.Background(), "test-instance", config)
		assert.NoError(t, err)
		assert.Len(t, instances, 2)
		assert.Equal(t, "test-instance-0", instances[0].Name)
		assert.Equal(t, "test-instance-1", instances[1].Name)
	})

	t.Run("DeleteInstance", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		droplet := &godo.Droplet{
			ID:   54321,
			Name: "test-instance",
			Region: &godo.Region{
				Slug: "nyc1",
			},
		}
		mockClient.mockDropletService.droplets[droplet.ID] = droplet

		// Delete instance
		err := provider.DeleteInstance(context.Background(), "test-instance", "nyc1")
		assert.NoError(t, err)
		_, exists := mockClient.mockDropletService.droplets[droplet.ID]
		assert.False(t, exists)
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

		key := &godo.Key{
			ID:   12345,
			Name: "test-key",
		}
		mockClient.mockKeyService.keys[key.ID] = key

		// Call the method
		id, err := provider.getSSHKeyID(context.Background(), "test-key")

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, 12345, id)
	})

	t.Run("GetSSHKeyID_KeyNotFound", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		key := &godo.Key{
			ID:   12345,
			Name: "other-key",
		}
		mockClient.mockKeyService.keys[key.ID] = key

		// Call the method
		_, err := provider.getSSHKeyID(context.Background(), "test-key")

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SSH key 'test-key' not found")
	})
}
