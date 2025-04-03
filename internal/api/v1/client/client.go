package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
)

// DefaultTimeout is the default timeout for API requests
const DefaultTimeout = 30 * time.Second

// Client defines the interface for interacting with the Talis API
type Client interface {
	// Admin Endpoints
	AdminGetInstances(ctx context.Context) (infrastructure.ListInstancesResponse, error)
	AdminGetInstancesMetadata(ctx context.Context) (infrastructure.InstanceMetadataResponse, error)

	// Health Check
	HealthCheck(ctx context.Context) (map[string]string, error)

	// Instance Endpoints
	GetInstances(ctx context.Context, opts *models.ListOptions) (infrastructure.ListInstancesResponse, error)
	GetInstancesMetadata(ctx context.Context, opts *models.ListOptions) (infrastructure.InstanceMetadataResponse, error)
	GetInstancesPublicIPs(ctx context.Context, opts *models.ListOptions) (infrastructure.PublicIPsResponse, error)
	GetInstance(ctx context.Context, id string) (models.Instance, error)
	CreateInstance(ctx context.Context, req infrastructure.InstancesRequest) error
	DeleteInstance(ctx context.Context, req infrastructure.DeleteInstanceRequest) error

	// Jobs Endpoints
	GetJobs(ctx context.Context, opts *models.ListOptions) (infrastructure.ListJobsResponse, error)
	GetJob(ctx context.Context, id string) (models.Job, error)
	GetMetadataByJobID(ctx context.Context, id string, opts *models.ListOptions) (infrastructure.InstanceMetadataResponse, error)
	GetInstancesByJobID(ctx context.Context, id string, opts *models.ListOptions) (infrastructure.JobInstancesResponse, error)
	GetJobStatus(ctx context.Context, id string) (models.JobStatus, error)
	CreateJob(ctx context.Context, req infrastructure.JobRequest) error
	UpdateJob(ctx context.Context, id string, req infrastructure.JobRequest) error
	DeleteJob(ctx context.Context, id string) error

	//User Endpoints
	GetUserByID(ctx context.Context, id string) (models.User, error)
	GetUsers(ctx context.Context, opts *models.UserQueryOptions) (infrastructure.UserResponse, error)
	CreateUser(ctx context.Context, req infrastructure.CreateUserRequest) (infrastructure.CreateUserResponse, error)
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

// doRequest sends the HTTP request and processes the response
func (c *APIClient) doRequest(agent *fiber.Agent, v interface{}) error {
	// Execute the request
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return fmt.Errorf("error sending request: %w", errs[0])
	}

	// Check for non-success status codes
	if statusCode < 200 || statusCode >= 300 {
		// If we can't decode the error response, return an error with the raw body as the message
		return &fiber.Error{
			Code:    statusCode,
			Message: string(body),
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

// executeRequest creates an agent, sends the request, and processes the response
func (c *APIClient) executeRequest(ctx context.Context, method, endpoint string, body, response interface{}) error {
	agent, err := c.createAgent(ctx, method, endpoint, body)
	if err != nil {
		return err
	}

	return c.doRequest(agent, response)
}

// Admin methods implementation

// AdminGetInstances retrieves all instances
func (c *APIClient) AdminGetInstances(ctx context.Context) (infrastructure.ListInstancesResponse, error) {
	endpoint := routes.AdminInstancesURL()
	var response infrastructure.ListInstancesResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.ListInstancesResponse{}, err
	}
	return response, nil
}

// AdminGetInstancesMetadata retrieves metadata for all instances
func (c *APIClient) AdminGetInstancesMetadata(ctx context.Context) (infrastructure.InstanceMetadataResponse, error) {
	endpoint := routes.AdminInstancesMetadataURL()
	var response infrastructure.InstanceMetadataResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.InstanceMetadataResponse{}, err
	}
	return response, nil
}

// Health check implementation

// HealthCheck checks the health of the API
func (c *APIClient) HealthCheck(ctx context.Context) (map[string]string, error) {
	endpoint := routes.HealthCheckURL()
	var response map[string]string
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return map[string]string{}, err
	}
	return response, nil
}

// Instance methods implementation

func getUsersQueryParams(opts *models.UserQueryOptions) (url.Values, error) {
	q := url.Values{}
	if opts == nil {
		return q, nil
	}
	if opts.Username != "" {
		q.Set("username", opts.Username)
	}
	return q, nil
}

// getQueryParams creates url.Values from ListOptions
func getQueryParams(opts *models.ListOptions) (url.Values, error) {
	q := url.Values{}
	if opts == nil {
		return q, nil
	}

	if opts.IncludeDeleted {
		q.Set("include_deleted", "true")
	}
	if opts.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}
	if opts.Status != nil {
		status := *opts.Status
		var statusStr string
		switch status {
		case models.InstanceStatusUnknown:
			statusStr = "unknown"
		case models.InstanceStatusPending:
			statusStr = "pending"
		case models.InstanceStatusProvisioning:
			statusStr = "provisioning"
		case models.InstanceStatusReady:
			statusStr = "ready"
		case models.InstanceStatusTerminated:
			statusStr = "terminated"
		default:
			return nil, fmt.Errorf("invalid instance status: %d", status)
		}
		q.Set("status", statusStr)
	}

	// Add status_filter if provided, regardless of status value
	if opts.StatusFilter != "" {
		q.Set("status_filter", string(opts.StatusFilter))
	}

	return q, nil
}

// GetInstances lists instances with optional filtering
func (c *APIClient) GetInstances(ctx context.Context, opts *models.ListOptions) (infrastructure.ListInstancesResponse, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return infrastructure.ListInstancesResponse{}, err
	}

	endpoint := routes.GetInstancesURL(q)
	var response infrastructure.ListInstancesResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.ListInstancesResponse{}, err
	}
	return response, nil
}

