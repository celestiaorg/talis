package compute

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ximeraTypes "github.com/celestiaorg/talis/internal/compute/types"
	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/test/mocks"
)

// Test helper functions

// newTestXimeraProvider creates a new XimeraProvider with a mock client for testing
func newTestXimeraProvider() (*XimeraProvider, *mocks.MockXimeraAPIClient) {
	mockClient := mocks.NewMockXimeraAPIClient()
	provider := &XimeraProvider{
		cache: &XimeraServerCache{
			mutex: sync.RWMutex{},
		},
		cacheTTL: 5 * time.Minute,
	}
	provider.SetClient(mockClient)
	return provider, mockClient
}

// newTestXimeraProviderWithShortTTL creates a provider with a short cache TTL for testing cache expiration
func newTestXimeraProviderWithShortTTL() (*XimeraProvider, *mocks.MockXimeraAPIClient) {
	mockClient := mocks.NewMockXimeraAPIClient()
	provider := &XimeraProvider{
		cache: &XimeraServerCache{
			mutex: sync.RWMutex{},
		},
		cacheTTL: 100 * time.Millisecond, // Very short TTL for testing
	}
	provider.SetClient(mockClient)
	return provider, mockClient
}

// Tests grouped by interface/struct implementation

// XimeraProvider struct and methods tests

// TestXimeraProvider tests the basic provider functionality
func TestXimeraProvider(t *testing.T) {
	// Save and restore environment variables
	originalAPIURL := os.Getenv("XIMERA_API_URL")
	originalAPIToken := os.Getenv("XIMERA_API_TOKEN")
	originalUserID := os.Getenv("XIMERA_USER_ID")
	originalHypervisorID := os.Getenv("XIMERA_HYPERVISOR_GROUP_ID")
	originalPackageID := os.Getenv("XIMERA_PACKAGE_ID")
	originalSSHKeyName := os.Getenv(constants.EnvTalisSSHKeyName)

	defer func() {
		os.Setenv("XIMERA_API_URL", originalAPIURL)
		os.Setenv("XIMERA_API_TOKEN", originalAPIToken)
		os.Setenv("XIMERA_USER_ID", originalUserID)
		os.Setenv("XIMERA_HYPERVISOR_GROUP_ID", originalHypervisorID)
		os.Setenv("XIMERA_PACKAGE_ID", originalPackageID)
		os.Setenv(constants.EnvTalisSSHKeyName, originalSSHKeyName)
	}()

	// Set required environment variables for tests
	os.Setenv("XIMERA_API_URL", "https://api.ximera.test")
	os.Setenv("XIMERA_API_TOKEN", "test-token")
	os.Setenv("XIMERA_USER_ID", "100")
	os.Setenv("XIMERA_HYPERVISOR_GROUP_ID", "10")
	os.Setenv("XIMERA_PACKAGE_ID", "1")
	os.Setenv(constants.EnvTalisSSHKeyName, "123")

	t.Run("NewXimeraProvider_Success", func(t *testing.T) {
		provider, err := NewXimeraProvider()
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.NotNil(t, provider.client)
		assert.NotNil(t, provider.cache)
		assert.Equal(t, 5*time.Minute, provider.cacheTTL)
	})

	t.Run("NewXimeraProvider_MissingConfig", func(t *testing.T) {
		// Clear required env var
		os.Setenv("XIMERA_API_URL", "")

		provider, err := NewXimeraProvider()
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "failed to initialize ximera config")

		// Restore for other tests
		os.Setenv("XIMERA_API_URL", "https://api.ximera.test")
	})

	t.Run("ValidateCredentials_Success", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()
		err := provider.ValidateCredentials()
		assert.NoError(t, err)
	})

	t.Run("ValidateCredentials_Error", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateAuthenticationFailure()

		err := provider.ValidateCredentials()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ximera credential validation failed")
	})

	t.Run("GetEnvironmentVars", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()
		envVars := provider.GetEnvironmentVars()

		expectedVars := []string{
			"XIMERA_API_URL",
			"XIMERA_API_TOKEN",
			"XIMERA_USER_ID",
			"XIMERA_HYPERVISOR_GROUP_ID",
			"XIMERA_PACKAGE_ID",
		}

		for _, envVar := range expectedVars {
			_, exists := envVars[envVar]
			assert.True(t, exists, "Environment variable %s should be present", envVar)
		}
	})

	t.Run("ConfigureProvider", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()
		err := provider.ConfigureProvider(nil)
		assert.NoError(t, err)
	})

	t.Run("SetClient", func(t *testing.T) {
		provider := &XimeraProvider{}
		mockClient := mocks.NewMockXimeraAPIClient()

		// Initially the client should be nil
		assert.Nil(t, provider.client)

		// Set the client
		provider.SetClient(mockClient)

		// Verify the client was set
		assert.NotNil(t, provider.client)
		assert.Equal(t, mockClient, provider.client)
	})
}

