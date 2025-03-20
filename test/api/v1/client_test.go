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
	Name: "test-job",
}

var defaultInstancesRequest = infrastructure.InstancesRequest{
	JobName:     "test-job",
	ProjectName: "test-project",
	Instances: []infrastructure.InstanceRequest{
		defaultInstanceRequest,
	},
}

var defaultInstanceRequest = infrastructure.InstanceRequest{
	Provider:          models.ProviderID("digitalocean-mock"),
	NumberOfInstances: 1,
	SSHKeyName:        "test-key",
	Region:            "nyc1",
	Size:              "s-1vcpu-1gb",
	Image:             "ubuntu-20-04-x64",
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

	// List instances and verify there is one
	err = suite.Retry(func() error {
		instanceList, err := suite.APIClient.AdminGetInstances(suite.Context())
		if err != nil {
			return err
		}
		if len(instanceList.Instances) != 1 {
			return fmt.Errorf("expected 1 instance, got %d", len(instanceList.Instances))
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// List instances metadata and verify there are none
	instanceMetadata, err := suite.APIClient.AdminGetInstancesMetadata(suite.Context())
	require.NoError(t, err)
	require.NotEmpty(t, instanceMetadata.Instances)
	require.Equal(t, 1, len(instanceMetadata.Instances))
	require.Equal(t, defaultInstanceRequest.Provider, instanceMetadata.Instances[0].ProviderID)
	require.Equal(t, defaultInstanceRequest.Region, instanceMetadata.Instances[0].Region)
	require.Equal(t, defaultInstanceRequest.Size, instanceMetadata.Instances[0].Size)
	require.Equal(t, defaultInstanceRequest.Image, instanceMetadata.Instances[0].Image)
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

	// List instances and verify there are none
	instanceList, err := suite.APIClient.GetInstances(suite.Context())
	require.NoError(t, err)
	require.Empty(t, instanceList.Instances)

	// Create a job to create an instance against
	jobRequest := defaultJobRequest
	err = suite.APIClient.CreateJob(suite.Context(), jobRequest)
	require.NoError(t, err)

	// Create an instance
	err = suite.APIClient.CreateInstance(suite.Context(), defaultInstancesRequest)
	require.NoError(t, err)

	// Wait for the instance to be available
	err = suite.Retry(func() error {
		instanceList, err := suite.APIClient.GetInstances(suite.Context())
		if err != nil {
			return err
		}
		if len(instanceList.Instances) != 1 {
			return fmt.Errorf("expected 1 instance, got %d", len(instanceList.Instances))
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)

	// Grab the instance from the list of instances
	instanceRequest := defaultInstancesRequest.Instances[0]
	instanceList, err = suite.APIClient.GetInstances(suite.Context())
	require.NoError(t, err)
	require.NotEmpty(t, instanceList.Instances)
	require.Equal(t, 1, len(instanceList.Instances))
	require.Equal(t, instanceRequest.Name, instanceList.Instances[0].Name)
	require.Equal(t, instanceRequest.Provider, instanceList.Instances[0].ProviderID)
	require.Equal(t, instanceRequest.Region, instanceList.Instances[0].Region)
	require.Equal(t, instanceRequest.Size, instanceList.Instances[0].Size)
	require.Equal(t, instanceRequest.Image, instanceList.Instances[0].Image)

	actualInstance := instanceList.Instances[0]

	// Get instance metadata
	instanceMetadata, err := suite.APIClient.GetInstancesMetadata(suite.Context())
	require.NoError(t, err)
	require.NotEmpty(t, instanceMetadata.Instances)
	require.Equal(t, 1, len(instanceMetadata.Instances))
	require.Equal(t, actualInstance, instanceMetadata.Instances[0])

	// Get public IPs
	// TODO: this testing currently isn't create because there isn't a great way to link it to the instance. The return type is more geared towards the job.
	publicIPs, err := suite.APIClient.GetInstancesPublicIPs(suite.Context())
	require.NoError(t, err)
	require.NotEmpty(t, publicIPs.PublicIPs)
	require.Equal(t, 1, len(publicIPs.PublicIPs))

	// Delete the instance
	err = suite.APIClient.DeleteInstance(suite.Context(), fmt.Sprint(actualInstance.ID))
	require.NoError(t, err)

	// Verify the the instance eventually gets deleted
	err = suite.Retry(func() error {
		instanceList, err := suite.APIClient.GetInstances(suite.Context())
		if err != nil {
			return err
		}
		if len(instanceList.Instances) > 0 {
			return fmt.Errorf("instance not deleted: %v", instanceList.Instances)
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)
}

func TestClientJobMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// Get jobs to verify there are none
	jobsList, err := suite.APIClient.GetJobs(suite.Context(), handlers.DefaultPageSize, "")
	require.NoError(t, err)
	require.Empty(t, jobsList.Jobs)

	// Create a job
	jobRequest := defaultJobRequest
	err = suite.APIClient.CreateJob(suite.Context(), jobRequest)
	require.NoError(t, err)

	// Wait for the job to be available
	err = suite.Retry(func() error {
		jobsList, err := suite.APIClient.GetJobs(suite.Context(), handlers.DefaultPageSize, "")
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
	jobList, err := suite.APIClient.GetJobs(suite.Context(), handlers.DefaultPageSize, "")
	require.NoError(t, err)
	require.NotEmpty(t, jobList.Jobs)
	require.Equal(t, 1, len(jobList.Jobs))
	require.Equal(t, jobRequest.Name, jobList.Jobs[0].Name)

	actualJob := jobList.Jobs[0]

	// Get the job
	job, err := suite.APIClient.GetJob(suite.Context(), fmt.Sprint(actualJob.ID))
	require.NoError(t, err)
	require.NotNil(t, job)

	// Get instance metadata for the job
	instanceMetadata, err := suite.APIClient.GetMetadataByJobID(suite.Context(), fmt.Sprint(actualJob.ID))
	require.NoError(t, err)
	require.NotNil(t, instanceMetadata)
	require.Equal(t, 1, len(instanceMetadata.Instances))

	// Get instances for the job
	instances, err := suite.APIClient.GetInstancesByJobID(suite.Context(), fmt.Sprint(actualJob.ID))
	require.NoError(t, err)
	require.NotNil(t, instances)
	require.Equal(t, 1, len(instances.Instances))

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
		jobsList, err := suite.APIClient.GetJobs(suite.Context(), handlers.DefaultPageSize, "")
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