// GetInstancesMetadata retrieves metadata for all instances
func (c *APIClient) GetInstancesMetadata(ctx context.Context, opts *models.ListOptions) (infrastructure.InstanceMetadataResponse, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return infrastructure.InstanceMetadataResponse{}, err
	}

	endpoint := routes.GetInstanceMetadataURL(q)
	var response infrastructure.InstanceMetadataResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.InstanceMetadataResponse{}, err
	}
	return response, nil
}

// GetInstancesPublicIPs retrieves public IPs for all instances
func (c *APIClient) GetInstancesPublicIPs(ctx context.Context, opts *models.ListOptions) (infrastructure.PublicIPsResponse, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return infrastructure.PublicIPsResponse{}, err
	}

	endpoint := routes.GetPublicIPsURL(q)
	var response infrastructure.PublicIPsResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.PublicIPsResponse{}, err
	}
	return response, nil
}

// GetInstance retrieves an instance by ID
func (c *APIClient) GetInstance(ctx context.Context, id string) (models.Instance, error) {
	endpoint := routes.GetInstanceURL(id)
	var response models.Instance
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return models.Instance{}, err
	}
	return response, nil
}

// CreateInstance creates a new instance
func (c *APIClient) CreateInstance(ctx context.Context, req infrastructure.InstancesRequest) error {
	endpoint := routes.CreateInstanceURL()
	return c.executeRequest(ctx, http.MethodPost, endpoint, req, nil)
}

// DeleteInstance deletes an instance by ID
func (c *APIClient) DeleteInstance(ctx context.Context, req infrastructure.DeleteInstanceRequest) error {
	endpoint := routes.TerminateInstancesURL()
	return c.executeRequest(ctx, http.MethodDelete, endpoint, req, nil)
}

// Job methods implementation

// GetJobs lists jobs with optional filtering
func (c *APIClient) GetJobs(ctx context.Context, opts *models.ListOptions) (infrastructure.ListJobsResponse, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return infrastructure.ListJobsResponse{}, err
	}

	endpoint := routes.GetJobsURL(q)
	var response infrastructure.ListJobsResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.ListJobsResponse{}, err
	}
	return response, nil
}

// GetJob retrieves a job by ID
func (c *APIClient) GetJob(ctx context.Context, id string) (models.Job, error) {
	endpoint := routes.GetJobURL(id)
	var response models.Job
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return models.Job{}, err
	}
	return response, nil
}

// GetMetadataByJobID retrieves metadata for a job by ID
func (c *APIClient) GetMetadataByJobID(ctx context.Context, id string, opts *models.ListOptions) (infrastructure.InstanceMetadataResponse, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return infrastructure.InstanceMetadataResponse{}, err
	}

	endpoint := routes.GetJobMetadataURL(id, q)
	var response infrastructure.InstanceMetadataResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.InstanceMetadataResponse{}, err
	}
	return response, nil
}

// GetInstancesByJobID retrieves instances for a job by ID
func (c *APIClient) GetInstancesByJobID(ctx context.Context, id string, opts *models.ListOptions) (infrastructure.JobInstancesResponse, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return infrastructure.JobInstancesResponse{}, err
	}

	endpoint := routes.GetJobInstancesURL(id, q)
	var response infrastructure.JobInstancesResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.JobInstancesResponse{}, err
	}
	return response, nil
}

// GetJobStatus retrieves the status of a job by ID
func (c *APIClient) GetJobStatus(ctx context.Context, id string) (models.JobStatus, error) {
	endpoint := routes.GetJobStatusURL(id)
	var response models.JobStatus
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return "", err
	}
	return response, nil
}

// CreateJob creates a new job
func (c *APIClient) CreateJob(ctx context.Context, req infrastructure.JobRequest) error {
	endpoint := routes.CreateJobURL()
	return c.executeRequest(ctx, http.MethodPost, endpoint, req, nil)
}

// UpdateJob updates a job by ID
func (c *APIClient) UpdateJob(ctx context.Context, id string, req infrastructure.JobRequest) error {
	endpoint := routes.UpdateJobURL(id)
	return c.executeRequest(ctx, http.MethodPut, endpoint, req, nil)
}

// DeleteJob deletes a job by ID
func (c *APIClient) DeleteJob(ctx context.Context, id string) error {
	endpoint := routes.DeleteJobURL(id)
	return c.executeRequest(ctx, http.MethodDelete, endpoint, nil, nil)
}

// User method implementation

// GetUserByID retrieves a user by id
func (c *APIClient) GetUserByID(ctx context.Context, id string) (models.User, error) {
	endpoint := routes.GetUserByIDURL(id)
	var response models.User
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return models.User{}, err
	}
	return response, nil
}

// GetUsers retrieves a user by username
func (c *APIClient) GetUsers(ctx context.Context, opts *models.UserQueryOptions) (infrastructure.UserResponse, error) {
	q, err := getUsersQueryParams(opts)
	if err != nil {
		return infrastructure.UserResponse{}, err
	}
	endpoint := routes.GetUsersURL(q)
	var response infrastructure.UserResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return infrastructure.UserResponse{}, err
	}
	return response, nil
}

// CreateUser creates a new user
func (c *APIClient) CreateUser(ctx context.Context, req infrastructure.CreateUserRequest) (infrastructure.CreateUserResponse, error) {
	var response infrastructure.CreateUserResponse
	endpoint := routes.CreateUserURL()
	if err := c.executeRequest(ctx, http.MethodPost, endpoint, req, &response); err != nil {
		return infrastructure.CreateUserResponse{}, err
	}
	return response, nil
}
