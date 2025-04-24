package compute

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/test/mocks"
)

// Test helper functions

// newTestProvider creates a new DigitalOceanProvider with a mock client for testing
func newTestProvider() (*DigitalOceanProvider, *mocks.MockDOClient) {
	mockClient := mocks.NewMockDOClient()
	mockClient.ResetToStandard() // Reset to standard responses
	provider := &DigitalOceanProvider{doClient: mockClient}
	return provider, mockClient
}

// Tests grouped by interface/struct implementation

// DOClient interface and implementations tests

// TestDOClient tests the DOClient interface implementation
func TestDOClient(t *testing.T) {
	client := NewDOClient("test-token")
	assert.NotNil(t, client)

	dropletService := client.Droplets()
	assert.NotNil(t, dropletService)

	keyService := client.Keys()
	assert.NotNil(t, keyService)
}

// DropletService interface and implementations tests

// TestDropletService tests the DropletService interface implementation
func TestDropletService(t *testing.T) {
	t.Run("Create_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Call the method - will use standard success response
		droplet, _, err := provider.doClient.Droplets().Create(context.Background(), &godo.DropletCreateRequest{
			Name:   "test-droplet",
			Region: "nyc1",
		})

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, droplet)
		assert.Equal(t, mockClient.StandardResponses.Droplets.DefaultDroplet.ID, droplet.ID)
		assert.Equal(t, "test-droplet", droplet.Name)
		assert.Equal(t, "nyc1", droplet.Region.Slug)
	})

	t.Run("Create_Error_RateLimit", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Simulate rate limit error
		mockClient.SimulateRateLimit()

		// Call the method
		droplet, _, err := provider.doClient.Droplets().Create(context.Background(), &godo.DropletCreateRequest{
			Name:   "test-droplet",
			Region: "nyc1",
		})

		// Verify results
		assert.Error(t, err)
		assert.Equal(t, mockClient.StandardResponses.Droplets.RateLimitError, err)
		assert.Nil(t, droplet)
	})

	t.Run("Create_Error_Authentication", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Simulate authentication error
		mockClient.SimulateAuthenticationFailure()

		// Call the method
		droplet, _, err := provider.doClient.Droplets().Create(context.Background(), &godo.DropletCreateRequest{
			Name:   "test-droplet",
			Region: "nyc1",
		})

		// Verify results
		assert.Error(t, err)
		assert.Equal(t, mockClient.StandardResponses.Droplets.AuthenticationError, err)
		assert.Nil(t, droplet)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		provider, _ := newTestProvider()

		// Call the method - will use standard success response
		_, err := provider.doClient.Droplets().Delete(context.Background(), 12345)

		// Verify results
		assert.NoError(t, err)
	})

	t.Run("Delete_Error_NotFound", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Simulate not found error
		mockClient.SimulateNotFound()

		// Call the method
		_, err := provider.doClient.Droplets().Delete(context.Background(), 12345)

		// Verify results
		assert.Error(t, err)
		assert.Equal(t, mockClient.StandardResponses.Droplets.NotFoundError, err)
	})

	t.Run("Get_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Call the method - will use standard success response
		droplet, _, err := provider.doClient.Droplets().Get(context.Background(), 12345)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, droplet)
		assert.Equal(t, mockClient.StandardResponses.Droplets.DefaultDroplet.ID, droplet.ID)
	})

	t.Run("Get_Error_NotFound", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Simulate not found error
		mockClient.SimulateNotFound()

		// Call the method
		droplet, _, err := provider.doClient.Droplets().Get(context.Background(), 12345)

		// Verify results
		assert.Error(t, err)
		assert.Equal(t, mockClient.StandardResponses.Droplets.NotFoundError, err)
		assert.Nil(t, droplet)
	})
}

// KeyService interface and implementations tests

