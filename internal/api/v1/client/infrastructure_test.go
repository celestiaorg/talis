package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

func TestAPIClient_CreateInfrastructure(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/instances", r.URL.Path)

		// Check the request body
		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-infra", req["name"])

		// Return a successful response
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 123, "status": "created"}`))
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Create a test request
	req := map[string]interface{}{
		"name":         "test-infra",
		"project_name": "test-project",
		"instances": []map[string]interface{}{
			{
				"provider":            "aws",
				"number_of_instances": 1,
				"region":              "us-west-2",
				"size":                "t2.micro",
			},
		},
	}

	// Call the method
	resp, err := client.CreateInfrastructure(context.Background(), req)
	require.NoError(t, err)

	// Check the response
	createResp, ok := resp.(*infrastructure.Response)
	require.True(t, ok)
	assert.Equal(t, uint(123), createResp.ID)
	assert.Equal(t, "created", createResp.Status)
}

func TestAPIClient_DeleteInfrastructure(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/instances", r.URL.Path)

		// Check the request body
		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, float64(123), req["id"])

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 123, "status": "deleted"}`))
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Create a test request
	req := map[string]interface{}{
		"id":           123,
		"name":         "test-infra",
		"project_name": "test-project",
		"instances": []map[string]interface{}{
			{
				"provider":            "aws",
				"number_of_instances": 1,
				"region":              "us-west-2",
				"size":                "t2.micro",
			},
		},
	}

	// Call the method
	resp, err := client.DeleteInfrastructure(context.Background(), req)
	require.NoError(t, err)

	// Check the response
	deleteResp, ok := resp.(*infrastructure.Response)
	require.True(t, ok)
	assert.Equal(t, uint(123), deleteResp.ID)
	assert.Equal(t, "deleted", deleteResp.Status)
}

func TestAPIClient_GetInfrastructure(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/instances/123", r.URL.Path)

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 123, "status": "active"}`))
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Call the method
	resp, err := client.GetInfrastructure(context.Background(), "123")
	require.NoError(t, err)

	// Check the response
	createResp, ok := resp.(*infrastructure.Response)
	require.True(t, ok)
	assert.Equal(t, uint(123), createResp.ID)
	assert.Equal(t, "active", createResp.Status)
}

func TestAPIClient_ListInfrastructure(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/instances", r.URL.Path)

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id": 123, "status": "active"}, {"id": 456, "status": "active"}]`))
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Call the method
	resp, err := client.ListInfrastructure(context.Background())
	require.NoError(t, err)

	// Check the response
	createResps, ok := resp.([]infrastructure.Response)
	require.True(t, ok)
	assert.Len(t, createResps, 2)
	assert.Equal(t, uint(123), createResps[0].ID)
	assert.Equal(t, "active", createResps[0].Status)
	assert.Equal(t, uint(456), createResps[1].ID)
	assert.Equal(t, "active", createResps[1].Status)
}
