package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				BaseURL:    "http://example.com",
				HTTPClient: &http.Client{Timeout: 10 * time.Second},
				Timeout:    10 * time.Second,
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

func TestAPIClient_doRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)
	apiClient := client.(*APIClient)

	t.Run("success", func(t *testing.T) {
		req, err := apiClient.newRequest(context.Background(), http.MethodGet, "/success", nil)
		require.NoError(t, err)

		var response CreateResponse
		err = apiClient.doRequest(req, &response)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "success", response.Status)
	})

	t.Run("error response", func(t *testing.T) {
		req, err := apiClient.newRequest(context.Background(), http.MethodGet, "/error", nil)
		require.NoError(t, err)

		var response CreateResponse
		err = apiClient.doRequest(req, &response)
		assert.Error(t, err)

		apiErr, ok := IsAPIError(err)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
		assert.Equal(t, "Invalid request", apiErr.Message)
		assert.True(t, apiErr.IsBadRequest())
	})

	t.Run("invalid json", func(t *testing.T) {
		req, err := apiClient.newRequest(context.Background(), http.MethodGet, "/invalid-json", nil)
		require.NoError(t, err)

		var response CreateResponse
		err = apiClient.doRequest(req, &response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding response")
	})

	t.Run("not found", func(t *testing.T) {
		req, err := apiClient.newRequest(context.Background(), http.MethodGet, "/not-found", nil)
		require.NoError(t, err)

		var response CreateResponse
		err = apiClient.doRequest(req, &response)
		assert.Error(t, err)

		apiErr, ok := IsAPIError(err)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
		assert.True(t, apiErr.IsNotFound())
	})
}

func TestAPIClient_newRequest(t *testing.T) {
	client, err := NewClient(&ClientOptions{
		BaseURL: "http://example.com",
	})
	require.NoError(t, err)
	apiClient := client.(*APIClient)

	t.Run("valid request", func(t *testing.T) {
		req, err := apiClient.newRequest(context.Background(), http.MethodGet, "/test", nil)
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com/test", req.URL.String())
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", req.Header.Get("Accept"))
	})

	t.Run("invalid endpoint", func(t *testing.T) {
		req, err := apiClient.newRequest(context.Background(), http.MethodGet, "://invalid", nil)
		assert.Error(t, err)
		assert.Nil(t, req)
	})
}

func TestMarshalRequest(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		reader, err := marshalRequest(nil)
		assert.NoError(t, err)
		assert.Nil(t, reader)
	})

	t.Run("valid request", func(t *testing.T) {
		req := map[string]interface{}{
			"id":     1,
			"status": "active",
		}
		reader, err := marshalRequest(req)
		assert.NoError(t, err)
		assert.NotNil(t, reader)

		// Verify the marshaled JSON
		var unmarshaled map[string]interface{}
		err = json.NewDecoder(reader).Decode(&unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), unmarshaled["id"])
		assert.Equal(t, "active", unmarshaled["status"])
	})

	t.Run("invalid request", func(t *testing.T) {
		// Create a value that can't be marshaled to JSON
		req := make(chan int)
		reader, err := marshalRequest(req)
		assert.Error(t, err)
		assert.Nil(t, reader)
	})
}
