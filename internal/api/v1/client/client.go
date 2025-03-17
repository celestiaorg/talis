package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// DefaultTimeout is the default timeout for API requests
const DefaultTimeout = 30 * time.Second

// Client defines the interface for interacting with the Talis API
type Client interface {
	// Jobs methods
	CreateJob(ctx context.Context, req infrastructure.CreateRequest) (*infrastructure.Response, error)
	GetJob(ctx context.Context, id string) (*infrastructure.JobStatus, error)
	ListJobs(ctx context.Context, limit int, status string) ([]infrastructure.JobStatus, error)

	// Job instances methods
	CreateJobInstance(ctx context.Context, jobID string, req infrastructure.InstanceRequest) (*infrastructure.InstanceInfo, error)
	DeleteJobInstance(ctx context.Context, jobID string, req infrastructure.DeleteInstanceRequest) (*infrastructure.Response, error)
	GetJobInstance(ctx context.Context, jobID string, instanceID string) (*infrastructure.InstanceInfo, error)
	GetJobInstances(ctx context.Context, jobID string) ([]infrastructure.InstanceInfo, error)
	GetJobPublicIPs(ctx context.Context, jobID string) ([]string, error)

	// Instance methods
	GetInstance(ctx context.Context, id string) (*infrastructure.InstanceInfo, error)
	GetInstanceMetadata(ctx context.Context) (map[string]interface{}, error)
	ListInstances(ctx context.Context) ([]infrastructure.InstanceInfo, error)

	// Health check
	HealthCheck(ctx context.Context) (map[string]string, error)
}

// ClientOptions contains configuration options for the API client
type ClientOptions struct {
	// BaseURL is the base URL of the API
	BaseURL string

	// Timeout is the request timeout
	Timeout time.Duration
}

// DefaultOptions returns the default client options
func DefaultOptions() *ClientOptions {
	return &ClientOptions{
		BaseURL: routes.DefaultBaseURL,
		Timeout: DefaultTimeout,
	}
}

// APIClient implements the Client interface
type APIClient struct {
	baseURL string
	timeout time.Duration
}

// NewClient creates a new API client with the given options
func NewClient(opts *ClientOptions) (Client, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Validate the base URL
	_, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &APIClient{
		baseURL: opts.BaseURL,
		timeout: opts.Timeout,
	}, nil
}

// createAgent creates a new Fiber Agent for the given method and endpoint
func (c *APIClient) createAgent(ctx context.Context, method, endpoint string, body interface{}) (*fiber.Agent, error) {
	// Resolve the endpoint URL
	fullURL := c.baseURL + endpoint

	// Create a new agent based on the HTTP method
	var agent *fiber.Agent
	switch method {
	case http.MethodGet:
		agent = fiber.Get(fullURL)
	case http.MethodPost:
		agent = fiber.Post(fullURL)
	case http.MethodPut:
		agent = fiber.Put(fullURL)
	case http.MethodDelete:
		agent = fiber.Delete(fullURL)
	case http.MethodPatch:
		agent = fiber.Patch(fullURL)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// Set timeout from context or client default
	if deadline, ok := ctx.Deadline(); ok {
		agent.Timeout(time.Until(deadline))
	} else {
		agent.Timeout(c.timeout)
	}

	// Set common headers
	agent.Set("Content-Type", "application/json")
	agent.Set("Accept", "application/json")

	// Add body if provided
	if body != nil {
		agent.JSON(body)
	}

	return agent, nil
}

// executeRequest creates an agent, sends the request, and processes the response
func (c *APIClient) executeRequest(ctx context.Context, method, endpoint string, body, response interface{}) error {
	agent, err := c.createAgent(ctx, method, endpoint, body)
	if err != nil {
		return err
	}

	return c.doRequest(agent, response)
}

// doRequest sends the HTTP request and processes the response
func (c *APIClient) doRequest(agent *fiber.Agent, v interface{}) error {
	// Execute the request
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return fmt.Errorf("error sending request: %w", errs[0])
	}

	// Check for non-success status codes
	if statusCode < 200 || statusCode >= 300 {
		// Try to decode the error response
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return &fiber.Error{
				Code:    statusCode,
				Message: errResp.Message,
			}
		}

		// If we can't decode the error response, return a generic error
		return &fiber.Error{
			Code:    statusCode,
			Message: "unknown error",
		}
	}

	// Decode the response body if a target is provided
	if v != nil && len(body) > 0 {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
	}

	return nil
}

