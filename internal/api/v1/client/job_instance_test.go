package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

func TestAPIClient_CreateJobInstance(t *testing.T) {
	// Test data
	jobID := "job-123"
	instanceReq := infrastructure.InstanceRequest{
		Provider:          "aws",
		NumberOfInstances: 1,
		Provision:         true,
		Region:            "us-west-2",
		Size:              "t2.micro",
		Image:             "ami-12345",
		Tags:              []string{"test", "dev"},
		SSHKeyName:        "test-key",
	}

	// Expected response
	expectedResp := infrastructure.InstanceInfo{
		Name:     "instance-456",
		IP:       "10.0.0.1",
		Provider: "aws",
		Region:   "us-west-2",
		Size:     "t2.micro",
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, routes.CreateJobInstanceURL(jobID), r.URL.Path)

		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var receivedReq infrastructure.InstanceRequest
		err = json.Unmarshal(body, &receivedReq)
		require.NoError(t, err)

		// Verify the request matches what we sent
		assert.Equal(t, instanceReq.Provider, receivedReq.Provider)
		assert.Equal(t, instanceReq.NumberOfInstances, receivedReq.NumberOfInstances)
		assert.Equal(t, instanceReq.Region, receivedReq.Region)
		assert.Equal(t, instanceReq.Size, receivedReq.Size)

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		respBytes, _ := json.Marshal(expectedResp)
		_, _ = w.Write(respBytes)
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Call the method
	resp, err := client.CreateJobInstance(context.Background(), jobID, instanceReq)
	require.NoError(t, err)

	// Check the response
	assert.NotNil(t, resp)
	assert.Equal(t, expectedResp.Name, resp.Name)
	assert.Equal(t, expectedResp.IP, resp.IP)
	assert.Equal(t, expectedResp.Provider, resp.Provider)
	assert.Equal(t, expectedResp.Region, resp.Region)
	assert.Equal(t, expectedResp.Size, resp.Size)
}

func TestAPIClient_DeleteJobInstance(t *testing.T) {
	// Test data
	jobID := "job-123"
	deleteReq := infrastructure.DeleteInstanceRequest{
		ID:           456,
		InstanceName: "test-job",
		ProjectName:  "test-project",
		Instances: []infrastructure.InstanceRequest{
			{
				Provider:          "aws",
				NumberOfInstances: 1,
				Region:            "us-west-2",
			},
		},
	}

	// Expected response
	expectedResp := infrastructure.Response{
		ID:     456,
		Status: "deleted",
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, routes.DeleteJobInstanceURL(jobID), r.URL.Path)

		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var receivedReq infrastructure.DeleteInstanceRequest
		err = json.Unmarshal(body, &receivedReq)
		require.NoError(t, err)

		// Verify the request matches what we sent
		assert.Equal(t, deleteReq.ID, receivedReq.ID)
		assert.Equal(t, deleteReq.InstanceName, receivedReq.InstanceName)
		assert.Equal(t, deleteReq.ProjectName, receivedReq.ProjectName)
		assert.Len(t, receivedReq.Instances, len(deleteReq.Instances))

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		respBytes, _ := json.Marshal(expectedResp)
		_, _ = w.Write(respBytes)
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Call the method
	resp, err := client.DeleteJobInstance(context.Background(), jobID, deleteReq)
	require.NoError(t, err)

	// Check the response
	assert.NotNil(t, resp)
	assert.Equal(t, expectedResp.ID, resp.ID)
	assert.Equal(t, expectedResp.Status, resp.Status)
}

func TestAPIClient_GetJobInstance(t *testing.T) {
	// Test data
	jobID := "job-123"
	instanceID := "instance-456"

	// Expected response
	expectedResp := infrastructure.InstanceInfo{
		Name:     "instance-456",
		IP:       "10.0.0.1",
		Provider: "aws",
		Region:   "us-west-2",
		Size:     "t2.micro",
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, routes.GetJobInstanceURL(jobID, instanceID), r.URL.Path)

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		respBytes, _ := json.Marshal(expectedResp)
		_, _ = w.Write(respBytes)
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Call the method
	resp, err := client.GetJobInstance(context.Background(), jobID, instanceID)
	require.NoError(t, err)

	// Check the response
	assert.NotNil(t, resp)
	assert.Equal(t, expectedResp.Name, resp.Name)
	assert.Equal(t, expectedResp.IP, resp.IP)
	assert.Equal(t, expectedResp.Provider, resp.Provider)
	assert.Equal(t, expectedResp.Region, resp.Region)
	assert.Equal(t, expectedResp.Size, resp.Size)
}

func TestAPIClient_GetJobInstances(t *testing.T) {
	// Test data
	jobID := "job-123"

	// Expected response
	expectedResp := []infrastructure.InstanceInfo{
		{
			Name:     "instance-456",
			IP:       "10.0.0.1",
			Provider: "aws",
			Region:   "us-west-2",
			Size:     "t2.micro",
		},
		{
			Name:     "instance-789",
			IP:       "10.0.0.2",
			Provider: "aws",
			Region:   "us-west-2",
			Size:     "t2.micro",
		},
	}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, routes.GetJobInstancesURL(jobID), r.URL.Path)

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		respBytes, _ := json.Marshal(expectedResp)
		_, _ = w.Write(respBytes)
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Call the method
	resp, err := client.GetJobInstances(context.Background(), jobID)
	require.NoError(t, err)

	// Check the response
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, expectedResp[0].Name, resp[0].Name)
	assert.Equal(t, expectedResp[0].IP, resp[0].IP)
	assert.Equal(t, expectedResp[1].Name, resp[1].Name)
	assert.Equal(t, expectedResp[1].IP, resp[1].IP)
}

func TestAPIClient_GetJobPublicIPs(t *testing.T) {
	// Test data
	jobID := "job-123"

	// Expected response
	expectedResp := []string{"1.2.3.4", "5.6.7.8"}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, routes.GetJobPublicIPsURL(jobID), r.URL.Path)

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		respBytes, _ := json.Marshal(expectedResp)
		_, _ = w.Write(respBytes)
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Call the method
	resp, err := client.GetJobPublicIPs(context.Background(), jobID)
	require.NoError(t, err)

	// Check the response
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)
	assert.Equal(t, expectedResp[0], resp[0])
	assert.Equal(t, expectedResp[1], resp[1])
}
