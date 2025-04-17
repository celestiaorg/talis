package api_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/test"
)

var defaultInstancesRequest = types.InstancesRequest{
	ProjectName:  "test-project",
	TaskName:     "test-project",
	InstanceName: "test-instance",
	Instances: []types.InstanceRequest{
		defaultInstanceRequest1,
		defaultInstanceRequest2,
	},
}

var defaultInstanceRequest1 = types.InstanceRequest{
	Provider:          models.ProviderID("digitalocean-mock"),
	OwnerID:           models.AdminID,
	NumberOfInstances: 1,
	Name:              "test-instance-1",
	SSHKeyName:        "test-key",
	Region:            "nyc1",
	Size:              "s-1vcpu-1gb",
	Image:             "ubuntu-20-04-x64",
	Volumes: []types.VolumeConfig{
		{
			Name:       "test-volume",
			SizeGB:     10,
			MountPoint: "/mnt/data",
		},
	},
}

var defaultInstanceRequest2 = types.InstanceRequest{
	Provider:          models.ProviderID("digitalocean-mock"),
	OwnerID:           models.AdminID,
	NumberOfInstances: 1,
	Name:              "custom-instance",
	SSHKeyName:        "test-key",
	Region:            "nyc1",
	Size:              "s-1vcpu-1gb",
	Image:             "ubuntu-20-04-x64",
	Volumes: []types.VolumeConfig{
		{
			Name:       "test-volume",
			SizeGB:     10,
			MountPoint: "/mnt/data",
		},
	},
}

var defaultUser1 = types.CreateUserRequest{
	Username: "user1",
	Email:    "user1@example.com",
	Role:     1,
}
var defaultUser2 = types.CreateUserRequest{
	Username: "user12",
}

// Create a project request for testing
var defaultProjectParams = handlers.ProjectCreateParams{
	Name:        "test-project",
	Description: "Test project for instances",
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
	err = suite.APIClient.CreateInstance(suite.Context(), defaultInstancesRequest)
	require.NoError(t, err)

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
				return fmt.Errorf("expected instance %s to be non-terminated, got %s", instance.Name, instance.Status)
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

func TestClientHealthCheck(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	// Get the health check
	healthCheck, err := suite.APIClient.HealthCheck(suite.Context())
	require.NoError(t, err)
	require.NotNil(t, healthCheck)
}

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
	err = suite.APIClient.CreateInstance(suite.Context(), defaultInstancesRequest)
	require.NoError(t, err)

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
			if instance.Status == models.InstanceStatusTerminated {
				return fmt.Errorf("expected instance %s to be non-terminated, got %s", instance.Name, instance.Status)
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
	deleteRequest := types.DeleteInstanceRequest{
		ProjectName:   defaultProjectParams.Name,
		InstanceNames: []string{actualInstances[0].Name, actualInstances[1].Name},
	}
	response, err := suite.APIClient.DeleteInstance(suite.Context(), deleteRequest)
	require.NoError(t, err)
	require.NotEmpty(t, response.TaskName)

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
				return fmt.Errorf("expected instance %s to be terminated, got %s", instance.Name, instance.Status)
			}
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// Submit the same deletion request again - should be a no-op
	// We do this after verifying termination to ensure the first deletion completed
	response, err = suite.APIClient.DeleteInstance(suite.Context(), deleteRequest)
	require.NoError(t, err)
	require.NotEmpty(t, response.TaskName)

	// Add a small delay to avoid database lock issues
	time.Sleep(500 * time.Millisecond)

	// Verify that the default list (non-terminated) shows no instances
	instanceList, err = suite.APIClient.GetInstances(suite.Context(), &models.ListOptions{})
	require.NoError(t, err)
	require.Empty(t, instanceList.Instances, "expected no non-terminated instances")
}

