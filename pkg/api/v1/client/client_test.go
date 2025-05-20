// Package client provides unit tests for the Talis API client.
//
// These tests verify the client's functionality by testing:
// - Client creation and configuration
// - HTTP request/response handling
// - Error handling
// - Parameter conversion
//
// The tests use httptest to create a mock server that simulates the Talis API,
// allowing the client to be tested without requiring an actual API server.
package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// internaltypes "github.com/celestiaorg/talis/internal/types" // Use public alias
	"github.com/celestiaorg/talis/pkg/db/models"
	// Import public types alias
)

// Define status variables used in tests
// var pendingStatus = models.InstanceStatusPending // REMOVED - Unused

// invalidInstanceStatus represents an invalid instance status value used for testing error handling.
var invalidInstanceStatus = models.InstanceStatus(999)

// instanceStatusPtr is a helper function that returns a pointer to an InstanceStatus value.
// This is needed for tests that require a pointer to an InstanceStatus.
func instanceStatusPtr(status models.InstanceStatus) *models.InstanceStatus {
	return &status
}

// TestNewClient tests the NewClient function with various configurations.
// It verifies:
// - Default options are applied when nil options are provided
// - Custom options are properly applied
// - Invalid base URLs are properly rejected
func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		opts       *Options
		wantErr    bool
		validateFn func(t *testing.T, client Client)
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: false,
			validateFn: func(t *testing.T, client Client) {
				// Client should use default options
				apiClient, ok := client.(*APIClient)
				assert.True(t, ok, "client should be an *APIClient")

				// Verify default values are set
				expectedDefaults := DefaultOptions()
				assert.Equal(t, expectedDefaults.BaseURL, apiClient.baseURL)
				assert.Equal(t, expectedDefaults.Timeout, apiClient.timeout)
			},
		},
		{
			name: "valid options",
			opts: &Options{
				BaseURL: "http://example.com",
				Timeout: 10 * time.Second,
			},
			wantErr: false,
			validateFn: func(t *testing.T, client Client) {
				// Client should use provided options
				apiClient, ok := client.(*APIClient)
				assert.True(t, ok, "client should be an *APIClient")

				assert.Equal(t, "http://example.com", apiClient.baseURL)
				assert.Equal(t, 10*time.Second, apiClient.timeout)
			},
		},
		{
			name: "invalid base URL",
			opts: &Options{
				BaseURL: "://invalid-url",
			},
			wantErr:    true,
			validateFn: nil, // No validation for error case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)

				// Additional validation specific to each test case
				if tt.validateFn != nil {
					tt.validateFn(t, client)
				}
			}
		})
	}
}

