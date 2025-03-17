package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ClientOptions
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
			opts: &ClientOptions{
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
			opts: &ClientOptions{
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

// TestEndToEndClientWorkflow tests the end-to-end client workflow.
// Right now it is pretty abstracted. In the future we should add mocks to the
// DB package and pull this test out into a /test package to better replicate e2e testing.
func TestEndToEndClientWorkflow(t *testing.T) {
	// Setup test data
	jobID := "job-123"
	instanceID := "instance-456"
	instanceIP := "10.20.30.40"

	// In-memory data store to simulate a database
	type testDB struct {
		// Jobs data
		jobs       map[string]*infrastructure.JobStatus
		jobCounter uint

		// Instances data
		instances      map[string]*infrastructure.InstanceInfo
		instancesByJob map[string][]*infrastructure.InstanceInfo
	}

	db := &testDB{
		jobs:           make(map[string]*infrastructure.JobStatus),
		instances:      make(map[string]*infrastructure.InstanceInfo),
		instancesByJob: make(map[string][]*infrastructure.InstanceInfo),
	}

	// Create a more realistic mock server that uses the in-memory DB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Setup response content type
		w.Header().Set("Content-Type", "application/json")

		switch {
		// ===== Job endpoints =====
		case r.URL.Path == routes.CreateJobURL() && r.Method == http.MethodPost:
			// Create a new job
			var req infrastructure.CreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request body"})
				return
			}

			db.jobCounter++
			newJobID := jobID // Use predefined ID for test simplicity

			newJob := &infrastructure.JobStatus{
				JobID:     newJobID,
				Status:    "pending",
				CreatedAt: time.Now().Format(time.RFC3339),
			}
			db.jobs[newJobID] = newJob

			// Return response
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(&infrastructure.Response{
				ID:     db.jobCounter,
				Status: "pending",
			})

		case r.URL.Path == routes.ListJobsURL() && r.Method == http.MethodGet:
			// List all jobs
			jobList := make([]infrastructure.JobStatus, 0, len(db.jobs))
			for _, job := range db.jobs {
				jobList = append(jobList, *job)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jobList)

		case r.Method == http.MethodGet && r.URL.Path == routes.GetJobURL(jobID):
			// Get job by ID
			job, exists := db.jobs[jobID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Job not found"})
				return
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(job)

		// ===== Job Instance endpoints =====
		case r.Method == http.MethodGet && r.URL.Path == routes.GetJobInstancesURL(jobID):
			// Get instances for job
			_, exists := db.jobs[jobID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Job not found"})
				return
			}

			instances := db.instancesByJob[jobID]
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(instances)

		case r.Method == http.MethodPost && r.URL.Path == routes.CreateJobInstanceURL(jobID):
			// Create instance for job
			job, exists := db.jobs[jobID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Job not found"})
				return
			}

			var req infrastructure.InstanceRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request body"})
				return
			}

			// Create the new instance
			newInstance := &infrastructure.InstanceInfo{
				Name:     instanceID, // Use predefined ID for test simplicity
				IP:       instanceIP,
				Provider: req.Provider,
				Region:   req.Region,
				Size:     req.Size,
			}

			// Store in our "database"
			db.instances[instanceID] = newInstance
			db.instancesByJob[jobID] = append(db.instancesByJob[jobID], newInstance)

			// Update job status if needed
			job.Status = "running"

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(newInstance)

		case r.Method == http.MethodGet && r.URL.Path == routes.GetJobInstanceURL(jobID, instanceID):
			// Get specific instance for job
			_, jobExists := db.jobs[jobID]
			if !jobExists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Job not found"})
				return
			}

			instance, instanceExists := db.instances[instanceID]
			if !instanceExists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Instance not found"})
				return
			}

			// Verify this instance belongs to the job
			belongs := false
			for _, inst := range db.instancesByJob[jobID] {
				if inst.Name == instanceID {
					belongs = true
					break
				}
			}

			if !belongs {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Instance not found for this job"})
				return
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(instance)

		case r.Method == http.MethodGet && r.URL.Path == routes.GetJobPublicIPsURL(jobID):
			// Get IPs for job
			_, jobExists := db.jobs[jobID]
			if !jobExists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Job not found"})
				return
			}

			// Collect IPs
			var ips []string
			for _, instance := range db.instancesByJob[jobID] {
				if instance.IP != "" {
					ips = append(ips, instance.IP)
				}
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(ips)

		case r.Method == http.MethodDelete && r.URL.Path == routes.DeleteJobInstanceURL(jobID):
			// Delete instance for job
			_, jobExists := db.jobs[jobID]
			if !jobExists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Job not found"})
				return
			}

			var req infrastructure.DeleteInstanceRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid request body"})
				return
			}

			// Remove the instances (for simplicity, we'll remove all instances for this job)
			for _, inst := range db.instancesByJob[jobID] {
				delete(db.instances, inst.Name)
			}
			delete(db.instancesByJob, jobID)

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(&infrastructure.Response{
				ID:     req.ID,
				Status: "deleted",
			})

		// ===== Instance endpoints =====
		case r.Method == http.MethodGet && r.URL.Path == routes.ListInstancesURL():
			// List all instances
			allInstances := make([]infrastructure.InstanceInfo, 0)
			for _, instance := range db.instances {
				allInstances = append(allInstances, *instance)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(allInstances)

		case r.Method == http.MethodGet && r.URL.Path == routes.GetInstanceURL(instanceID):
			// Get instance by ID
			instance, exists := db.instances[instanceID]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Instance not found"})
				return
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(instance)

		case r.Method == http.MethodGet && r.URL.Path == routes.GetInstanceMetadataURL():
			// Get instance metadata
			allInstances := make([]infrastructure.InstanceInfo, 0)
			for _, instance := range db.instances {
				allInstances = append(allInstances, *instance)
			}

			metadata := &InstanceMetadataResponse{
				Instances: allInstances,
				Total:     len(allInstances),
				Page:      1,
				Limit:     10,
				Offset:    0,
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(metadata)

		case r.Method == http.MethodPost && r.URL.Path == routes.CreateJobInstanceURL("nonexistent-job"):
			// Special case for error testing
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Job not found"})

		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Endpoint not found"})
		}
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(&ClientOptions{
		BaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// 1. Try to create an instance for a non-existent job (should fail)
	t.Run("Try to create instance for non-existent job", func(t *testing.T) {
		instanceReq := infrastructure.InstanceRequest{
			Provider:          "aws",
			NumberOfInstances: 1,
			Provision:         true,
			Region:            "us-west-2",
			Size:              "t2.micro",
		}

		_, err := client.CreateJobInstance(ctx, "nonexistent-job", instanceReq)
		require.Error(t, err, "Should fail when creating instance for non-existent job")

		var fiberErr *fiber.Error
		require.True(t, errors.As(err, &fiberErr))
		require.Equal(t, http.StatusNotFound, fiberErr.Code)
	})

	// 2. Create a job
	t.Run("Create a job", func(t *testing.T) {
		createReq := infrastructure.CreateRequest{
			Name:        "test-job",
			ProjectName: "test-project",
			WebhookURL:  "https://example.com/webhook",
			Instances:   []infrastructure.InstanceRequest{},
		}

		resp, err := client.CreateJob(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, uint(1), resp.ID)
		require.Equal(t, "pending", resp.Status)

		// Verify it's in our DB
		require.Len(t, db.jobs, 1)
		require.Contains(t, db.jobs, jobID)
	})

	// 3. List jobs to verify the new job appears
	t.Run("List jobs", func(t *testing.T) {
		jobs, err := client.ListJobs(ctx, 0, "")
		require.NoError(t, err)
		require.Len(t, jobs, 1)
		require.Equal(t, jobID, jobs[0].JobID)
		require.Equal(t, "pending", jobs[0].Status)
	})

	// 4. Get the job by ID to verify it's returned
	t.Run("Get job by ID", func(t *testing.T) {
		job, err := client.GetJob(ctx, jobID)
		require.NoError(t, err)
		require.NotNil(t, job)
		require.Equal(t, jobID, job.JobID)
		require.Equal(t, "pending", job.Status)
	})

	// 5. Verify no instances exist for the job yet
	t.Run("Check no instances exist initially", func(t *testing.T) {
		instances, err := client.GetJobInstances(ctx, jobID)
		require.NoError(t, err)
		require.Empty(t, instances, "Should have no instances initially")

		// Verify our DB state
		require.Empty(t, db.instancesByJob[jobID])
	})

	// 6. Create an instance for the job
	t.Run("Create job instance", func(t *testing.T) {
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

		instance, err := client.CreateJobInstance(ctx, jobID, instanceReq)
		require.NoError(t, err)
		require.NotNil(t, instance)
		require.Equal(t, instanceID, instance.Name)
		require.Equal(t, instanceIP, instance.IP)

		// Verify our DB state
		require.Len(t, db.instances, 1)
		require.Contains(t, db.instances, instanceID)
		require.Len(t, db.instancesByJob[jobID], 1)

		// Job status should be updated
		require.Equal(t, "running", db.jobs[jobID].Status)
	})

	// 7. Verify the instance was created with GetJobInstance and GetJobInstances
	t.Run("Verify instance with GetJobInstance", func(t *testing.T) {
		instance, err := client.GetJobInstance(ctx, jobID, instanceID)
		require.NoError(t, err)
		require.NotNil(t, instance)
		require.Equal(t, instanceID, instance.Name)
		require.Equal(t, instanceIP, instance.IP)
		require.Equal(t, "aws", instance.Provider)
	})

	t.Run("Verify instance with GetJobInstances", func(t *testing.T) {
		instances, err := client.GetJobInstances(ctx, jobID)
		require.NoError(t, err)
		require.Len(t, instances, 1)
		require.Equal(t, instanceID, instances[0].Name)
		require.Equal(t, instanceIP, instances[0].IP)
	})

	// 8. Get the job's public IPs and verify they match
	t.Run("Verify job public IPs", func(t *testing.T) {
		ips, err := client.GetJobPublicIPs(ctx, jobID)
		require.NoError(t, err)
		require.Len(t, ips, 1)
		require.Equal(t, instanceIP, ips[0])
	})

	// 9. List all instances and verify our instance appears
	t.Run("Verify instance in ListInstances", func(t *testing.T) {
		instances, err := client.ListInstances(ctx)
		require.NoError(t, err)
		require.Len(t, instances, 1)
		require.Equal(t, instanceID, instances[0].Name)
		require.Equal(t, instanceIP, instances[0].IP)
	})

	// 10. Get instance by ID and check instance metadata
	t.Run("Get instance by ID", func(t *testing.T) {
		instance, err := client.GetInstance(ctx, instanceID)
		require.NoError(t, err)
		require.NotNil(t, instance)
		require.Equal(t, instanceID, instance.Name)
		require.Equal(t, instanceIP, instance.IP)
	})

	t.Run("Get instance metadata", func(t *testing.T) {
		metadata, err := client.GetInstanceMetadata(ctx)
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, 1, metadata.Total)
		require.Len(t, metadata.Instances, 1)
		require.Equal(t, instanceID, metadata.Instances[0].Name)
		require.Equal(t, instanceIP, metadata.Instances[0].IP)
	})

	// 11. Delete the instance and verify it's gone
	t.Run("Delete job instance", func(t *testing.T) {
		deleteReq := infrastructure.DeleteInstanceRequest{
			ID:          123,
			Name:        "test-job",
			ProjectName: "test-project",
			Instances: []infrastructure.InstanceRequest{
				{
					Provider:          "aws",
					NumberOfInstances: 1,
					Region:            "us-west-2",
				},
			},
		}

		resp, err := client.DeleteJobInstance(ctx, jobID, deleteReq)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, uint(123), resp.ID)
		require.Equal(t, "deleted", resp.Status)

		// Verify the instance is deleted from our DB
		require.Empty(t, db.instancesByJob[jobID])
		require.NotContains(t, db.instances, instanceID)

		// Verify the job still exists
		require.Contains(t, db.jobs, jobID)
	})
}
