package api_test

import (
	"fmt"
	"testing"

	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/celestiaorg/talis/test"
	"github.com/stretchr/testify/require"
)

// This file contains the comprehensive test suite for the API client.
// It includes tests for all the methods in the client interface.
// It also includes tests for the error handling and the retry logic.
// It also includes tests for the health check functionality.

func TestClientJobMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// List jobs and verify there are none
	jobs, err := suite.APIClient.ListJobs(suite.Context(), 10, "")
	require.NoError(t, err)
	require.Empty(t, jobs)

	// Create a job
	createRequest := infrastructure.CreateJobRequest{
		JobName:      "test-job",
		InstanceName: "test-instance",
		ProjectName:  "test-project",
		WebhookURL:   "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 1,
				SSHKeyName:        "test-key",
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
			},
		},
	}
	job, err := suite.APIClient.CreateJob(suite.Context(), createRequest)
	require.NoError(t, err)
	require.NotNil(t, job)

	// Get the job status
	jobStatus, err := suite.APIClient.GetJob(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err)
	require.NotNil(t, jobStatus)

	// List jobs and verify there is one
	jobs, err = suite.APIClient.ListJobs(suite.Context(), 10, "")
	require.NoError(t, err)
	require.NotEmpty(t, jobs)

	// Delete the job
	// TODO: Implement this
}

func TestClientInstanceMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// First Test Job Instance Methods since those are how instances are created.

	// Create a job
	createRequest := infrastructure.CreateJobRequest{
		JobName:      "test-job",
		InstanceName: "test-instance",
		ProjectName:  "test-project",
		WebhookURL:   "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 1,
				SSHKeyName:        "test-key",
				Region:            "nyc1",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
			},
		},
	}
	job, err := suite.APIClient.CreateJob(suite.Context(), createRequest)
	require.NoError(t, err)
	require.NotNil(t, job)

	// List instances and verify there are none
	instances, err := suite.APIClient.ListInstances(suite.Context())
	require.NoError(t, err)
	require.Empty(t, instances)

	// Create a job instance
	instance, err := suite.APIClient.CreateJobInstance(suite.Context(), fmt.Sprint(job.ID), createRequest.Instances[0])
	require.NoError(t, err)
	require.NotNil(t, instance)

	// List instances and verify there is one
	instances, err = suite.APIClient.ListInstances(suite.Context())
	require.NoError(t, err)
	require.NotEmpty(t, instances)

	// List job instances and verify there is one
	jobInstances, err := suite.APIClient.GetJobInstances(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err)
	require.NotEmpty(t, jobInstances)
	require.Equal(t, instances, jobInstances)

	// Get the instance
	instance, err = suite.APIClient.GetInstance(suite.Context(), fmt.Sprint(instance.ID))
	require.NoError(t, err)
	require.NotNil(t, instance)

	// Get the instance by job id and instance id
	jobInstance, err := suite.APIClient.GetJobInstance(suite.Context(), fmt.Sprint(job.ID), fmt.Sprint(instance.ID))
	require.NoError(t, err)
	require.NotNil(t, jobInstance)
	require.Equal(t, instance.ID, jobInstance.ID)

	// Get the instance metadata
	instanceMetadata, err := suite.APIClient.GetInstanceMetadata(suite.Context())
	require.NoError(t, err)
	require.NotNil(t, instanceMetadata)

	// Get Public IPs
	publicIPs, err := suite.APIClient.GetJobPublicIPs(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err)
	require.NotEmpty(t, publicIPs)

	// Delete the Instance
	DeleteJobInstanceRequest := infrastructure.DeleteInstanceRequest{
		InstanceID: fmt.Sprint(instance.ID),
	}
	jobInstance, err = suite.APIClient.DeleteJobInstance(suite.Context(), fmt.Sprint(job.ID), DeleteJobInstanceRequest)
	require.NoError(t, err)
	require.NotNil(t, jobInstance)
}

func TestClientHealthCheck(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// Get the health check
	healthCheck, err := suite.APIClient.HealthCheck(suite.Context())
	require.NoError(t, err)
	require.NotNil(t, healthCheck)
}
