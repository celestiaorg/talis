package client

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIntegrationTest skips the test if the INTEGRATION_TEST environment variable is not set
func skipIntegrationTest(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}
}

func TestIntegration_CreateAndDeleteInfrastructure(t *testing.T) {
	skipIntegrationTest(t)

	// Create a client with the default options (localhost:8080)
	client, err := NewClient(nil)
	require.NoError(t, err)

	// Create a test request
	createReq := map[string]interface{}{
		"name":         "test-integration-infra",
		"project_name": "test-project",
		"instances": []map[string]interface{}{
			{
				"provider":            "aws",
				"number_of_instances": 1,
				"region":              "us-west-2",
				"size":                "t2.micro",
				"image":               "ami-12345",
				"tags":                []string{"test", "integration"},
				"ssh_key_name":        "test-key",
			},
		},
	}

	// Create infrastructure
	createResp, err := client.CreateInfrastructure(context.Background(), createReq)
	require.NoError(t, err)

	// Check the response
	resp, ok := createResp.(*CreateResponse)
	require.True(t, ok)
	assert.NotZero(t, resp.ID)
	assert.Equal(t, "created", resp.Status)

	// Create a delete request
	deleteReq := map[string]interface{}{
		"id":           resp.ID,
		"name":         "test-integration-infra",
		"project_name": "test-project",
		"instances": []map[string]interface{}{
			{
				"provider":            "aws",
				"number_of_instances": 1,
				"region":              "us-west-2",
				"size":                "t2.micro",
				"image":               "ami-12345",
				"tags":                []string{"test", "integration"},
				"ssh_key_name":        "test-key",
			},
		},
	}

	// Delete infrastructure
	deleteResp, err := client.DeleteInfrastructure(context.Background(), deleteReq)
	require.NoError(t, err)

	// Check the response
	delResp, ok := deleteResp.(*DeleteResponse)
	require.True(t, ok)
	assert.Equal(t, resp.ID, delResp.ID)
	assert.Equal(t, "deleted", delResp.Status)
}

func TestIntegration_GetAndListJobs(t *testing.T) {
	skipIntegrationTest(t)

	// Create a client with the default options (localhost:8080)
	client, err := NewClient(nil)
	require.NoError(t, err)

	// List jobs
	listResp, err := client.ListJobs(context.Background(), 5, "")
	require.NoError(t, err)

	// Check the response
	jobs, ok := listResp.([]map[string]interface{})
	require.True(t, ok)

	// If there are jobs, get the first one
	if len(jobs) > 0 {
		jobID, ok := jobs[0]["job_id"].(string)
		require.True(t, ok)

		// Get the job
		getResp, err := client.GetJob(context.Background(), jobID)
		require.NoError(t, err)

		// Check the response
		job, ok := getResp.(*map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, jobID, (*job)["job_id"])
	} else {
		t.Log("No jobs found to test GetJob")
	}
}

func TestIntegration_GetAndListInfrastructure(t *testing.T) {
	skipIntegrationTest(t)

	// Create a client with the default options (localhost:8080)
	client, err := NewClient(nil)
	require.NoError(t, err)

	// List infrastructure
	listResp, err := client.ListInfrastructure(context.Background())
	require.NoError(t, err)

	// Check the response
	infras, ok := listResp.([]CreateResponse)
	require.True(t, ok)

	// If there are infrastructures, get the first one
	if len(infras) > 0 {
		infraID := infras[0].ID

		// Get the infrastructure
		getResp, err := client.GetInfrastructure(context.Background(), fmt.Sprintf("%d", infraID))
		require.NoError(t, err)

		// Check the response
		infra, ok := getResp.(*CreateResponse)
		require.True(t, ok)
		assert.Equal(t, infraID, infra.ID)
	} else {
		t.Log("No infrastructure found to test GetInfrastructure")
	}
}