func TestClientUserMethods(t *testing.T) {
	suite := test.NewSuite(t)
	defer suite.Cleanup()

	/////////////////////
	// t.Run()
	// Start with 0, then increment this variable whenever a user is successfully created
	// and decrement when a user is successfully deleted
	expectedUserCount := 0
	users, err := suite.APIClient.GetUsers(suite.Context(), nil)
	require.NoError(t, err)
	require.NotNil(t, users)
	require.Empty(t, users.Users, "Expected no users in a fresh database")

	t.Run("CreateUser_Success", func(t *testing.T) {
		// Create first user
		newUser1, err := suite.APIClient.CreateUser(suite.Context(), defaultUser1)
		require.NoError(t, err)
		require.NotEmpty(t, newUser1.UserID, "User ID should not be empty")
		expectedUserCount++

		// Create second user
		newUser2, err := suite.APIClient.CreateUser(suite.Context(), defaultUser2)
		require.NoError(t, err)
		require.NotEmpty(t, newUser2.UserID, "User ID should not be empty")
		expectedUserCount++
	})

	t.Run("CreateUser_DuplicateUsername", func(t *testing.T) {
		// Try to create a user with the same username
		duplicateUser := defaultUser1
		_, err := suite.APIClient.CreateUser(suite.Context(), duplicateUser)
		require.Error(t, err, "Creating user with duplicate username should fail")
	})

	t.Run("GetUserByID_Success", func(t *testing.T) {
		// Create a user first
		newUser, err := suite.APIClient.CreateUser(suite.Context(), types.CreateUserRequest{
			Username:     "testuser_getbyid",
			Email:        "getbyid@example.com",
			Role:         1,
			PublicSSHKey: "ssh-rsa TESTKEY",
		})
		require.NoError(t, err)
		expectedUserCount++

		// Get the user by ID
		resp, err := suite.APIClient.GetUserByID(suite.Context(), fmt.Sprint(newUser.UserID))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "testuser_getbyid", resp.User.Username)
		require.Equal(t, "getbyid@example.com", resp.User.Email)
	})

	t.Run("GetUserByUsername_Success", func(t *testing.T) {
		// Create a user first
		uniqueUsername := "unique_username_test"
		_, err := suite.APIClient.CreateUser(suite.Context(), types.CreateUserRequest{
			Username:     uniqueUsername,
			Email:        "unique@example.com",
			Role:         1,
			PublicSSHKey: "ssh-rsa UNIQUEKEY",
		})
		require.NoError(t, err)
		expectedUserCount++

		// Get the user by username
		userResp, err := suite.APIClient.GetUsers(suite.Context(), &models.UserQueryOptions{Username: uniqueUsername})
		require.NoError(t, err)
		require.NotNil(t, userResp.User)
		require.Equal(t, uniqueUsername, userResp.User.Username)
		require.Equal(t, "unique@example.com", userResp.User.Email)
	})

	t.Run("GetUserByUsername_NotFound", func(t *testing.T) {
		// Try to get a non-existent username
		_, err := suite.APIClient.GetUsers(suite.Context(), &models.UserQueryOptions{Username: "nonexistent_user"})
		require.Error(t, err, "Getting non-existent username should return error")
	})

	t.Run("Get_All_Users", func(t *testing.T) {
		users, err := suite.APIClient.GetUsers(suite.Context(), &models.UserQueryOptions{})
		require.NoError(t, err)
		require.Equal(t, expectedUserCount, len(users.Users))
	})

	t.Run("DeleteUser_Success", func(t *testing.T) {
		deletedUsername := "deleted_username_test"
		user, err := suite.APIClient.CreateUser(suite.Context(), types.CreateUserRequest{
			Username:     deletedUsername,
			Email:        "deleted@example.com",
			Role:         1,
			PublicSSHKey: "ssh-rsa deletedKEY",
		})
		require.NoError(t, err)
		expectedUserCount++

		// Delete a existing user
		err = suite.APIClient.DeleteUser(suite.Context(), fmt.Sprint(user.UserID))
		require.NoError(t, err)
		expectedUserCount--

		// Verify the user is actually deleted
		_, err = suite.APIClient.GetUserByID(suite.Context(), fmt.Sprint(user.UserID))
		require.Error(t, err, "User should no longer exist after deletion")

		// Delete an non existing user
		nonExistingUserID := "234245"
		err = suite.APIClient.DeleteUser(suite.Context(), nonExistingUserID)
		require.Error(t, err)
	})
}
