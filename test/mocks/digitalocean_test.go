package mocks

import (
	"context"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
)

func TestMockDOClient(t *testing.T) {
	mockClient := NewMockDOClient()
	assert.NotNil(t, mockClient)
	assert.NotNil(t, mockClient.StandardResponses)

	dropletService := mockClient.Droplets()
	assert.NotNil(t, dropletService)

	keyService := mockClient.Keys()
	assert.NotNil(t, keyService)
}

func TestMockDOClientBehavior(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockDOClient()

	// Helper function to call all methods and verify responses
	callAllMethods := func(expectedErr error) {
		t.Helper()

		// Droplet service calls
		droplet, _, err := mockClient.Droplets().Create(ctx, &godo.DropletCreateRequest{
			Name:   "test-droplet",
			Region: "test-region",
			Size:   "test-size",
		})
		if expectedErr != nil {
			assert.Equal(t, expectedErr, err)
			assert.Nil(t, droplet)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, droplet)
			assert.Equal(t, DefaultDropletID1, droplet.ID)
			assert.Equal(t, "test-droplet", droplet.Name)
			assert.Equal(t, "test-region", droplet.Region.Slug)
			assert.Equal(t, "test-size", droplet.Size.Slug)
		}

		droplets, _, err := mockClient.Droplets().CreateMultiple(ctx, &godo.DropletMultiCreateRequest{
			Names:  []string{"test-1", "test-2"},
			Region: "test-region",
			Size:   "test-size",
		})
		if expectedErr != nil {
			assert.Equal(t, expectedErr, err)
			assert.Nil(t, droplets)
		} else {
			assert.NoError(t, err)
			assert.Len(t, droplets, 2)
			assert.Equal(t, "test-1", droplets[0].Name)
			assert.Equal(t, "test-2", droplets[1].Name)
		}

		droplet, _, err = mockClient.Droplets().Get(ctx, DefaultDropletID1)
		if expectedErr != nil {
			assert.Equal(t, expectedErr, err)
			assert.Nil(t, droplet)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, droplet)
			assert.Equal(t, DefaultDropletID1, droplet.ID)
		}

		droplets, _, err = mockClient.Droplets().List(ctx, nil)
		if expectedErr != nil {
			assert.Equal(t, expectedErr, err)
			assert.Nil(t, droplets)
		} else {
			assert.NoError(t, err)
			assert.Len(t, droplets, 2)
			assert.Equal(t, DefaultDropletID1, droplets[0].ID)
			assert.Equal(t, DefaultDropletID2, droplets[1].ID)
		}

		_, err = mockClient.Droplets().Delete(ctx, DefaultDropletID1)
		if expectedErr != nil {
			assert.Equal(t, expectedErr, err)
		} else {
			assert.NoError(t, err)
		}

		// Key service calls
		keys, _, err := mockClient.Keys().List(ctx, nil)
		if expectedErr != nil {
			assert.Equal(t, expectedErr, err)
			assert.Nil(t, keys)
		} else {
			assert.NoError(t, err)
			assert.Len(t, keys, 2)
			assert.Equal(t, DefaultKeyID1, keys[0].ID)
			assert.Equal(t, DefaultKeyID2, keys[1].ID)
		}
	}

	// Test 1: Verify standard success responses
	t.Run("Standard Success Responses", func(_ *testing.T) {
		callAllMethods(nil)
	})

	// Test 2: Verify authentication failure responses
	t.Run("Authentication Failure", func(_ *testing.T) {
		mockClient.SimulateAuthenticationFailure()
		callAllMethods(ErrAuthentication)
	})

	// Test 3: Verify rate limit responses
	t.Run("Rate Limit", func(_ *testing.T) {
		mockClient.SimulateRateLimit()
		callAllMethods(ErrRateLimit)
	})

	// Test 4: Verify not found responses
	t.Run("Not Found", func(t *testing.T) {
		mockClient.SimulateNotFound()

		// Only Get and Delete should return NotFound for droplets
		droplet, _, err := mockClient.Droplets().Get(ctx, DefaultDropletID1)
		assert.Equal(t, ErrDropletNotFound, err)
		assert.Nil(t, droplet)

		_, err = mockClient.Droplets().Delete(ctx, DefaultDropletID1)
		assert.Equal(t, ErrDropletNotFound, err)

		// List should return NotFound for keys
		keys, _, err := mockClient.Keys().List(ctx, nil)
		assert.Equal(t, ErrKeyNotFound, err)
		assert.Nil(t, keys)
	})

	// Test 5: Reset and verify success responses again
	t.Run("Reset to Standard Success", func(_ *testing.T) {
		mockClient.ResetToStandard()
		callAllMethods(nil)
	})
}