// TestXimeraProviderCreateInstance tests the CreateInstance functionality
func TestXimeraProviderCreateInstance(t *testing.T) {
	// Set required environment variables
	os.Setenv(constants.EnvTalisSSHKeyName, "123")
	defer os.Unsetenv(constants.EnvTalisSSHKeyName)

	t.Run("CreateInstance_Success", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		req := &types.InstanceRequest{
			ProjectName: "test-project",
			Image:       "200", // Use string format as expected by Ximera
			Memory:      2048,
			CPU:         2,
			Volumes: []types.VolumeConfig{
				{
					Name:       "test-volume",
					SizeGB:     20,
					MountPoint: "/data",
				},
			},
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.NoError(t, err)
		assert.NotZero(t, req.ProviderInstanceID)
		assert.NotEmpty(t, req.PublicIP)
	})

	t.Run("CreateInstance_NoVolumes", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		req := &types.InstanceRequest{
			ProjectName: "test-project",
			Image:       "200",
			Memory:      2048,
			CPU:         2,
			Volumes:     []types.VolumeConfig{}, // Empty volumes
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no volume details provided")
	})

	t.Run("CreateInstance_MultipleVolumes", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		req := &types.InstanceRequest{
			ProjectName: "test-project",
			Image:       "200",
			Memory:      2048,
			CPU:         2,
			Volumes: []types.VolumeConfig{
				{Name: "volume1", SizeGB: 20},
				{Name: "volume2", SizeGB: 30}, // Multiple volumes not supported
			},
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only one volume is supported")
	})

	t.Run("CreateInstance_MissingMemory", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		req := &types.InstanceRequest{
			ProjectName: "test-project",
			Image:       "200",
			CPU:         2,
			Volumes:     []types.VolumeConfig{{Name: "test", SizeGB: 20}},
			// Memory missing
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "memory is required for Ximera")
	})

	t.Run("CreateInstance_MissingCPU", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		req := &types.InstanceRequest{
			ProjectName: "test-project",
			Image:       "200",
			Memory:      2048,
			Volumes:     []types.VolumeConfig{{Name: "test", SizeGB: 20}},
			// CPU missing
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cpu is required for Ximera")
	})

	t.Run("CreateInstance_MissingSSHKey", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()
		os.Unsetenv(constants.EnvTalisSSHKeyName) // Remove SSH key env var

		req := &types.InstanceRequest{
			ProjectName: "test-project",
			Image:       "200",
			Memory:      2048,
			CPU:         2,
			Volumes:     []types.VolumeConfig{{Name: "test", SizeGB: 20}},
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "environment variable")
		assert.Contains(t, err.Error(), "is not set but required for Ximera")

		// Restore for other tests
		os.Setenv(constants.EnvTalisSSHKeyName, "123")
	})

	t.Run("CreateInstance_CreateServerError", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateServerCreationError()

		req := &types.InstanceRequest{
			ProjectName: "test-project",
			Image:       "200",
			Memory:      2048,
			CPU:         2,
			Volumes:     []types.VolumeConfig{{Name: "test", SizeGB: 20}},
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create ximera server")
	})

	t.Run("CreateInstance_SizeParameterWarning", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		req := &types.InstanceRequest{
			ProjectName: "test-project",
			Image:       "200",
			Memory:      2048,
			CPU:         2,
			Size:        "s-1vcpu-1gb", // This should trigger a warning
			Volumes:     []types.VolumeConfig{{Name: "test", SizeGB: 20}},
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.NoError(t, err) // Should still succeed but with warning
	})
}

// TestXimeraProviderDeleteInstance tests the DeleteInstance functionality
func TestXimeraProviderDeleteInstance(t *testing.T) {
	t.Run("DeleteInstance_Success", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		err := provider.DeleteInstance(context.Background(), mocks.DefaultXimeraServerID1)
		assert.NoError(t, err)
	})

	t.Run("DeleteInstance_NotFound", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateNotFound()

		err := provider.DeleteInstance(context.Background(), 99999)
		assert.Error(t, err)
		assert.Equal(t, mocks.ErrXimeraServerNotFound, err)
	})

	t.Run("DeleteInstance_InvalidatesCache", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		// First, populate the cache
		_, err := provider.getCachedServers()
		assert.NoError(t, err)
		assert.True(t, provider.isCacheValid())

		// Delete an instance, which should invalidate the cache
		err = provider.DeleteInstance(context.Background(), mocks.DefaultXimeraServerID1)
		assert.NoError(t, err)

		// Cache should be invalidated
		assert.False(t, provider.isCacheValid())
	})
}

