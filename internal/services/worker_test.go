package services

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
)

func TestWorker_getProvider(t *testing.T) {
	// Create a worker with nil services as they are not used by getProvider
	// Use a short backoff for testing purposes if needed, though not relevant here.
	w := NewWorkerPool(nil, nil, nil, nil, nil, time.Millisecond*10)

	// Define provider IDs for testing
	// Assuming "digitalocean-mock" is a valid provider ID that compute.NewComputeProvider can handle
	validProviderID := models.ProviderID("digitalocean-mock")
	// Assuming this ID will cause compute.NewComputeProvider to return an error
	invalidProviderID := models.ProviderID("invalid-provider-id-for-test")

	t.Run("Basic", func(t *testing.T) {
		provider, err := w.getProvider(validProviderID)
		require.NoError(t, err, "Getting a valid provider should not return an error")
		require.NotNil(t, provider, "Provider instance should not be nil for a valid ID")

		// Store the instance for comparison in the next test
		firstProviderInstance := provider

		// Check internal state
		w.computeMU.RLock()
		cachedProvider, exists := w.providers[validProviderID]
		w.computeMU.RUnlock()
		require.True(t, exists, "Provider should be cached after first retrieval")
		require.Same(t, firstProviderInstance, cachedProvider, "Cached provider instance should match the returned one")

		provider, err = w.getProvider(validProviderID)
		require.NoError(t, err, "Getting a cached valid provider should not return an error")
		require.NotNil(t, provider, "Cached provider instance should not be nil")

		// Crucial check: Ensure it's the *same* instance, not a new one
		require.Same(t, firstProviderInstance, provider, "Second call for the same provider ID should return the identical cached instance")

		// Get the invalid provider
		provider, err = w.getProvider(invalidProviderID)

		// Check that an error occurred
		require.Error(t, err, "Getting an invalid provider should return an error")
		require.Nil(t, provider, "Provider instance should be nil when an error occurs")

		// Check if the error message indicates an unsupported provider (adjust based on actual error)
		// This depends on the error returned by compute.NewComputeProvider for invalid IDs
		require.Contains(t, err.Error(), "unsupported provider", "Error message should indicate an unsupported provider")

		// Check internal state
		w.computeMU.RLock()
		_, exists = w.providers[invalidProviderID]
		w.computeMU.RUnlock()
		require.False(t, exists, "Invalid provider should not be added to the cache")
	})

	// Test concurrent access to the provider cache, rapidly accessing the provider map
	t.Run("Concurrency", func(_ *testing.T) {
		// Test concurrent access to the provider cache
		numGoroutines := 10
		wg := sync.WaitGroup{}
		wg.Add(numGoroutines)
		done := make(chan struct{})
		go func() {
			time.Sleep(time.Second)
			close(done)
		}()
		for i := 0; i < numGoroutines; i++ {
			providerID := validProviderID
			if i >= numGoroutines/2 {
				providerID = invalidProviderID
			}
			go func(done chan struct{}, providerID models.ProviderID) {
				defer wg.Done()
				for {
					select {
					case <-done:
						return
					default:
						// We are only testing the concurrency of the provider cache, not the correctness of the provider
						_, _ = w.getProvider(providerID)
					}
				}
			}(done, providerID)
		}
		wg.Wait()
	})
}

func TestWorker_getProvisioner(t *testing.T) {
	// Create a worker with nil services as they are not used by getProvisioner
	w := NewWorkerPool(nil, nil, nil, nil, nil, time.Millisecond*10)

	// Define a provider ID for testing. getProvisioner works with any valid ProviderID.
	// We'll use the same mock ID as in the provider test for consistency.
	providerID := models.ProviderID("digitalocean-mock")

	t.Run("Basic", func(t *testing.T) {
		// --- Get provisioner for the first time ---
		provisioner, err := w.getProvisioner(providerID)
		require.NoError(t, err, "Getting a valid provisioner should not return an error")
		require.NotNil(t, provisioner, "Provisioner instance should not be nil")

		// Store the instance for comparison
		firstProvisionerInstance := provisioner

		// Check internal state
		w.computeMU.RLock()
		cachedProvisioner, exists := w.provisioners[providerID]
		w.computeMU.RUnlock()
		require.True(t, exists, "Provisioner should be cached after first retrieval")
		require.Same(t, firstProvisionerInstance, cachedProvisioner, "Cached provisioner instance should match the returned one")

		// --- Get the same provisioner again (should return cached instance) ---
		provisioner, err = w.getProvisioner(providerID)
		require.NoError(t, err, "Getting a cached valid provisioner should not return an error")
		require.NotNil(t, provisioner, "Cached provisioner instance should not be nil")

		// Crucial check: Ensure it's the *same* instance, not a new one
		require.Same(t, firstProvisionerInstance, provisioner, "Second call for the same provider ID should return the identical cached instance")
	})

	// Test concurrent access to the provisioner cache, rapidly accessing the map
	t.Run("Concurrency", func(_ *testing.T) {
		numGoroutines := 10
		wg := sync.WaitGroup{}
		wg.Add(numGoroutines)
		done := make(chan struct{}) // Signal channel to stop goroutines

		// Goroutine to close the done channel after a short delay
		go func() {
			time.Sleep(time.Second) // Let goroutines run for a bit
			close(done)
		}()

		// Launch goroutines that repeatedly call getProvisioner
		for i := 0; i < numGoroutines; i++ {
			go func(stop <-chan struct{}) {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return // Stop signal received
					default:
						// Repeatedly get the provisioner to test concurrent map access
						_, _ = w.getProvisioner(providerID)
					}
				}
			}(done) // Pass the done channel to the goroutine
		}

		wg.Wait() // Wait for all goroutines to finish
	})
}
