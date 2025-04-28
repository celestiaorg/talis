package compute

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/celestiaorg/talis/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHypervisorService_List(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantError  bool
		wantCount  int
	}{
		{
			name: "successful list",
			response: `{
				"hypervisors": [
					{
						"id": 1,
						"name": "hypervisor-1",
						"status": "active",
						"ipAddress": "1.2.3.4",
						"location": "nyc",
						"totalMemory": 32768,
						"usedMemory": 16384,
						"totalDisk": 1000,
						"usedDisk": 500
					},
					{
						"id": 2,
						"name": "hypervisor-2",
						"status": "active",
						"ipAddress": "1.2.3.5",
						"location": "nyc",
						"totalMemory": 32768,
						"usedMemory": 8192,
						"totalDisk": 1000,
						"usedDisk": 250
					}
				]
			}`,
			statusCode: http.StatusOK,
			wantError:  false,
			wantCount:  2,
		},
		{
			name:       "empty list",
			response:   `{"hypervisors": []}`,
			statusCode: http.StatusOK,
			wantError:  false,
			wantCount:  0,
		},
		{
			name:       "error response",
			response:   `{"code": "internal_error", "message": "Internal server error"}`,
			statusCode: http.StatusInternalServerError,
			wantError:  true,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/compute/hypervisors", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			// Create client and service
			client, err := NewClient(&config.VirtFusionConfig{
				APIToken: "test-token",
				Host:     server.URL,
			})
			require.NoError(t, err)

			service := newHypervisorService(client)

			// Test List method
			hypervisors, resp, err := service.List(context.Background())
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, hypervisors)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.statusCode, resp.StatusCode)
				assert.Len(t, hypervisors, tt.wantCount)

				if tt.wantCount > 0 {
					assert.Equal(t, 1, hypervisors[0].ID)
					assert.Equal(t, "hypervisor-1", hypervisors[0].Name)
					assert.Equal(t, "active", hypervisors[0].Status)
				}
			}
		})
	}
}

func TestHypervisorService_Get(t *testing.T) {
	tests := []struct {
		name         string
		hypervisorID int
		response     string
		statusCode   int
		wantError    bool
	}{
		{
			name:         "successful get",
			hypervisorID: 1,
			response: `{
				"id": 1,
				"name": "hypervisor-1",
				"status": "active",
				"ipAddress": "1.2.3.4",
				"location": "nyc",
				"totalMemory": 32768,
				"usedMemory": 16384,
				"totalDisk": 1000,
				"usedDisk": 500
			}`,
			statusCode: http.StatusOK,
			wantError:  false,
		},
		{
			name:         "not found",
			hypervisorID: 999,
			response: `{
				"code": "not_found",
				"message": "Hypervisor not found"
			}`,
			statusCode: http.StatusNotFound,
			wantError:  true,
		},
		{
			name:         "server error",
			hypervisorID: 1,
			response: `{
				"code": "internal_error",
				"message": "Internal server error"
			}`,
			statusCode: http.StatusInternalServerError,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/compute/hypervisors/%d", tt.hypervisorID)
				assert.Equal(t, expectedPath, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			// Create client and service
			client, err := NewClient(&config.VirtFusionConfig{
				APIToken: "test-token",
				Host:     server.URL,
			})
			require.NoError(t, err)

			service := newHypervisorService(client)

			// Test Get method
			hypervisor, resp, err := service.Get(context.Background(), tt.hypervisorID)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, hypervisor)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.statusCode, resp.StatusCode)
				assert.NotNil(t, hypervisor)
				assert.Equal(t, tt.hypervisorID, hypervisor.ID)
				assert.Equal(t, "hypervisor-1", hypervisor.Name)
				assert.Equal(t, "active", hypervisor.Status)
			}
		})
	}
}