func TestKeyService(t *testing.T) {
	t.Run("List_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Call the method - will use standard success response
		keys, _, err := provider.doClient.Keys().List(context.Background(), nil)

		// Verify results
		assert.NoError(t, err)
		assert.NotNil(t, keys)
		assert.Equal(t, mockClient.StandardResponses.Keys.DefaultKeyList, keys)
	})

	t.Run("List_Error_RateLimit", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Simulate rate limit error
		mockClient.SimulateRateLimit()

		// Call the method
		keys, _, err := provider.doClient.Keys().List(context.Background(), nil)

		// Verify results
		assert.Error(t, err)
		assert.Equal(t, mockClient.StandardResponses.Keys.RateLimitError, err)
		assert.Nil(t, keys)
	})

	t.Run("List_Error_Authentication", func(t *testing.T) {
		provider, mockClient := newTestProvider()

		// Simulate authentication error
		mockClient.SimulateAuthenticationFailure()

		// Call the method
		keys, _, err := provider.doClient.Keys().List(context.Background(), nil)

		// Verify results
		assert.Error(t, err)
		assert.Equal(t, mockClient.StandardResponses.Keys.AuthenticationError, err)
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
		config := types.InstanceRequest{
			Region:     "nyc1",
			Size:       "s-1vcpu-1gb",
			Image:      "ubuntu-20-04-x64",
			SSHKeyName: "test-key",
		}

		request := provider.createDropletRequest(&config, 12345)
		assert.Equal(t, config.Name, request.Name)
		assert.Equal(t, config.Region, request.Region)
		assert.Equal(t, config.Size, request.Size)
		assert.Equal(t, config.Image, request.Image)
		assert.Equal(t, config.SSHKeyName, request.SSHKeys[0].ID)
	})

	t.Run("CreateInstance_SingleInstance", func(t *testing.T) {
		provider, _ := newTestProvider()
		keys, _, err := provider.doClient.Keys().List(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, keys)

		// Create instance
		config := types.InstanceRequest{
			Region:     "nyc1",
			Size:       "s-1vcpu-1gb",
			Image:      "ubuntu-20-04-x64",
			SSHKeyName: keys[0].Name,
		}

		err = provider.CreateInstance(context.Background(), &config)
		assert.NoError(t, err)
		assert.Equal(t, "test-instance", config.Name)
		assert.Equal(t, mocks.DefaultDropletIP1, config.PublicIP)
		assert.Equal(t, fmt.Sprintf("%d", mocks.DefaultDropletID1), config.ProviderInstanceID)
	})

	t.Run("DeleteInstance", func(t *testing.T) {
		provider, _ := newTestProvider()

		// Call the method - will use standard success response
		_, err := provider.doClient.Droplets().Delete(context.Background(), 12345)

		// Verify results
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

	t.Run("CreateInstance_SSHKey_Success", func(t *testing.T) {
		provider, mockClient := newTestProvider()
		keys, _, err := mockClient.Keys().List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, keys)
		require.NotEmpty(t, keys)

		config := types.InstanceRequest{
			Region:            "nyc1",
			Size:              "s-1vcpu-1gb",
			Image:             "ubuntu-22-04-x64",
			SSHKeyName:        keys[0].Name,
			NumberOfInstances: 1,
		}

		// Call CreateInstance which internally uses getSSHKeyID
		err = provider.CreateInstance(context.Background(), &config)

		// Verify results
		assert.NoError(t, err)
		assert.Equal(t, "test-instance", config.Name)
		assert.Equal(t, mocks.DefaultDropletIP1, config.PublicIP)
		assert.Equal(t, fmt.Sprintf("%d", mocks.DefaultDropletID1), config.ProviderInstanceID)
	})

	t.Run("CreateInstance_SSHKey_NotFound", func(t *testing.T) {
		provider, _ := newTestProvider()

		config := types.InstanceRequest{
			Region:            "nyc1",
			Size:              "s-1vcpu-1gb",
			Image:             "ubuntu-22-04-x64",
			SSHKeyName:        "not-existing-key",
			NumberOfInstances: 1,
		}

		// Call CreateInstance which internally uses getSSHKeyID
		err := provider.CreateInstance(context.Background(), &config)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get SSH key")
	})

	t.Run("SetClient", func(t *testing.T) {
		provider := &DigitalOceanProvider{}
		mockClient := mocks.NewMockDOClient()

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

	t.Run("DeleteInstance_Success", func(t *testing.T) {
		provider, _ := newTestProvider()

		// Call DeleteInstance which internally uses waitForDeletion
		err := provider.DeleteInstance(context.Background(), 12345)

		// Verify results
		assert.NoError(t, err)
	})

	t.Run("CreateInstance_Success_With_IP", func(t *testing.T) {
		provider, _ := newTestProvider()

		config := types.InstanceRequest{
			Region:            "nyc1",
			Size:              "s-1vcpu-1gb",
			Image:             "ubuntu-22-04-x64",
			SSHKeyName:        "test-key",
			NumberOfInstances: 1,
		}

		// Call CreateInstance which internally uses waitForIP
		err := provider.CreateInstance(context.Background(), &config)

		// Verify results
		assert.NoError(t, err)
		// The following fields should have been updated during the create process
		assert.NotEmpty(t, config.PublicIP)
		assert.NotEmpty(t, config.ProviderInstanceID)
		assert.NotEmpty(t, config.VolumeIDs)
		assert.NotEmpty(t, config.VolumeDetails)
	})

	t.Run("GetDroplet", func(t *testing.T) {
		mockClient := mocks.NewMockDOClient()
		provider := &DigitalOceanProvider{doClient: mockClient}

		t.Run("droplet not found", func(t *testing.T) {
			// Simular error de droplet no encontrado
			mockClient.SimulateNotFound()

			droplet, _, err := provider.doClient.Droplets().Get(context.Background(), 123)
			assert.Error(t, err)
			assert.Nil(t, droplet)
			assert.Equal(t, mocks.ErrDropletNotFound, err)
		})

		t.Run("successful get", func(t *testing.T) {
			// Resetear a respuestas estándar
			mockClient.ResetToStandard()

			droplet, _, err := provider.doClient.Droplets().Get(context.Background(), 123)
			assert.NoError(t, err)
			assert.NotNil(t, droplet)
		})
	})

	t.Run("DeleteDroplet", func(t *testing.T) {
		mockClient := mocks.NewMockDOClient()
		provider := &DigitalOceanProvider{doClient: mockClient}

		t.Run("droplet not found", func(t *testing.T) {
			// Simular error de droplet no encontrado
			mockClient.SimulateNotFound()

			_, err := provider.doClient.Droplets().Delete(context.Background(), 123)
			assert.Error(t, err)
			assert.Equal(t, mocks.ErrDropletNotFound, err)
		})

		t.Run("successful delete", func(t *testing.T) {
			// Resetear a respuestas estándar
			mockClient.ResetToStandard()

			_, err := provider.doClient.Droplets().Delete(context.Background(), 123)
			assert.NoError(t, err)
		})
	})
}