// marshalRequest marshals the request body to JSON
func marshalRequest(req interface{}) (io.Reader, error) {
	if req == nil {
		return nil, nil
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	return bytes.NewBuffer(jsonData), nil
}

// Jobs methods implementation

// CreateJob creates a new job
func (c *APIClient) CreateJob(ctx context.Context, req infrastructure.CreateRequest) (*infrastructure.Response, error) {
	endpoint := routes.CreateJobURL()
	var response infrastructure.Response
	if err := c.executeRequest(ctx, http.MethodPost, endpoint, req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetJob retrieves a job by ID
func (c *APIClient) GetJob(ctx context.Context, id string) (*infrastructure.JobStatus, error) {
	endpoint := routes.GetJobURL(id)
	var response infrastructure.JobStatus
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// ListJobs lists jobs with optional filtering
func (c *APIClient) ListJobs(ctx context.Context, limit int, status string) ([]infrastructure.JobStatus, error) {
	endpoint := routes.ListJobsURL()

	// Add query parameters
	if limit > 0 {
		endpoint += fmt.Sprintf("?limit=%d", limit)
		if status != "" {
			endpoint += fmt.Sprintf("&status=%s", status)
		}
	} else if status != "" {
		endpoint += fmt.Sprintf("?status=%s", status)
	}

	var response []infrastructure.JobStatus
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// Job instances methods implementation

// CreateJobInstance creates a new instance for a job
func (c *APIClient) CreateJobInstance(ctx context.Context, jobID string, req infrastructure.InstanceRequest) (*infrastructure.InstanceInfo, error) {
	endpoint := routes.CreateJobInstanceURL(jobID)
	var response infrastructure.InstanceInfo
	if err := c.executeRequest(ctx, http.MethodPost, endpoint, req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteJobInstance deletes an instance of a job
func (c *APIClient) DeleteJobInstance(ctx context.Context, jobID string, req infrastructure.DeleteInstanceRequest) (*infrastructure.Response, error) {
	endpoint := routes.DeleteJobInstanceURL(jobID)
	var response infrastructure.Response
	if err := c.executeRequest(ctx, http.MethodDelete, endpoint, req, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetJobInstance retrieves a specific instance of a job
func (c *APIClient) GetJobInstance(ctx context.Context, jobID string, instanceID string) (*infrastructure.InstanceInfo, error) {
	endpoint := routes.GetJobInstanceURL(jobID, instanceID)
	var response infrastructure.InstanceInfo
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetJobInstances retrieves instances for a job
func (c *APIClient) GetJobInstances(ctx context.Context, jobID string) ([]infrastructure.InstanceInfo, error) {
	endpoint := routes.GetJobInstancesURL(jobID)
	var response []infrastructure.InstanceInfo
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// GetJobPublicIPs retrieves public IPs for a job
func (c *APIClient) GetJobPublicIPs(ctx context.Context, jobID string) ([]string, error) {
	endpoint := routes.GetJobPublicIPsURL(jobID)
	var response []string
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// Instance methods implementation

// GetInstance retrieves an instance by ID
func (c *APIClient) GetInstance(ctx context.Context, id string) (*infrastructure.InstanceInfo, error) {
	endpoint := routes.GetInstanceURL(id)
	var response infrastructure.InstanceInfo
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// GetInstanceMetadata retrieves metadata for all instances
func (c *APIClient) GetInstanceMetadata(ctx context.Context) (map[string]interface{}, error) {
	endpoint := routes.GetInstanceMetadataURL()
	var response map[string]interface{}
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// ListInstances lists all instances
func (c *APIClient) ListInstances(ctx context.Context) ([]infrastructure.InstanceInfo, error) {
	endpoint := routes.ListInstancesURL()
	var response []infrastructure.InstanceInfo
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

// Health check implementation

// HealthCheck checks the health of the API
func (c *APIClient) HealthCheck(ctx context.Context) (map[string]string, error) {
	endpoint := routes.HealthCheckURL()
	var response map[string]string
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}
