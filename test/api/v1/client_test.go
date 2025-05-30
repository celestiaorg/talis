// Package api_test provides integration tests for the Talis API client.
//
// Unlike the unit tests in the client package, these tests interact with a real
// (or realistically mocked) API server to verify the client's functionality in
// an end-to-end manner. The tests cover:
//
// - Admin operations (listing instances, metadata)
// - Health check functionality
// - Instance operations (creating, listing, deleting)
// - Task operations (listing tasks by instance ID)
//
// These tests use the test.Suite helper to set up a test environment with
// a running API server and database, allowing for comprehensive testing of
// the client's interaction with the actual API.
package api_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/test"
)

func init() {
	// Set environment variable for SSH key name for all tests in this package
	err := os.Setenv(constants.EnvTalisSSHKeyName, "test-key")
	if err != nil {
		// Since we can't use testing.T in init, we'll log the error and continue
		fmt.Printf("Warning: Failed to set %s: %v\n", constants.EnvTalisSSHKeyName, err)
	}
}

// Test data for instance creation requests
var defaultInstancesRequest = []types.InstanceRequest{
	defaultInstanceRequest1,
	defaultInstanceRequest2,
}

// defaultInstanceRequest1 defines the first test instance configuration
var defaultInstanceRequest1 = types.InstanceRequest{
	Provider:          models.ProviderID("digitalocean-mock"),
	OwnerID:           models.AdminID,
	NumberOfInstances: 1,
	ProjectName:       "test-project",
	Region:            "nyc1",
	Size:              "s-1vcpu-1gb",
	Image:             "ubuntu-20-04-x64",
	Provision:         false, // Explicitly disable provisioning for tests
	Volumes: []types.VolumeConfig{
		{
			Name:       "test-volume",
			SizeGB:     10,
			MountPoint: "/mnt/data",
		},
	},
}

// defaultInstanceRequest2 defines the second test instance configuration
var defaultInstanceRequest2 = types.InstanceRequest{
	Provider:          models.ProviderID("digitalocean-mock"),
	OwnerID:           models.AdminID,
	NumberOfInstances: 1,
	ProjectName:       "test-project",
	Region:            "nyc1",
	Size:              "s-1vcpu-1gb",
	Image:             "ubuntu-20-04-x64",
	Provision:         false, // Explicitly disable provisioning for tests
	Volumes: []types.VolumeConfig{
		{
			Name:       "test-volume",
			SizeGB:     10,
			MountPoint: "/mnt/data",
		},
	},
}

// defaultUser1 defines test user parameters with complete information
var defaultUser1 = handlers.CreateUserParams{
	Username: "user1",
	Email:    "user1@example.com",
	Role:     1,
}

// defaultUser2 defines minimal test user parameters
var defaultUser2 = handlers.CreateUserParams{
	Username: "user12",
}

// Create a project request for testing
var defaultProjectParams = handlers.ProjectCreateParams{
	Name:        "test-project",
	Description: "Test project for instances",
	OwnerID:     models.AdminID,
}

// This file contains the comprehensive test suite for the API client.