// TestXimeraProviderCaching tests the caching mechanism
func TestXimeraProviderCaching(t *testing.T) {
	t.Run("getCachedServers_CacheMiss", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()

		// Cache should be empty initially
		assert.False(t, provider.isCacheValid())

		// First call should fetch from API
		servers, err := provider.getCachedServers()
		assert.NoError(t, err)
		assert.NotNil(t, servers)
		assert.Len(t, servers.Data, 2) // Default mock returns 2 servers

		// Cache should now be valid
		assert.True(t, provider.isCacheValid())

		// Verify it was called once
		assert.Equal(t, 1, mockClient.GetAttemptCount())
	})

	t.Run("getCachedServers_CacheHit", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()

		// First call to populate cache
		servers1, err := provider.getCachedServers()
		assert.NoError(t, err)
		assert.NotNil(t, servers1)

		// Reset attempt count to track cache hits
		mockClient.ResetAttemptCount()

		// Second call should use cache
		servers2, err := provider.getCachedServers()
		assert.NoError(t, err)
		assert.NotNil(t, servers2)
		assert.Equal(t, servers1, servers2) // Should be same data

		// Should not have called API again
		assert.Equal(t, 0, mockClient.GetAttemptCount())
	})

	t.Run("getCachedServers_CacheExpiry", func(t *testing.T) {
		provider, mockClient := newTestXimeraProviderWithShortTTL()

		// First call to populate cache
		_, err := provider.getCachedServers()
		assert.NoError(t, err)
		assert.True(t, provider.isCacheValid())

		// Wait for cache to expire
		time.Sleep(150 * time.Millisecond)
		assert.False(t, provider.isCacheValid())

		// Reset attempt count
		mockClient.ResetAttemptCount()

		// Next call should fetch from API again
		_, err = provider.getCachedServers()
		assert.NoError(t, err)

		// Should have called API again
		assert.Equal(t, 1, mockClient.GetAttemptCount())
	})

	t.Run("isCacheValid_EmptyCache", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()
		assert.False(t, provider.isCacheValid())
	})

	t.Run("invalidateCache", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		// Populate cache
		_, err := provider.getCachedServers()
		assert.NoError(t, err)
		assert.True(t, provider.isCacheValid())

		// Invalidate cache
		provider.invalidateCache()
		assert.False(t, provider.isCacheValid())
	})

	t.Run("getCachedServers_APIError", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateNetworkError()

		_, err := provider.getCachedServers()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list ximera servers")
	})
}

// TestXimeraProviderGetInstanceByTag tests the search by tag functionality
func TestXimeraProviderGetInstanceByTag(t *testing.T) {
	t.Run("GetInstanceByTag_Found", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateServerListWithTag() // Use server list containing the tag

		instance, err := provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.Equal(t, mocks.DefaultXimeraServerID2, instance.ProviderInstanceID)
		assert.Equal(t, mocks.TestTagName, instance.Name)
		assert.Equal(t, mocks.DefaultXimeraServerIP2, instance.PublicIP)
	})

	t.Run("GetInstanceByTag_NotFound", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateEmptyServerList() // No servers to search

		instance, err := provider.GetInstanceByTag(context.Background(), "nonexistent-tag")
		assert.Error(t, err)
		assert.Nil(t, instance)
		assert.Contains(t, err.Error(), "instance with tag 'nonexistent-tag' not found")
	})

	t.Run("GetInstanceByTag_NotFoundInNonEmptyList", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()
		// Default server list doesn't contain the tag we're looking for

		instance, err := provider.GetInstanceByTag(context.Background(), "nonexistent-tag")
		assert.Error(t, err)
		assert.Nil(t, instance)
		assert.Contains(t, err.Error(), "instance with tag 'nonexistent-tag' not found")
	})

	t.Run("GetInstanceByTag_UsesCaching", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateServerListWithTag()

		// First call should populate cache
		_, err := provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.NoError(t, err)
		assert.True(t, provider.isCacheValid())

		// Reset attempt count to track cache usage
		mockClient.ResetAttemptCount()

		// Second call should use cache for ListServers but still call GetServer
		_, err = provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.NoError(t, err)

		// Should only have called GetServer (1 attempt), not ListServers again
		assert.Equal(t, 1, mockClient.GetAttemptCount())
	})

	t.Run("GetInstanceByTag_APIError", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateNetworkError()

		instance, err := provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.Error(t, err)
		assert.Nil(t, instance)
		assert.Contains(t, err.Error(), "failed to get ximera servers")
	})

	t.Run("GetInstanceByTag_GetServerDetailsError", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateServerListWithTag()

		// Configure GetServer to fail
		mockClient.GetServerFunc = func(id int) (*ximeraTypes.XimeraServerResponse, error) {
			return nil, mocks.ErrXimeraServerNotFound
		}

		instance, err := provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.Error(t, err)
		assert.Nil(t, instance)
		assert.Contains(t, err.Error(), "failed to get full server details")
	})

	t.Run("GetInstanceByTag_ExactMatch", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()

		// Create a custom server list with partial matches to ensure exact matching
		customServerList := &ximeraTypes.XimeraServersListResponse{
			Data: []struct {
				ID           int    `json:"id"`
				OwnerID      int    `json:"ownerId"`
				HypervisorID int    `json:"hypervisorId"`
				Name         string `json:"name"`
				Hostname     string `json:"hostname"`
				UUID         string `json:"uuid"`
				State        string `json:"state"`
			}{
				{
					ID:   1,
					Name: "talis-123-456-suffix", // Should not match "talis-123-456"
				},
				{
					ID:   2,
					Name: "prefix-talis-123-456", // Should not match "talis-123-456"
				},
				{
					ID:   3,
					Name: mocks.TestTagName, // Exact match - should be found
				},
			},
		}

		mockClient.ListServersFunc = func() (*ximeraTypes.XimeraServersListResponse, error) {
			return customServerList, nil
		}

		instance, err := provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.Equal(t, 3, instance.ProviderInstanceID) // Should find the exact match (ID 3)
	})
}

