package compute

import (
	"context"
	"strconv"
	"testing"

	"github.com/celestiaorg/talis/internal/config"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestProvider() (*VirtFusionProvider, error) {
	cfg := &config.VirtFusionConfig{
		Host:     "http://mock-virtfusion",
		APIToken: "test-token",
	}
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &VirtFusionProvider{
		client: client,
		config: cfg,
	}, nil
}

func TestVirtFusionProvider_CreateInstance(t *testing.T) {
	provider, err := setupTestProvider()
	require.NoError(t, err)

	ctx := context.Background()
	instanceName := "test-instance"
	config := types.InstanceConfig{
		Size:  "small",
		Image: "ubuntu-22.04",
		Volumes: []types.VolumeConfig{
			{
				Name:       "data",
				SizeGB:     20,
				MountPoint: "/mnt/data",
			},
		},
	}

	instances, err := provider.CreateInstance(ctx, instanceName, config)
	require.NoError(t, err)
	require.Len(t, instances, 1)

	instance := instances[0]
	assert.NotEmpty(t, instance.ID)
	assert.Equal(t, instanceName, instance.Name)
}

func TestVirtFusionProvider_DeleteInstance(t *testing.T) {
	provider, err := setupTestProvider()
	require.NoError(t, err)

	ctx := context.Background()
	instanceName := "test-instance"

	// Create an instance first
	instances, err := provider.CreateInstance(ctx, instanceName, types.InstanceConfig{
		Size:  "small",
		Image: "ubuntu-22.04",
		Volumes: []types.VolumeConfig{
			{
				Name:       "data",
				SizeGB:     20,
				MountPoint: "/mnt/data",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, instances, 1)

	// Delete the instance
	err = provider.DeleteInstance(ctx, instanceName, "")
	require.NoError(t, err)

	// Verify the instance is deleted
	list, err := provider.ListInstances(ctx)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestVirtFusionProvider_GetInstanceStatus(t *testing.T) {
	provider, err := setupTestProvider()
	require.NoError(t, err)

	ctx := context.Background()
	instanceName := "test-instance"

	// Create a test instance
	instances, err := provider.CreateInstance(ctx, instanceName, types.InstanceConfig{
		Size:  "small",
		Image: "ubuntu-22.04",
		Volumes: []types.VolumeConfig{
			{
				Name:       "data",
				SizeGB:     20,
				MountPoint: "/mnt/data",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, instances, 1)

	// Test getting status
	status, err := provider.GetInstanceStatus(ctx, instances[0].ID)
	require.NoError(t, err)
	assert.NotEmpty(t, status)

	// Test non-existent instance
	_, err = provider.GetInstanceStatus(ctx, "999999")
	assert.Error(t, err)
}

func TestVirtFusionProvider_ListInstances(t *testing.T) {
	provider, err := setupTestProvider()
	require.NoError(t, err)

	ctx := context.Background()

	// Create test instances
	instance1, err := provider.CreateInstance(ctx, "test1", types.InstanceConfig{
		Size:  "small",
		Image: "ubuntu-22.04",
		Volumes: []types.VolumeConfig{
			{
				Name:       "data",
				SizeGB:     20,
				MountPoint: "/mnt/data",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, instance1, 1)

	instance2, err := provider.CreateInstance(ctx, "test2", types.InstanceConfig{
		Size:  "small",
		Image: "ubuntu-22.04",
		Volumes: []types.VolumeConfig{
			{
				Name:       "data",
				SizeGB:     20,
				MountPoint: "/mnt/data",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, instance2, 1)

	// Test listing instances
	instances, err := provider.ListInstances(ctx)
	require.NoError(t, err)
	assert.Len(t, instances, 2)

	// Verify instance details
	instanceMap := make(map[string]types.InstanceInfo)
	for _, instance := range instances {
		instanceMap[instance.Name] = instance
	}

	assert.NotEmpty(t, instanceMap["test1"].ID)
	assert.NotEmpty(t, instanceMap["test2"].ID)
}

func TestVirtFusionProvider_RecoverInstance(t *testing.T) {
	provider, err := setupTestProvider()
	require.NoError(t, err)

	ctx := context.Background()
	instanceName := "test-instance"

	// Create a test instance
	instances, err := provider.CreateInstance(ctx, instanceName, types.InstanceConfig{
		Size:  "small",
		Image: "ubuntu-22.04",
		Volumes: []types.VolumeConfig{
			{
				Name:       "data",
				SizeGB:     20,
				MountPoint: "/mnt/data",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, instances, 1)

	instanceID := instances[0].ID

	// Test recovery of a failed instance
	t.Run("Recover failed instance", func(t *testing.T) {
		err := provider.RecoverInstance(ctx, instanceID)
		require.NoError(t, err)

		// Verify instance status after recovery
		status, err := provider.GetInstanceStatus(ctx, instanceID)
		require.NoError(t, err)
		assert.NotEqual(t, "failed", status)
	})

	// Test recovery of a suspended instance
	t.Run("Recover suspended instance", func(t *testing.T) {
		// Get the server service to manipulate instance state
		serverID, err := strconv.Atoi(instanceID)
		require.NoError(t, err)

		// Suspend the instance
		_, err = provider.client.Servers().Suspend(ctx, serverID)
		require.NoError(t, err)

		// Verify instance is suspended
		status, err := provider.GetInstanceStatus(ctx, instanceID)
		require.NoError(t, err)
		assert.Equal(t, "suspended", status)

		// Attempt recovery
		err = provider.RecoverInstance(ctx, instanceID)
		require.NoError(t, err)

		// Verify instance is no longer suspended
		status, err = provider.GetInstanceStatus(ctx, instanceID)
		require.NoError(t, err)
		assert.NotEqual(t, "suspended", status)
	})

	// Test recovery with invalid instance ID
	t.Run("Recover with invalid instance ID", func(t *testing.T) {
		err := provider.RecoverInstance(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid instance ID")
	})

	// Test recovery of non-existent instance
	t.Run("Recover non-existent instance", func(t *testing.T) {
		err := provider.RecoverInstance(ctx, "999999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get server status")
	})
}
