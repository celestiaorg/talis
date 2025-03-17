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

	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    *ClientOptions
		wantErr bool
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "valid options",
			opts: &ClientOptions{
				BaseURL: "http://example.com",
				Timeout: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid base URL",
			opts: &ClientOptions{
				BaseURL: "://invalid-url",
			},
			wantErr: true,
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
			_, _ = w.Write([]byte(`{"error": "bad_request", "message": "Invalid request", "status": 400}`))
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
	client, err := NewClient(&ClientOptions{
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
	client, err := NewClient(&ClientOptions{
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
