package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/api/v1/routes"
)

func TestAPIClient_GetJob(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method and path
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, routes.GetJobURL("123"), r.URL.Path)

		// Return a successful response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"job_id": "123", "status": "completed", "created_at": "2023-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	// Call the method
	resp, err := client.GetJob(context.Background(), "123")
	require.NoError(t, err)

	// Check the response
	assert.NotNil(t, resp)
	assert.Equal(t, "123", resp.JobID)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, "2023-01-01T00:00:00Z", resp.CreatedAt)
}

func TestAPIClient_ListJobs(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		status string
		path   string
	}{
		{
			name:   "no filters",
			limit:  0,
			status: "",
			path:   routes.ListJobsURL(),
		},
		{
			name:   "with limit",
			limit:  5,
			status: "",
			path:   routes.ListJobsURL() + "?limit=5",
		},
		{
			name:   "with status",
			limit:  0,
			status: "running",
			path:   routes.ListJobsURL() + "?status=running",
		},
		{
			name:   "with limit and status",
			limit:  5,
			status: "running",
			path:   routes.ListJobsURL() + "?limit=5&status=running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request method and path
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.path, r.URL.RequestURI())

				// Return a successful response
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[
					{"job_id": "123", "status": "running", "created_at": "2023-01-01T00:00:00Z"},
					{"job_id": "456", "status": "completed", "created_at": "2023-01-02T00:00:00Z"}
				]`))
			}))
			defer server.Close()

			// Create a client with the test server URL
			client, err := NewClient(&ClientOptions{
				BaseURL: server.URL,
			})
			require.NoError(t, err)

			// Call the method
			resp, err := client.ListJobs(context.Background(), tt.limit, tt.status)
			require.NoError(t, err)

			// Check the response
			assert.Len(t, resp, 2)
			assert.Equal(t, "123", resp[0].JobID)
			assert.Equal(t, "running", resp[0].Status)
			assert.Equal(t, "456", resp[1].JobID)
			assert.Equal(t, "completed", resp[1].Status)
		})
	}
}