// TestClientAdminMethods tests the admin methods of the API client.
//
// TODO: once ownerID is implemented, we should test the admin methods with a specific ownerID and that it can see instances and projects across different ownerIDs.
func TestClientAdminMethods(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// Create a project
	_, err := suite.APIClient.CreateProject(suite.Context(), defaultProjectParams)
	require.NoError(t, err)

	// Create an instance
	instancesRequest := defaultInstancesRequest
	instancesRequest[0].ProjectName = defaultProjectParams.Name
	instancesRequest[1].ProjectName = defaultProjectParams.Name
	createdInstances, err := suite.APIClient.CreateInstance(suite.Context(), instancesRequest)
	require.NoError(t, err)
	require.NotNil(t, createdInstances)
	require.Len(t, createdInstances, 2, "Expected 2 instances to be created")

	// List instances and verify there are two (using include_deleted to ensure we see all instances)
	err = suite.Retry(func() error {
		instanceList, err := suite.APIClient.AdminGetInstances(suite.Context())
		if err != nil {
			return err
		}
		if len(instanceList) != 2 {
			return fmt.Errorf("expected 2 instances, got %d", len(instanceList))
		}
		// Verify both instances are in non-terminated state
		for _, instance := range instanceList {
			if instance.Status == models.InstanceStatusTerminated {
				return fmt.Errorf("expected instance ID %d to be non-terminated, got %s", instance.ID, instance.Status)
			}
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// List instances metadata and verify there are two
	instances, err := suite.APIClient.AdminGetInstancesMetadata(suite.Context())
	require.NoError(t, err)
	require.NotEmpty(t, instances)
	require.Equal(t, 2, len(instances))
	require.Equal(t, defaultInstanceRequest1.Provider, instances[0].ProviderID)
	require.Equal(t, defaultInstanceRequest1.Region, instances[0].Region)
	require.Equal(t, defaultInstanceRequest1.Size, instances[0].Size)
	require.Equal(t, defaultInstanceRequest2.Provider, instances[1].ProviderID)
	require.Equal(t, defaultInstanceRequest2.Region, instances[1].Region)
	require.Equal(t, defaultInstanceRequest2.Size, instances[1].Size)
}

// TestClientHealthCheck verifies the HealthCheck method of the API client.
// It ensures that the client can successfully connect to the API and receive
// a valid health check response.
func TestClientHealthCheck(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// Get the health check
	healthCheck, err := suite.APIClient.HealthCheck(suite.Context())
	require.NoError(t, err)
	require.NotNil(t, healthCheck)
}

// TestClientInstanceMethods tests the instance-related methods of the API client.
// This test verifies the full lifecycle of instances:
// 1. Creating instances in a project
// 2. Listing instances and verifying their properties
// 3. Getting instance metadata
// 4. Retrieving public IP addresses
// 5. Deleting instances and verifying they are terminated
// 6. Verifying that terminated instances don't appear in default listings
func TestClientInstanceMethods(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// List instances and verify there are none (using include_deleted to ensure we see all instances)
	instanceList, err := suite.APIClient.GetInstances(suite.Context(), &models.ListOptions{IncludeDeleted: true})
	require.NoError(t, err)
	require.Empty(t, instanceList)

	// Create a project for the instances
	_, err = suite.APIClient.CreateProject(suite.Context(), defaultProjectParams)
	require.NoError(t, err)

	// Create 2 instances
	instancesRequest := defaultInstancesRequest
	instancesRequest[0].ProjectName = defaultProjectParams.Name
	instancesRequest[1].ProjectName = defaultProjectParams.Name
	createdInstances, err := suite.APIClient.CreateInstance(suite.Context(), instancesRequest)
	require.NoError(t, err)
	require.NotNil(t, createdInstances)
	require.Len(t, createdInstances, 2, "Expected 2 instances to be created")

	// Wait for the instances to be available
	err = suite.Retry(func() error {
		instanceList, err := suite.APIClient.GetInstances(suite.Context(), &models.ListOptions{IncludeDeleted: true})
		if err != nil {
			return err
		}
		if len(instanceList) != 2 {
			return fmt.Errorf("expected 2 instances, got %d", len(instanceList))
		}
		// Verify both instances are in non-terminated state
		for _, instance := range instanceList {
			if instance.Status != models.InstanceStatusReady {
				return fmt.Errorf("expected instance ID %d to be ready, got %s", instance.ID, instance.Status)
			}
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// Grab the instance from the list of instances
	instanceList, err = suite.APIClient.GetInstances(suite.Context(), &models.ListOptions{IncludeDeleted: true})
	require.NoError(t, err)
	require.NotEmpty(t, instanceList)
	require.Equal(t, 2, len(instanceList))
	actualInstances := instanceList

	// Get instance metadata
	instances, err := suite.APIClient.GetInstancesMetadata(suite.Context(), &models.ListOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, instances)
	require.Equal(t, 2, len(instances))
	require.Equal(t, actualInstances, instances)

	// Get public IPs
	publicIPs, err := suite.APIClient.GetInstancesPublicIPs(suite.Context(), &models.ListOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, publicIPs.PublicIPs)
	require.Equal(t, 2, len(publicIPs.PublicIPs))

	// Delete both instances
	deleteRequest := types.DeleteInstancesRequest{
		ProjectName: defaultProjectParams.Name,
		InstanceIDs: []uint{createdInstances[0].ID, createdInstances[1].ID},
		OwnerID:     models.AdminID,
	}
	err = suite.APIClient.DeleteInstances(suite.Context(), deleteRequest)
	require.NoError(t, err)

	// Verify the instances eventually get terminated
	err = suite.Retry(func() error {
		// Use status filter to specifically look for terminated instances
		terminatedStatus := models.InstanceStatusTerminated
		instanceList, err := suite.APIClient.GetInstances(suite.Context(), &models.ListOptions{
			IncludeDeleted: true,
			InstanceStatus: &terminatedStatus,
		})
		if err != nil {
			return err
		}
		if len(instanceList) != 2 {
			return fmt.Errorf("expected 2 terminated instances, got %d", len(instanceList))
		}
		// Verify both instances are in terminated state
		for _, instance := range instanceList {
			if instance.Status != models.InstanceStatusTerminated {
				return fmt.Errorf("expected instance ID %d to be terminated, got %s", instance.ID, instance.Status)
			}
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// Submit the same deletion request again - should be a no-op
	// We do this after verifying termination to ensure the first deletion completed
	err = suite.APIClient.DeleteInstances(suite.Context(), deleteRequest)
	require.NoError(t, err)

	// Add a small delay to avoid database lock issues
	time.Sleep(500 * time.Millisecond)

	// Verify that the default list (non-terminated) shows no instances
	instanceList, err = suite.APIClient.GetInstances(suite.Context(), &models.ListOptions{})
	require.NoError(t, err)
	require.Empty(t, instanceList, "expected no non-terminated instances")
}

// TestClient_ListTasksByInstanceID tests the ListTasksByInstanceID method of the API client.
// This comprehensive test verifies:
// 1. Creating instances and their associated tasks
// 2. Retrieving all tasks for a specific instance
// 3. Filtering tasks by action type (create_instances, terminate_instances)
// 4. Pagination of task results
// 5. Handling of non-existent instance IDs
// 6. Proper ordering of tasks by creation time
func TestClient_ListTasksByInstanceID(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// 1. Create a project
	project, err := suite.APIClient.CreateProject(suite.Context(), defaultProjectParams)
	require.NoError(t, err)

	// 2. Create two instances
	instanceReqs := []types.InstanceRequest{
		{
			Provider: models.ProviderID("digitalocean-mock"), OwnerID: models.AdminID, NumberOfInstances: 1,
			ProjectName: project.Name, Region: "sfo3", Size: "s-1vcpu-1gb", Image: "ubuntu-22-04-x64",
			Provision: false, // Explicitly disable provisioning for tests
			Volumes: []types.VolumeConfig{{
				Name: "vol-inst1", SizeGB: 5, MountPoint: "/mnt/data", Region: "sfo3", FileSystem: "ext4",
			}},
		},
		{
			Provider: models.ProviderID("digitalocean-mock"), OwnerID: models.AdminID, NumberOfInstances: 1,
			ProjectName: project.Name, Region: "ams3", Size: "s-2vcpu-2gb", Image: "fedora-38-x64",
			Provision: false, // Explicitly disable provisioning for tests
			Volumes: []types.VolumeConfig{{
				Name: "vol-inst2", SizeGB: 8, MountPoint: "/mnt/data", Region: "ams3", FileSystem: "xfs",
			}},
		},
	}
	createdInstances, err := suite.APIClient.CreateInstance(suite.Context(), instanceReqs)
	require.NoError(t, err)
	require.Len(t, createdInstances, 2)
	instance1 := createdInstances[0]
	instance2 := createdInstances[1]

	// Wait for instances to be ready (and thus have associated create_instance tasks)
	// This also ensures the create_instance tasks associated by the service are in the DB.
	err = suite.Retry(func() error {
		for _, inst := range createdInstances {
			dbInst, err := suite.InstanceRepo.Get(suite.Context(), models.AdminID, inst.ID)
			if err != nil {
				return fmt.Errorf("failed to get instance %d: %w", inst.ID, err)
			}
			if dbInst.Status != models.InstanceStatusReady {
				return fmt.Errorf("instance %d not ready, status: %s", inst.ID, dbInst.Status)
			}
		}
		return nil
	}, 60, 500*time.Millisecond) // Increased retries/timeout for instance readiness
	require.NoError(t, err, "Instances did not become ready in time")

	// 3. Create additional tasks directly for precise control
	taskTerminateInst1 := models.Task{OwnerID: models.AdminID, ProjectID: project.ID, InstanceID: instance1.ID, Action: models.TaskActionTerminateInstances, Status: models.TaskStatusPending, CreatedAt: time.Now().Add(-2 * time.Minute)}
	taskCreate2Inst1 := models.Task{OwnerID: models.AdminID, ProjectID: project.ID, InstanceID: instance1.ID, Action: models.TaskActionCreateInstances, Status: models.TaskStatusFailed, CreatedAt: time.Now().Add(-1 * time.Minute)}
	taskCreateInst2 := models.Task{OwnerID: models.AdminID, ProjectID: project.ID, InstanceID: instance2.ID, Action: models.TaskActionCreateInstances, Status: models.TaskStatusPending, CreatedAt: time.Now()}

	// The create_instance tasks are already created by the service. Let's find them to assert against them.
	// We will also add our manually defined tasks.
	initialTasksInst1, err := suite.APIClient.ListTasksByInstanceID(suite.Context(), models.AdminID, instance1.ID, string(models.TaskActionCreateInstances), nil)
	require.NoError(t, err)
	require.NotEmpty(t, initialTasksInst1, "Instance 1 should have at least one create_instances task from service")

	additionalTasks := []*models.Task{&taskTerminateInst1, &taskCreate2Inst1, &taskCreateInst2}
	// Note: taskCreate1Inst1 is not added here because there should already be a create task from service for instance1
	// If we want it distinct, we should give it a different creation time than the service-created one.
	// For simplicity, we rely on service creating one, and we add a terminate and another create for inst1.

	for _, task := range additionalTasks {
		err = suite.TaskRepo.Create(suite.Context(), task)
		require.NoError(t, err, "Failed to create task %v", task)
	}

	// 4. Test ListTasksByInstanceID for instance1
	// --- Case 1: All tasks for instance1 (should be initial create + terminate + create2) ---
	allTasksInst1, err := suite.APIClient.ListTasksByInstanceID(suite.Context(), models.AdminID, instance1.ID, "", nil)
	require.NoError(t, err)
	// Expecting 3 tasks for instance1: one from service CreateInstance, plus taskTerminateInst1, taskCreate2Inst1.
	require.Len(t, allTasksInst1, 3, "Expected 3 tasks for instance1")
	// Verify order (CreatedAt DESC)
	assert.True(t, allTasksInst1[0].CreatedAt.After(allTasksInst1[1].CreatedAt) || allTasksInst1[0].CreatedAt.Equal(allTasksInst1[1].CreatedAt))
	assert.True(t, allTasksInst1[1].CreatedAt.After(allTasksInst1[2].CreatedAt) || allTasksInst1[1].CreatedAt.Equal(allTasksInst1[2].CreatedAt))

	// --- Case 2: Filter by action "create_instances" for instance1 ---
	createTasksInst1, err := suite.APIClient.ListTasksByInstanceID(suite.Context(), models.AdminID, instance1.ID, string(models.TaskActionCreateInstances), nil)
	require.NoError(t, err)
	require.Len(t, createTasksInst1, 2, "Expected 2 create_instances tasks for instance1")
	for _, task := range createTasksInst1 {
		assert.Equal(t, models.TaskActionCreateInstances, task.Action)
		assert.Equal(t, instance1.ID, task.InstanceID)
	}

	// --- Case 3: Filter by action "terminate_instances" for instance1 ---
	terminateTasksInst1, err := suite.APIClient.ListTasksByInstanceID(suite.Context(), models.AdminID, instance1.ID, string(models.TaskActionTerminateInstances), nil)
	require.NoError(t, err)
	require.Len(t, terminateTasksInst1, 1, "Expected 1 terminate_instances task for instance1")
	assert.Equal(t, models.TaskActionTerminateInstances, terminateTasksInst1[0].Action)
	assert.Equal(t, instance1.ID, terminateTasksInst1[0].InstanceID)
	assert.Equal(t, taskTerminateInst1.ID, terminateTasksInst1[0].ID) // Check it's the specific one we added

	// --- Case 4: Pagination for instance1 (Limit 1, Offset 1) ---
	pagedTasksInst1, err := suite.APIClient.ListTasksByInstanceID(suite.Context(), models.AdminID, instance1.ID, "", &models.ListOptions{Limit: 1, Offset: 1})
	require.NoError(t, err)
	require.Len(t, pagedTasksInst1, 1, "Expected 1 task with limit 1, offset 1 for instance1")
	// Based on CreatedAt DESC order of 3 tasks for instance1, the task at offset 1 (i.e., the second task) should be allTasksInst1[1].
	// Ensure allTasksInst1 has enough elements before accessing index 1.
	require.GreaterOrEqual(t, len(allTasksInst1), 2, "Not enough tasks in allTasksInst1 to check pagination order")
	assert.Equal(t, allTasksInst1[1].ID, pagedTasksInst1[0].ID, "Paginated task ID does not match the expected second task from the full list for instance1")

	// 5. Test ListTasksByInstanceID for a non-existent instance ID
	nonExistentTasks, err := suite.APIClient.ListTasksByInstanceID(suite.Context(), models.AdminID, 99999, "", nil)
	require.NoError(t, err)
	require.Empty(t, nonExistentTasks, "Expected no tasks for a non-existent instance ID")

	// 6. Test ListTasksByInstanceID for instance2 (should have one create_instances task from service + taskCreateInst2)
	tasksInst2, err := suite.APIClient.ListTasksByInstanceID(suite.Context(), models.AdminID, instance2.ID, "", nil)
	require.NoError(t, err)
	// Expecting 2 tasks: one from service CreateInstance for instance2, and taskCreateInst2.
	require.Len(t, tasksInst2, 2, "Expected 2 tasks for instance2")
	foundTaskCreateInst2 := false
	for _, task := range tasksInst2 {
		assert.Equal(t, instance2.ID, task.InstanceID)
		if task.ID == taskCreateInst2.ID {
			foundTaskCreateInst2 = true
		}
	}
	assert.True(t, foundTaskCreateInst2, "Manually added taskCreateInst2 not found for instance2")
}
