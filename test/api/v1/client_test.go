package api_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/api/v1/handlers"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/celestiaorg/talis/test"
)

var defaultJobRequest = infrastructure.JobRequest{
	Name:    "test-job",
	OwnerID: models.AdminID,
}

var defaultInstancesRequest = infrastructure.InstancesRequest{
	JobName:      "test-job",
	ProjectName:  "test-project",
	InstanceName: "test-instance",
	Instances: []infrastructure.InstanceRequest{
		defaultInstanceRequest1,
		defaultInstanceRequest2,
	},
}

var defaultInstanceRequest1 = infrastructure.InstanceRequest{
	Provider:          models.ProviderID("digitalocean-mock"),
	OwnerID:           models.AdminID,
	NumberOfInstances: 1,
	SSHKeyName:        "test-key",
	Region:            "nyc1",
	Size:              "s-1vcpu-1gb",
	Image:             "ubuntu-20-04-x64",
}

var defaultInstanceRequest2 = infrastructure.InstanceRequest{
	Provider:          models.ProviderID("digitalocean-mock"),
	OwnerID:           models.AdminID,
	NumberOfInstances: 1,
	Name:              "custom-instance",
	SSHKeyName:        "test-key",
	Region:            "nyc1",
	Size:              "s-1vcpu-1gb",
	Image:             "ubuntu-20-04-x64",
}

var defaultUser1 = infrastructure.CreateUserRequest{
	Username: "user1",
	Email:    "user1@example.com",
	Role:     1,
}
var defaultUser2 = infrastructure.CreateUserRequest{
	Username: "user12",
}

// This file contains the comprehensive test suite for the API client.

// TestClientAdminMethods tests the admin methods of the API client.
//
// TODO: once ownerID is implemented, we should test the admin methods with a specific ownerID and that it can see instances and jobs across different ownerIDs.
func TestClientAdminMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// Create a job
	err := suite.APIClient.CreateJob(suite.Context(), defaultJobRequest)
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
		if len(instanceList.Instances) != 2 {
			return fmt.Errorf("expected 2 instances, got %d", len(instanceList.Instances))
		}
		// Verify both instances are in non-terminated state
		for _, instance := range instanceList.Instances {
			if instance.Status == models.InstanceStatusTerminated {
				return fmt.Errorf("expected instance %s to be non-terminated, got %s", instance.Name, instance.Status)
			}
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// List instances metadata and verify there are two
	instanceMetadata, err := suite.APIClient.AdminGetInstancesMetadata(suite.Context())
	require.NoError(t, err)
	require.NotEmpty(t, instanceMetadata.Instances)
	require.Equal(t, 2, len(instanceMetadata.Instances))
	require.Equal(t, defaultInstanceRequest1.Provider, instanceMetadata.Instances[0].ProviderID)
	require.Equal(t, defaultInstanceRequest1.Region, instanceMetadata.Instances[0].Region)
	require.Equal(t, defaultInstanceRequest1.Size, instanceMetadata.Instances[0].Size)
	require.Equal(t, defaultInstanceRequest2.Provider, instanceMetadata.Instances[1].ProviderID)
	require.Equal(t, defaultInstanceRequest2.Region, instanceMetadata.Instances[1].Region)
	require.Equal(t, defaultInstanceRequest2.Size, instanceMetadata.Instances[1].Size)
}

func TestClientHealthCheck(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// Get the health check
	healthCheck, err := suite.APIClient.HealthCheck(suite.Context())
	require.NoError(t, err)
	require.NotNil(t, healthCheck)
}

func TestClientInstanceMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// List instances and verify there are none (using include_deleted to ensure we see all instances)
	instanceList, err := suite.APIClient.GetInstances(suite.Context(), &models.ListOptions{IncludeDeleted: true})
	require.NoError(t, err)
	require.Empty(t, instanceList.Instances)

	// Create a job to create an instance against
	jobRequest := defaultJobRequest
	err = suite.APIClient.CreateJob(suite.Context(), jobRequest)
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
		if len(instanceList.Instances) != 2 {
			return fmt.Errorf("expected 2 instances, got %d", len(instanceList.Instances))
		}
		// Verify both instances are in non-terminated state
		for _, instance := range instanceList.Instances {
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
	require.NotEmpty(t, instanceList.Instances)
	require.Equal(t, 2, len(instanceList.Instances))
	actualInstances := instanceList.Instances

	// Get instance metadata
	instanceMetadata, err := suite.APIClient.GetInstancesMetadata(suite.Context(), &models.ListOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, instanceMetadata.Instances)
	require.Equal(t, 2, len(instanceMetadata.Instances))
	require.Equal(t, actualInstances, instanceMetadata.Instances)

	// Get public IPs
	publicIPs, err := suite.APIClient.GetInstancesPublicIPs(suite.Context(), &models.ListOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, publicIPs.PublicIPs)
	require.Equal(t, 2, len(publicIPs.PublicIPs))

	// Delete both instances
	deleteRequest := infrastructure.DeleteInstanceRequest{
		JobName:       jobRequest.Name,
		InstanceNames: []string{actualInstances[0].Name, actualInstances[1].Name},
	}
	err = suite.APIClient.DeleteInstance(suite.Context(), deleteRequest)
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
		if len(instanceList.Instances) != 2 {
			return fmt.Errorf("expected 2 terminated instances, got %d", len(instanceList.Instances))
		}
		// Verify both instances are in terminated state
		for _, instance := range instanceList.Instances {
			if instance.Status != models.InstanceStatusTerminated {
				return fmt.Errorf("expected instance %s to be terminated, got %s", instance.Name, instance.Status)
			}
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// Submit the same deletion request again - should be a no-op
	// We do this after verifying termination to ensure the first deletion completed
	err = suite.APIClient.DeleteInstance(suite.Context(), deleteRequest)
	require.NoError(t, err)

	// Verify that the default list (non-terminated) shows no instances
	instanceList, err = suite.APIClient.GetInstances(suite.Context(), &models.ListOptions{})
	require.NoError(t, err)
	require.Empty(t, instanceList.Instances, "expected no non-terminated instances")
}

func TestClientJobMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// Get jobs to verify there are none
	jobsList, err := suite.APIClient.GetJobs(suite.Context(), &models.ListOptions{
		Limit: handlers.DefaultPageSize,
	})
	require.NoError(t, err)
	require.Empty(t, jobsList.Jobs)

	// Create a job
	jobRequest := defaultJobRequest
	err = suite.APIClient.CreateJob(suite.Context(), jobRequest)
	require.NoError(t, err)

	// Wait for the job to be available
	err = suite.Retry(func() error {
		jobsList, err := suite.APIClient.GetJobs(suite.Context(), &models.ListOptions{
			Limit: handlers.DefaultPageSize,
		})
		if err != nil {
			return err
		}
		if len(jobsList.Jobs) != 1 {
			return fmt.Errorf("expected 1 job, got %d", len(jobsList.Jobs))
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// Grab the job from the list of jobs
	jobList, err := suite.APIClient.GetJobs(suite.Context(), &models.ListOptions{
		Limit: handlers.DefaultPageSize,
	})
	require.NoError(t, err)
	require.NotEmpty(t, jobList.Jobs)
	require.Equal(t, 1, len(jobList.Jobs))
	require.Equal(t, jobRequest.Name, jobList.Jobs[0].Name)

	actualJob := jobList.Jobs[0]
	// Ignore unused variable warning
	_ = actualJob

	t.Skip("Skipping job ID issue")
	// Get the job
	job, err := suite.APIClient.GetJob(suite.Context(), fmt.Sprint(actualJob.ID))
	require.NoError(t, err)
	require.NotNil(t, job)
	// TODO: the job appears to be mismatched?

	// Get instance metadata for the job
	instanceMetadata, err := suite.APIClient.GetMetadataByJobID(suite.Context(), fmt.Sprint(actualJob.ID), &models.ListOptions{})
	require.NoError(t, err)
	require.NotNil(t, instanceMetadata)
	// TODO Job ID issue causing this as well.
	// require.Equal(t, 1, len(instanceMetadata.Instances))

	// Get instances for the job
	instances, err := suite.APIClient.GetInstancesByJobID(suite.Context(), fmt.Sprint(actualJob.ID), &models.ListOptions{})
	require.NoError(t, err)
	require.NotNil(t, instances)
	// TODO Job ID issue causing this as well.
	// require.Equal(t, 1, len(instances.Instances))

	// Verify returned instances are the same
	require.Equal(t, instanceMetadata.Instances, instances.Instances)

	// Get the job status
	jobStatus, err := suite.APIClient.GetJobStatus(suite.Context(), fmt.Sprint(actualJob.ID))
	require.NoError(t, err)
	require.NotNil(t, jobStatus)

	// Update the job
	// TODO: this is not implemented yet, update when it is
	err = suite.APIClient.UpdateJob(suite.Context(), fmt.Sprint(actualJob.ID), jobRequest)
	require.NoError(t, err)

	// Delete the job
	err = suite.APIClient.DeleteJob(suite.Context(), fmt.Sprint(actualJob.ID))
	require.NoError(t, err)

	// Verify the job is deleted
	err = suite.Retry(func() error {
		jobsList, err := suite.APIClient.GetJobs(suite.Context(), &models.ListOptions{
			Limit: handlers.DefaultPageSize,
		})
		if err != nil {
			return err
		}
		if len(jobsList.Jobs) > 0 {
			return fmt.Errorf("job not deleted: %v", jobsList.Jobs)
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)
}

func TestClientUserMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
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
		newUser, err := suite.APIClient.CreateUser(suite.Context(), infrastructure.CreateUserRequest{
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
		_, err := suite.APIClient.CreateUser(suite.Context(), infrastructure.CreateUserRequest{
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
		user, err := suite.APIClient.CreateUser(suite.Context(), infrastructure.CreateUserRequest{
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
