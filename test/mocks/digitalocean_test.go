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
	t.Run("Standard Success Responses", func(t *testing.T) {
		client := NewMockDOClient()
		client.ResetToStandard()

		// Test droplet operations
		droplet, _, err := client.Droplets().Create(context.Background(), &godo.DropletCreateRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, droplet)

		droplets, _, err := client.Droplets().List(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, droplets)

		// Test key operations
		keys, _, err := client.Keys().List(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, keys)

		// Test volume operations
		volume, _, err := client.Storage().CreateVolume(context.Background(), &godo.VolumeCreateRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, volume)

		volumes, _, err := client.Storage().ListVolumes(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, volumes)
	})

	t.Run("Authentication Failure", func(t *testing.T) {
		client := NewMockDOClient()
		client.SimulateAuthenticationFailure()

		_, _, err := client.Droplets().Create(context.Background(), &godo.DropletCreateRequest{})
		assert.Error(t, err)
		assert.Equal(t, ErrAuthentication, err)

		_, _, err = client.Keys().List(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, ErrAuthentication, err)

		_, _, err = client.Storage().CreateVolume(context.Background(), &godo.VolumeCreateRequest{})
		assert.Error(t, err)
		assert.Equal(t, ErrAuthentication, err)
	})

	t.Run("Rate Limit", func(t *testing.T) {
		client := NewMockDOClient()
		client.SimulateRateLimit()

		_, _, err := client.Droplets().Create(context.Background(), &godo.DropletCreateRequest{})
		assert.Error(t, err)
		assert.Equal(t, ErrRateLimit, err)

		_, _, err = client.Keys().List(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, ErrRateLimit, err)

		_, _, err = client.Storage().CreateVolume(context.Background(), &godo.VolumeCreateRequest{})
		assert.Error(t, err)
		assert.Equal(t, ErrRateLimit, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		client := NewMockDOClient()
		client.SimulateNotFound()

		_, _, err := client.Droplets().Create(context.Background(), &godo.DropletCreateRequest{})
		assert.Error(t, err)
		assert.Equal(t, ErrDropletNotFound, err)

		_, _, err = client.Keys().List(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, ErrKeyNotFound, err)

		_, _, err = client.Storage().CreateVolume(context.Background(), &godo.VolumeCreateRequest{})
		assert.Error(t, err)
		assert.Equal(t, ErrVolumeNotFound, err)
	})

	t.Run("Retry Behavior", func(t *testing.T) {
		client := NewMockDOClient()
		client.SimulateDelayedSuccess(2)

		// Primera llamada debería fallar
		_, _, err := client.Droplets().Create(context.Background(), &godo.DropletCreateRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated retry 1/2")

		// Segunda llamada debería fallar
		_, _, err = client.Droplets().Create(context.Background(), &godo.DropletCreateRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated retry 2/2")

		// Tercera llamada debería tener éxito
		droplet, _, err := client.Droplets().Create(context.Background(), &godo.DropletCreateRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, droplet)
	})
}
