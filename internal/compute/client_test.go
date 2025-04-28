package compute

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/celestiaorg/talis/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.VirtFusionConfig
		wantError bool
	}{
		{
			name: "valid configuration",
			config: &config.VirtFusionConfig{
				APIToken: "test-token",
				Host:     "https://api.virtfusion.com",
			},
			wantError: false,
		},
		{
			name: "missing token",
			config: &config.VirtFusionConfig{
				APIToken: "",
				Host:     "https://api.virtfusion.com",
			},
			wantError: true,
		},
		{
			name: "missing host",
			config: &config.VirtFusionConfig{
				APIToken: "test-token",
				Host:     "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.config.Host, client.baseURL)
			}
		})
	}
}

func TestClient_TestConnection(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantError  bool
	}{
		{
			name:       "successful connection",
			statusCode: http.StatusOK,
			response:   `{"status": "ok"}`,
			wantError:  false,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   `{"code": "unauthorized", "message": "Invalid token"}`,
			wantError:  true,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   `{"code": "internal_error", "message": "Internal server error"}`,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request
				assert.Equal(t, "/connect", r.URL.Path)
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Send response
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			// Create client
			client, err := NewClient(&config.VirtFusionConfig{
				APIToken: "test-token",
				Host:     server.URL,
			})
			require.NoError(t, err)

			// Test connection
			err = client.TestConnection(context.Background())
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_doRequest_Retries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(&config.VirtFusionConfig{
		APIToken: "test-token",
		Host:     server.URL,
	})
	require.NoError(t, err)

	resp, err := client.doRequest(context.Background(), http.MethodGet, "/test", nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attempts)
}
