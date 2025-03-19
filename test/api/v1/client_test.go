package api_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/api/v1/handlers"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/celestiaorg/talis/test"
)

var defaultCreateRequest = infrastructure.CreateJobRequest{
	JobName:      "test-job",
	InstanceName: "test-instance",
	ProjectName:  "test-project",
	WebhookURL:   "https://example.com/webhook",
	Instances: []infrastructure.InstanceRequest{
		defaultInstance,
	},
}

var defaultInstance = infrastructure.InstanceRequest{
	Provider:          "digitalocean",
	NumberOfInstances: 1,
	SSHKeyName:        "test-key",
	Region:            "nyc1",
	Size:              "s-1vcpu-1gb",
	Image:             "ubuntu-20-04-x64",
}

// This file contains the comprehensive test suite for the API client.

func TestClientJobMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// List jobs and verify there are none
	jobsList, err := suite.APIClient.ListJobs(suite.Context(), handlers.DefaultPageSize, "")
	require.NoError(t, err)
	require.Empty(t, jobsList.Jobs)

	// Create a job
	job, err := suite.APIClient.CreateJob(suite.Context(), defaultCreateRequest)
	require.NoError(t, err)
	require.NotNil(t, job)

	// Compare fields of job to the defaultCreateRequest
	require.Equal(t, defaultCreateRequest.JobName, job.Name)
	require.Equal(t, defaultCreateRequest.InstanceName, job.InstanceName)
	require.Equal(t, defaultCreateRequest.ProjectName, job.ProjectName)
	require.Equal(t, defaultCreateRequest.WebhookURL, job.WebhookURL)

	// Get the job status
	jobStatus, err := suite.APIClient.GetJob(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err)
	require.NotNil(t, jobStatus)

	// List jobs and verify there is one
	jobsList, err = suite.APIClient.ListJobs(suite.Context(), handlers.DefaultPageSize, "")
	require.NoError(t, err)
	require.NotEmpty(t, jobsList.Jobs)
	require.Equal(t, 1, len(jobsList.Jobs))
	require.Equal(t, job.ID, jobsList.Jobs[0].ID)

	// Delete the job
	// TODO: Implement this
}

func TestClientInstanceMethods(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// Create a job to create an instance against
	job, err := suite.APIClient.CreateJob(suite.Context(), defaultCreateRequest)
	require.NoError(t, err)
	require.NotNil(t, job)

	// List instances and verify there are none
	instanceList, err := suite.APIClient.ListInstances(suite.Context())
	require.NoError(t, err)
	require.Empty(t, instanceList.Instances)

	// List job instances and verify there are none
	jobInstances, err := suite.APIClient.GetJobInstances(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err)
	require.Equal(t, 0, len(jobInstances.Instances))

	// Create a job instance
	// The variable is named jobInstance because the underlying type is *models.Job
	jobInstance, err := suite.APIClient.CreateJobInstance(suite.Context(), fmt.Sprint(job.ID), defaultCreateRequest.Instances[0])
	require.NoError(t, err)
	require.NotNil(t, jobInstance)
	require.Equal(t, defaultCreateRequest.InstanceName, jobInstance.InstanceName)

	// Since the return type is *models.Job, we need to find the instance in the list of instances.  wut
	jobInstances, err = suite.APIClient.GetJobInstances(suite.Context(), fmt.Sprint(jobInstance.ID))
	require.NoError(t, err)
	require.NotEmpty(t, jobInstances)
	// Verify the response info
	require.Equal(t, 1, jobInstances.Total)
	require.Equal(t, 1, len(jobInstances.Instances))
	require.Equal(t, job.ID, jobInstances.JobID)
	// Verify the instance info
	expectedInstance := jobInstances.Instances[0]
	require.Equal(t, expectedInstance.Name, defaultCreateRequest.InstanceName)
	require.Equal(t, expectedInstance.ProviderID, defaultInstance.Provider)
	require.Equal(t, expectedInstance.Region, defaultInstance.Region)
	require.Equal(t, expectedInstance.Size, defaultInstance.Size)
	require.Equal(t, expectedInstance.Image, defaultInstance.Image)

	// Verify still only one job
	jobsList, err := suite.APIClient.ListJobs(suite.Context(), handlers.DefaultPageSize, "")
	require.NoError(t, err)
	require.NotEmpty(t, jobsList)
	require.Equal(t, 1, len(jobsList.Jobs))
	require.Equal(t, job.ID, jobsList.Jobs[0].ID)

	// Get the instance
	instance, err := suite.APIClient.GetInstance(suite.Context(), fmt.Sprint(expectedInstance.ID))
	require.NoError(t, err)
	require.NotNil(t, instance)
	require.Equal(t, instance, expectedInstance)

	// List instances and verify there is one
	instanceList, err = suite.APIClient.ListInstances(suite.Context())
	require.NoError(t, err)
	require.NotEmpty(t, instanceList.Instances)
	require.Equal(t, 1, len(instanceList.Instances))
	require.Equal(t, expectedInstance, instanceList.Instances[0])

	// Get the instance metadata
	// Note there is no actual metadata, this is just the instance again currently.
	instanceMetadata, err := suite.APIClient.GetInstanceMetadata(suite.Context())
	require.NoError(t, err)
	require.NotNil(t, instanceMetadata)
	require.Equal(t, 1, len(instanceMetadata.Instances))
	require.Equal(t, expectedInstance, instanceMetadata.Instances[0])

	// Get Public IPs
	publicIPs, err := suite.APIClient.GetJobPublicIPs(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err)
	require.NotEmpty(t, publicIPs)
	require.Equal(t, 1, len(publicIPs.PublicIPs))
	require.Equal(t, expectedInstance.PublicIP, publicIPs.PublicIPs[0].PublicIP)
	require.Equal(t, job.ID, publicIPs.PublicIPs[0].JobID)

	// Delete the Instance
	DeleteJobInstanceRequest := infrastructure.DeleteInstanceRequest{
		ID:           job.ID,
		InstanceName: expectedInstance.Name,
		ProjectName:  job.ProjectName,
		Instances: []infrastructure.InstanceRequest{
			{
				Name: expectedInstance.Name,
			},
		},
	}
	// We don't care about the return value because it is a *models.Job. wut
	_, err = suite.APIClient.DeleteJobInstance(suite.Context(), fmt.Sprint(job.ID), DeleteJobInstanceRequest)
	require.NoError(t, err)

	// Verify the the instance eventually gets deleted
	err = suite.Retry(func() error {
		instanceList, err := suite.APIClient.ListInstances(suite.Context())
		if err != nil {
			return err
		}
		if len(instanceList.Instances) > 0 {
			return fmt.Errorf("instance not deleted")
		}
		return nil
	}, 100, 100*time.Millisecond)
	require.NoError(t, err)
}

func TestClientHealthCheck(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Cleanup()

	// Get the health check
	healthCheck, err := suite.APIClient.HealthCheck(suite.Context())
	require.NoError(t, err)
	require.NotNil(t, healthCheck)
}