// TestXimeraProviderConcurrency tests concurrent access to caching
func TestXimeraProviderConcurrency(t *testing.T) {
	t.Run("ConcurrentCacheAccess", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		// Run multiple goroutines accessing cache concurrently
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := provider.getCachedServers()
				results <- err
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})

	t.Run("ConcurrentGetInstanceByTag", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateServerListWithTag()

		// Run multiple goroutines searching by tag concurrently
		const numGoroutines = 5
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
				results <- err
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})
}

// TestXimeraProviderEdgeCases tests edge cases and error scenarios
func TestXimeraProviderEdgeCases(t *testing.T) {
	t.Run("EmptyTagSearch", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		instance, err := provider.GetInstanceByTag(context.Background(), "")
		assert.Error(t, err)
		assert.Nil(t, instance)
		assert.Contains(t, err.Error(), "instance with tag '' not found")
	})

	t.Run("CacheThreadSafety_InvalidateWhileReading", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		// Populate cache
		_, err := provider.getCachedServers()
		assert.NoError(t, err)

		// Simulate concurrent access: one goroutine reading, another invalidating
		done := make(chan bool)
		go func() {
			provider.invalidateCache()
			done <- true
		}()

		// Should not panic or cause race condition
		provider.isCacheValid()

		<-done // Wait for invalidation goroutine to complete
	})

	t.Run("CacheWithNilServers", func(t *testing.T) {
		provider, _ := newTestXimeraProvider()

		// Manually set cache to nil servers
		provider.cache.Servers = nil
		provider.cache.Timestamp = time.Now()

		// Cache should be considered invalid
		assert.False(t, provider.isCacheValid())
	})
}

// TestXimeraProviderIntegration tests integration scenarios
func TestXimeraProviderIntegration(t *testing.T) {
	t.Run("CreateThenFindByTag", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()

		// Set up environment
		os.Setenv(constants.EnvTalisSSHKeyName, "123")
		defer os.Unsetenv(constants.EnvTalisSSHKeyName)

		// Create an instance
		req := &types.InstanceRequest{
			ProjectName: mocks.TestTagName, // Use the tag name as project name
			Image:       "200",
			Memory:      2048,
			CPU:         2,
			Volumes:     []types.VolumeConfig{{Name: "test", SizeGB: 20}},
		}

		err := provider.CreateInstance(context.Background(), req)
		assert.NoError(t, err)

		// Mock the server list to include our created server
		mockClient.SimulateServerListWithTag()

		// Try to find it by tag
		foundInstance, err := provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.NoError(t, err)
		assert.NotNil(t, foundInstance)
		assert.Equal(t, mocks.TestTagName, foundInstance.Name)
	})

	t.Run("DeleteInvalidatesSearchCache", func(t *testing.T) {
		provider, mockClient := newTestXimeraProvider()
		mockClient.SimulateServerListWithTag()

		// First, find instance by tag (populates cache)
		instance, err := provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.NoError(t, err)
		assert.NotNil(t, instance)
		assert.True(t, provider.isCacheValid())

		// Delete the instance (should invalidate cache)
		err = provider.DeleteInstance(context.Background(), instance.ProviderInstanceID)
		assert.NoError(t, err)
		assert.False(t, provider.isCacheValid())

		// Next search should fetch fresh data
		mockClient.ResetAttemptCount()
		mockClient.SimulateEmptyServerList() // Now return empty list

		_, err = provider.GetInstanceByTag(context.Background(), mocks.TestTagName)
		assert.Error(t, err) // Should not find it anymore
		assert.Contains(t, err.Error(), "not found")
	})
}
