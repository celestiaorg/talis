package compute

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/celestiaorg/talis/internal/types"
)

func TestMockDOClient_SimulateAuthenticationFailure(t *testing.T) {
	client := NewMockDOClient()
	client.SimulateAuthenticationFailure()

	// Test droplet operations
	_, err := client.CreateInstance(context.Background(), "test", types.InstanceConfig{})
	assert.ErrorIs(t, err, ErrAuthentication)

	// Test key operations
	_, _, err = client.Keys().List(context.Background(), nil)
	assert.ErrorIs(t, err, ErrAuthentication)

	// Test storage operations
	_, _, err = client.Storage().CreateVolume(context.Background(), nil)
	assert.ErrorIs(t, err, ErrAuthentication)
}

func TestMockDOClient_SimulateRateLimit(t *testing.T) {
	client := NewMockDOClient()
	client.SimulateRateLimit()

	// Test droplet operations
	_, err := client.CreateInstance(context.Background(), "test", types.InstanceConfig{})
	assert.ErrorIs(t, err, ErrRateLimit)

	// Test key operations
	_, _, err = client.Keys().List(context.Background(), nil)
	assert.ErrorIs(t, err, ErrRateLimit)

	// Test storage operations
	_, _, err = client.Storage().CreateVolume(context.Background(), nil)
	assert.ErrorIs(t, err, ErrRateLimit)
}

func TestMockDOClient_SimulateNotFound(t *testing.T) {
	client := NewMockDOClient()
	client.SimulateNotFound()

	// Test droplet operations
	_, err := client.CreateInstance(context.Background(), "test", types.InstanceConfig{})
	assert.ErrorIs(t, err, ErrDropletNotFound)

	// Test key operations
	_, _, err = client.Keys().List(context.Background(), nil)
	assert.ErrorIs(t, err, ErrKeyNotFound)

	// Test storage operations
	_, _, err = client.Storage().CreateVolume(context.Background(), nil)
	assert.ErrorIs(t, err, ErrVolumeNotFound)
}

func TestMockDOClient_ResetToStandard(t *testing.T) {
	client := NewMockDOClient()

	// First simulate some errors
	client.SimulateAuthenticationFailure()

	// Then reset
	client.ResetToStandard()

	// Test droplet operations
	_, err := client.CreateInstance(context.Background(), "test", types.InstanceConfig{})
	assert.NoError(t, err)

	// Test key operations
	_, _, err = client.Keys().List(context.Background(), nil)
	assert.NoError(t, err)

	// Test storage operations
	_, _, err = client.Storage().CreateVolume(context.Background(), nil)
	assert.NoError(t, err)
}