// setupTestServer creates a mock HTTP server for testing the client.
// It provides several endpoints that simulate different API responses:
// - /success: Returns a successful JSON response
// - /error: Returns a 400 Bad Request error
// - /invalid-json: Returns malformed JSON to test error handling
// - Any other path: Returns a 404 Not Found error
func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": 1, "status": "success"}`))
		case "/error":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Invalid request"))
		case "/invalid-json":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{invalid json`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// TestAPIClient_doRequest tests the doRequest method of the APIClient.
// It verifies the client correctly:
// - Processes successful responses and unmarshals JSON data
// - Handles HTTP error responses (4xx, 5xx status codes)
// - Handles malformed JSON responses
// - Handles 404 Not Found responses
func TestAPIClient_doRequest(t *testing.T) {
	// Create a test server
	server := setupTestServer()
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&Options{
		BaseURL: server.URL,
	})
	require.NoError(t, err)
	apiClient := client.(*APIClient)

	// Define a local struct for testing doRequest decoding
	type testResponse struct {
		ID     uint   `json:"id"`
		Status string `json:"status"`
	}

	t.Run("success", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), http.MethodGet, "/success", nil)
		require.NoError(t, err)

		var response testResponse // Use the struct defined above
		err = apiClient.doRequest(agent, &response)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "success", response.Status)
	})

	t.Run("error response", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), http.MethodGet, "/error", nil)
		require.NoError(t, err)

		var response testResponse // Use the struct defined above
		err = apiClient.doRequest(agent, &response)
		assert.Error(t, err)

		var fiberErr *fiber.Error
		assert.True(t, errors.As(err, &fiberErr))
		assert.Equal(t, http.StatusBadRequest, fiberErr.Code)
		assert.Equal(t, "Invalid request", fiberErr.Message)
	})

	t.Run("invalid json", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), http.MethodGet, "/invalid-json", nil)
		require.NoError(t, err)

		var response testResponse // Use the struct defined above
		err = apiClient.doRequest(agent, &response)
		assert.Error(t, err)

		var fiberErr *fiber.Error
		assert.False(t, errors.As(err, &fiberErr))
		assert.Contains(t, err.Error(), "error decoding response")
	})

	t.Run("not found", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), http.MethodGet, "/not-found", nil)
		require.NoError(t, err)

		var response testResponse // Use the struct defined above
		err = apiClient.doRequest(agent, &response)
		assert.Error(t, err)

		var fiberErr *fiber.Error
		assert.True(t, errors.As(err, &fiberErr))
		assert.Equal(t, http.StatusNotFound, fiberErr.Code)
	})
}

// TestAPIClient_createAgent tests the createAgent method of the APIClient.
// It verifies the client correctly:
// - Creates agents for valid HTTP methods
// - Rejects invalid HTTP methods
// - Properly attaches request bodies
// - Respects context deadlines
func TestAPIClient_createAgent(t *testing.T) {
	client, err := NewClient(&Options{
		BaseURL: "http://example.com",
	})
	require.NoError(t, err)
	apiClient := client.(*APIClient)

	t.Run("valid request", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), http.MethodGet, "/test", nil)
		assert.NoError(t, err)
		assert.NotNil(t, agent)
	})

	t.Run("unsupported method", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), "INVALID", "/test", nil)
		assert.Error(t, err)
		assert.Nil(t, agent)
		assert.Contains(t, err.Error(), "unsupported HTTP method")
	})

	t.Run("with body", func(t *testing.T) {
		body := map[string]interface{}{
			"id":     1,
			"status": "active",
		}
		agent, err := apiClient.createAgent(context.Background(), http.MethodPost, "/test", body)
		assert.NoError(t, err)
		assert.NotNil(t, agent)
	})

	t.Run("with context deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		agent, err := apiClient.createAgent(ctx, http.MethodGet, "/test", nil)
		assert.NoError(t, err)
		assert.NotNil(t, agent)
	})
}

// TestGetQueryParams tests the getQueryParams helper function.
// It verifies the function correctly:
// - Handles nil and empty ListOptions
// - Converts pagination parameters (limit, offset)
// - Handles boolean flags (include_deleted)
// - Converts enum values (StatusFilter, InstanceStatus)
// - Properly validates and rejects invalid enum values
func TestGetQueryParams(t *testing.T) {
	tests := []struct {
		name    string
		opts    *models.ListOptions
		want    url.Values
		wantErr bool
	}{
		{
			name: "nil options",
			opts: nil,
			want: url.Values{},
		},
		{
			name: "empty options",
			opts: &models.ListOptions{},
			want: url.Values{},
		},
		{
			name: "pagination only",
			opts: &models.ListOptions{
				Limit:  10,
				Offset: 20,
			},
			want: url.Values{
				"limit":  {"10"},
				"offset": {"20"},
			},
		},
		{
			name: "include deleted true",
			opts: &models.ListOptions{
				IncludeDeleted: true,
			},
			want: url.Values{
				"include_deleted": {"true"},
			},
		},
		{
			name: "include deleted false",
			opts: &models.ListOptions{
				IncludeDeleted: false,
			},
			want: url.Values{},
		},
		{
			name: "status filter equal",
			opts: &models.ListOptions{
				StatusFilter: models.StatusFilterEqual,
			},
			want: url.Values{
				"status_filter": {"equal"},
			},
		},
		{
			name: "status filter not equal",
			opts: &models.ListOptions{
				StatusFilter: models.StatusFilterNotEqual,
			},
			want: url.Values{
				"status_filter": {"not_equal"},
			},
		},
		{
			name: "instance status test",
			opts: &models.ListOptions{InstanceStatus: instanceStatusPtr(models.InstanceStatusPending)},
			want: url.Values{
				"instance_status": {"pending"},
			},
		},
		{
			name:    "invalid instance status",
			opts:    &models.ListOptions{InstanceStatus: instanceStatusPtr(invalidInstanceStatus)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getQueryParams(tt.opts)
			if tt.wantErr {
				require.Error(t, err, tt.name)
				return
			}
			require.NoError(t, err, tt.name)
			assert.Equal(t, tt.want, got, tt.name)
		})
	}
}
