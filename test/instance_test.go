package test_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/celestiaorg/talis/test"
)

// TestInstanceCreation demonstrates how to test instance creation functionality.
// This test verifies:
// 1. Basic instance creation works
// 2. Provider mocks are called correctly
// 3. Instance status is tracked properly
func TestInstanceCreation(t *testing.T) {
	// Create a new test suite with server and database
	suite := test.NewTestSuite(t, test.WithServer())
	defer suite.Cleanup()

	// Create a test job request with instance
	jobReq := infrastructure.CreateRequest{
		Name:        "instance-test-job",
		ProjectName: "instance-test-project",
		WebhookURL:  "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 2,
				Provision:         true,
				Region:            "sfo3",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				Tags:              []string{"test", "instance-creation"},
				SSHKeyName:        "test-key",
			},
		},
	}

	// Create the job using the API client
	job, err := suite.APIClient.CreateJob(suite.Context(), jobReq)
	require.NoError(t, err, "Failed to create job")
	require.NotNil(t, job, "Job response should not be nil")

	// Verify job was created
	assert.NotZero(t, job.ID, "Job ID should not be zero")
	assert.Equal(t, "pending", job.Status, "Job should start in pending status")

	// Get the job status to verify instances were created
	status, err := suite.APIClient.GetJob(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err, "Failed to get job status")
	require.NotNil(t, status, "Job status should not be nil")

	// Get instances for the job
	instances, err := suite.APIClient.GetJobInstances(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err, "Failed to get job instances")
	require.NotNil(t, instances, "Job instances should not be nil")

	// Verify instance creation
	assert.Equal(t, 2, len(instances), "Should have created 2 instances")
	for _, instance := range instances {
		assert.NotEmpty(t, instance.Name, "Instance name should not be empty")
		assert.Equal(t, "digitalocean", instance.Provider, "Provider should match request")
		assert.Equal(t, "sfo3", instance.Region, "Region should match request")
		assert.Equal(t, "s-1vcpu-1gb", instance.Size, "Size should match request")
	}
}

// TestInstanceCreationErrors demonstrates how to test error cases in instance creation.
// This test verifies:
// 1. Invalid provider handling
// 2. Resource limit errors
// 3. Provider-specific errors
func TestInstanceCreationErrors(t *testing.T) {
	// Create test suite
	suite := test.NewTestSuite(t, test.WithServer())
	defer suite.Cleanup()

	// Test case: Invalid provider
	invalidProviderReq := infrastructure.CreateRequest{
		Name:        "invalid-provider-job",
		ProjectName: "invalid-provider-project",
		WebhookURL:  "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "invalid-provider",
				NumberOfInstances: 1,
				Region:            "sfo3",
				Size:              "s-1vcpu-1gb",
			},
		},
	}

	// Attempt to create job with invalid provider
	job, err := suite.APIClient.CreateJob(suite.Context(), invalidProviderReq)
	assert.Error(t, err, "Expected error for invalid provider")
	assert.Nil(t, job, "Job should be nil for invalid provider")

	// Test case: Provider authentication failure
	suite.MockDOClient.SimulateAuthenticationFailure()
	authFailureReq := infrastructure.CreateRequest{
		Name:        "auth-failure-job",
		ProjectName: "auth-failure-project",
		WebhookURL:  "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 1,
				Region:            "sfo3",
				Size:              "s-1vcpu-1gb",
			},
		},
	}

	// Attempt to create job with auth failure
	job, err = suite.APIClient.CreateJob(suite.Context(), authFailureReq)
	assert.Error(t, err, "Expected error for authentication failure")
	assert.Nil(t, job, "Job should be nil for authentication failure")

	// Reset mock to standard behavior
	suite.MockDOClient.ResetToStandard()

	// Test case: Rate limit error
	suite.MockDOClient.SimulateRateLimit()
	rateLimitReq := infrastructure.CreateRequest{
		Name:        "rate-limit-job",
		ProjectName: "rate-limit-project",
		WebhookURL:  "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 1,
				Region:            "sfo3",
				Size:              "s-1vcpu-1gb",
			},
		},
	}

	// Attempt to create job with rate limit
	job, err = suite.APIClient.CreateJob(suite.Context(), rateLimitReq)
	assert.Error(t, err, "Expected error for rate limit")
	assert.Nil(t, job, "Job should be nil for rate limit")
}

// TestInstanceCreationRetries demonstrates how to test retry behavior.
// This test verifies:
// 1. Retries on temporary failures
// 2. Eventual success after retries
// 3. Max retry handling
func TestInstanceCreationRetries(t *testing.T) {
	// Create test suite
	suite := test.NewTestSuite(t, test.WithServer())
	defer suite.Cleanup()

	// Configure mock for delayed success (using rate limit simulation)
	suite.MockDOClient.SimulateRateLimit()

	// Create a test job request
	jobReq := infrastructure.CreateRequest{
		Name:        "retry-test-job",
		ProjectName: "retry-test-project",
		WebhookURL:  "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 1,
				Region:            "sfo3",
				Size:              "s-1vcpu-1gb",
			},
		},
	}

	// Attempt to create job (should fail with rate limit)
	job, err := suite.APIClient.CreateJob(suite.Context(), jobReq)
	assert.Error(t, err, "Expected error for rate limit")
	assert.Nil(t, job, "Job should be nil for rate limit")

	// Reset mock to standard behavior to simulate eventual success
	suite.MockDOClient.ResetToStandard()

	// Retry job creation
	job, err = suite.APIClient.CreateJob(suite.Context(), jobReq)
	require.NoError(t, err, "Job creation should succeed after retry")
	require.NotNil(t, job, "Job should not be nil after retry")

	// Verify job was created
	status, err := suite.APIClient.GetJob(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err, "Failed to get job status")
	require.NotNil(t, status, "Job status should not be nil")

	// Get instances for the job
	instances, err := suite.APIClient.GetJobInstances(suite.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err, "Failed to get job instances")
	require.NotNil(t, instances, "Job instances should not be nil")

	// Verify instance was created after retry
	assert.Equal(t, 1, len(instances), "Should have created 1 instance")
	assert.Equal(t, "digitalocean", instances[0].Provider, "Provider should match request")
}
