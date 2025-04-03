package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

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

func setupTestServer() *httptest.Server {
	// Create a test server
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

	t.Run("success", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), http.MethodGet, "/success", nil)
		require.NoError(t, err)

		var response infrastructure.Response
		err = apiClient.doRequest(agent, &response)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "success", response.Status)
	})

	t.Run("error response", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), http.MethodGet, "/error", nil)
		require.NoError(t, err)

		var response infrastructure.Response
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

		var response infrastructure.Response
		err = apiClient.doRequest(agent, &response)
		assert.Error(t, err)

		var fiberErr *fiber.Error
		assert.False(t, errors.As(err, &fiberErr))
		assert.Contains(t, err.Error(), "error decoding response")
	})

	t.Run("not found", func(t *testing.T) {
		agent, err := apiClient.createAgent(context.Background(), http.MethodGet, "/not-found", nil)
		require.NoError(t, err)

		var response infrastructure.Response
		err = apiClient.doRequest(agent, &response)
		assert.Error(t, err)

		var fiberErr *fiber.Error
		assert.True(t, errors.As(err, &fiberErr))
		assert.Equal(t, http.StatusNotFound, fiberErr.Code)
	})
}

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

func TestGetQueryParams(t *testing.T) {
	tests := []struct {
		name    string
		opts    *models.ListOptions
		want    map[string][]string
		wantErr bool
	}{
		{
			name: "nil options",
			opts: nil,
			want: map[string][]string{},
		},
		{
			name: "empty options",
			opts: &models.ListOptions{},
			want: map[string][]string{},
		},
		{
			name: "pagination only",
			opts: &models.ListOptions{
				Limit:  10,
				Offset: 20,
			},
			want: map[string][]string{
				"limit":  {"10"},
				"offset": {"20"},
			},
		},
		{
			name: "include deleted only",
			opts: &models.ListOptions{
				IncludeDeleted: true,
			},
			want: map[string][]string{
				"include_deleted": {"true"},
			},
		},
		{
			name: "status ready",
			opts: &models.ListOptions{
				Status: func() *models.InstanceStatus {
					s := models.InstanceStatusReady
					return &s
				}(),
			},
			want: map[string][]string{
				"status": {"ready"},
			},
		},
		{
			name: "status terminated",
			opts: &models.ListOptions{
				Status: func() *models.InstanceStatus {
					s := models.InstanceStatusTerminated
					return &s
				}(),
			},
			want: map[string][]string{
				"status": {"terminated"},
			},
		},
		{
			name: "all options with status filter equal",
			opts: &models.ListOptions{
				Limit:          10,
				Offset:         20,
				IncludeDeleted: true,
				Status: func() *models.InstanceStatus {
					s := models.InstanceStatusReady
					return &s
				}(),
				StatusFilter: models.StatusFilterEqual,
			},
			want: map[string][]string{
				"limit":           {"10"},
				"offset":          {"20"},
				"include_deleted": {"true"},
				"status":          {"ready"},
				"status_filter":   {"equal"},
			},
		},
		{
			name: "all options with status filter not equal",
			opts: &models.ListOptions{
				Limit:          10,
				Offset:         20,
				IncludeDeleted: true,
				Status: func() *models.InstanceStatus {
					s := models.InstanceStatusTerminated
					return &s
				}(),
				StatusFilter: models.StatusFilterNotEqual,
			},
			want: map[string][]string{
				"limit":           {"10"},
				"offset":          {"20"},
				"include_deleted": {"true"},
				"status":          {"terminated"},
				"status_filter":   {"not_equal"},
			},
		},
		{
			name: "invalid status",
			opts: &models.ListOptions{
				Status: func() *models.InstanceStatus {
					s := models.InstanceStatus(999)
					return &s
				}(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getQueryParams(tt.opts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, map[string][]string(got))
		})
	}
}
