package test_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/celestiaorg/talis/test"
)

// ExampleTest demonstrates how to use the test environment for integration testing.
// This example shows:
// 1. Setting up a test environment
// 2. Creating a job with an instance
// 3. Verifying the job was created successfully
// 4. Proper cleanup of resources
func TestExampleJobCreation(t *testing.T) {
	// Create a new test environment with server and database
	env := test.NewTestEnvironment(t, test.WithServer())
	defer env.Cleanup() // Always clean up resources

	// Create a test job request
	jobReq := infrastructure.CreateRequest{
		Name:        "test-job",
		ProjectName: "test-project",
		WebhookURL:  "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "digitalocean",
				NumberOfInstances: 1,
				Provision:         true,
				Region:            "sfo3",
				Size:              "s-1vcpu-1gb",
				Image:             "ubuntu-20-04-x64",
				Tags:              []string{"test", "integration"},
				SSHKeyName:        "test-key",
			},
		},
	}

	// Create the job using the API client
	job, err := env.APIClient.CreateJob(env.Context(), jobReq)
	require.NoError(t, err, "Failed to create job")
	require.NotNil(t, job, "Job response should not be nil")

	// Verify job details
	assert.NotZero(t, job.ID, "Job ID should not be zero")
	assert.Equal(t, "pending", job.Status, "Job should start in pending status")

	// Get the job status to verify it was created
	status, err := env.APIClient.GetJob(env.Context(), fmt.Sprint(job.ID))
	require.NoError(t, err, "Failed to get job status")
	require.NotNil(t, status, "Job status should not be nil")

	// Verify job status
	assert.Equal(t, fmt.Sprint(job.ID), status.JobID, "Job ID mismatch")
	assert.NotEmpty(t, status.Status, "Job status should not be empty")
	assert.NotEmpty(t, status.CreatedAt, "Job created at should not be empty")
}

// This example demonstrates how to use timeouts in tests.
// It shows:
// 1. Using the default test timeout
// 2. Creating a custom timeout for specific operations
// 3. Handling context cancellation
func TestExampleWithTimeout(t *testing.T) {
	// Create test environment
	env := test.NewTestEnvironment(t, test.WithServer())
	defer env.Cleanup()

	// Create a job with the default timeout
	jobReq := infrastructure.CreateRequest{
		Name:        "timeout-test-job",
		ProjectName: "timeout-test-project",
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

	// Create the job
	job, err := env.APIClient.CreateJob(env.Context(), jobReq)
	require.NoError(t, err, "Failed to create job")
	require.NotNil(t, job, "Job response should not be nil")

	// Use a custom timeout for getting the job status
	ctx, cancel := env.WithTimeout(test.DefaultTestTimeout / 2)
	defer cancel()

	// Get job status with custom timeout
	status, err := env.APIClient.GetJob(ctx, fmt.Sprint(job.ID))
	require.NoError(t, err, "Failed to get job status")
	require.NotNil(t, status, "Job status should not be nil")
}

// This example demonstrates how to test error cases.
// It shows:
// 1. Testing validation errors
// 2. Testing resource conflicts
// 3. Proper error handling
func TestExampleErrorCases(t *testing.T) {
	// Create test environment
	env := test.NewTestEnvironment(t, test.WithServer())
	defer env.Cleanup()

	// Test case: Invalid job request (missing required fields)
	invalidReq := infrastructure.CreateRequest{
		// Missing required fields
	}

	// Attempt to create job with invalid request
	job, err := env.APIClient.CreateJob(env.Context(), invalidReq)
	assert.Error(t, err, "Expected error for invalid request")
	assert.Nil(t, job, "Job should be nil for invalid request")

	// Test case: Duplicate project name
	dupReq := infrastructure.CreateRequest{
		Name:        "dup-test-job",
		ProjectName: "dup-test-project",
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

	// Create first job
	job1, err := env.APIClient.CreateJob(env.Context(), dupReq)
	require.NoError(t, err, "Failed to create first job")
	require.NotNil(t, job1, "First job should not be nil")

	// Attempt to create second job with same project name
	job2, err := env.APIClient.CreateJob(env.Context(), dupReq)
	assert.Error(t, err, "Expected error for duplicate project name")
	assert.Nil(t, job2, "Second job should be nil")
}
