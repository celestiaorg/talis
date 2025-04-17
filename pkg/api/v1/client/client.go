// Package client provides the API client for interacting with the Talis API
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/db/models"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/pkg/api/v1/routes"
)

// DefaultTimeout is the default timeout for API requests
const DefaultTimeout = 30 * time.Second

// Client is the interface for API client
type Client interface {
	// Admin Endpoints
	AdminGetInstances(ctx context.Context) ([]models.Instance, error)
	AdminGetInstancesMetadata(ctx context.Context) ([]models.Instance, error)

	// Health Check
	HealthCheck(ctx context.Context) (map[string]string, error)

	// Instance Endpoints
	GetInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error)
	GetInstancesMetadata(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error)
	GetInstancesPublicIPs(ctx context.Context, opts *models.ListOptions) (types.PublicIPsResponse, error)
	GetInstance(ctx context.Context, id string) (models.Instance, error)
	CreateInstance(ctx context.Context, req types.InstancesRequest) error
	DeleteInstance(ctx context.Context, req types.DeleteInstanceRequest) (types.TaskResponse, error)

	//User Endpoints
	GetUserByID(ctx context.Context, id string) (types.UserResponse, error)
	GetUsers(ctx context.Context, opts *models.UserQueryOptions) (types.UserResponse, error)
	CreateUser(ctx context.Context, req types.CreateUserRequest) (types.CreateUserResponse, error)
	DeleteUser(ctx context.Context, id string) error

	// Project methods
	CreateProject(ctx context.Context, params handlers.ProjectCreateParams) (models.Project, error)
	GetProject(ctx context.Context, params handlers.ProjectGetParams) (models.Project, error)
	ListProjects(ctx context.Context, params handlers.ProjectListParams) ([]models.Project, error)
	DeleteProject(ctx context.Context, params handlers.ProjectDeleteParams) error
	ListProjectInstances(ctx context.Context, params handlers.ProjectListInstancesParams) ([]models.Instance, error)

	// Task methods
	GetTask(ctx context.Context, params handlers.TaskGetParams) (models.Task, error)
	ListTasks(ctx context.Context, params handlers.TaskListParams) ([]models.Task, error)
	TerminateTask(ctx context.Context, params handlers.TaskTerminateParams) error
	UpdateTaskStatus(ctx context.Context, params handlers.TaskUpdateStatusParams) error
}

var _ Client = &APIClient{}

// Options contains configuration options for the API client
type Options struct {
	// BaseURL is the base URL of the API
	BaseURL string

	// Timeout is the request timeout
	Timeout time.Duration
}

// DefaultOptions returns the default client options
func DefaultOptions() *Options {
	return &Options{
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
func NewClient(opts *Options) (Client, error) {
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

	// Print the raw response for debugging
	// fmt.Printf("DEBUG - Response body: %s\n", string(body))

	// Check for non-success status codes
	if statusCode < 200 || statusCode >= 300 {
		// If we can't decode the error response, return an error with the raw body as the message
		return &fiber.Error{
			Code:    statusCode,
			Message: string(body),
		}
	}

	// Check if this is a SlugResponse
	if _, ok := v.(*types.TaskResponse); ok {
		var slugResponse types.SlugResponse
		if err := json.Unmarshal(body, &slugResponse); err != nil {
			return fmt.Errorf("error decoding slug response: %w", err)
		}

		// Extract the TaskResponse from the Data field
		if slugResponse.Data != nil {
			dataJSON, err := json.Marshal(slugResponse.Data)
			if err != nil {
				return fmt.Errorf("error marshaling data: %w", err)
			}

			return json.Unmarshal(dataJSON, v)
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

// RPCResponseWrapper is a wrapper for RPC responses
type RPCResponseWrapper struct {
	Data    json.RawMessage `json:"data"`
	Error   interface{}     `json:"error,omitempty"`
	ID      string          `json:"id,omitempty"`
	Success bool            `json:"success"`
}

// rpcRequest executes an RPC-style request to the API
func (c *APIClient) rpcRequest(ctx context.Context, method string, params interface{}, response interface{}) error {
	endpoint := routes.RPCURL()

	// Create the request body
	requestBody := map[string]interface{}{
		"method": method,
		"params": params,
	}

	// Create the agent
	agent, err := c.createAgent(ctx, http.MethodPost, endpoint, requestBody)
	if err != nil {
		return err
	}

	// Execute the request and get the response body
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return fmt.Errorf("error sending RPC request: %w", errs[0])
	}

	// Check for non-success status codes
	if statusCode < 200 || statusCode >= 300 {
		return &fiber.Error{
			Code:    statusCode,
			Message: string(body), // Raw body as error message
		}
	}

	// Unmarshal the response into the wrapper
	var wrapper RPCResponseWrapper
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return fmt.Errorf("error decoding RPC response wrapper: %w. Body: %s", err, string(body))
	}

	// Check for application-level errors in the response
	if !wrapper.Success {
		if wrapper.Error != nil {
			errBytes, _ := json.Marshal(wrapper.Error)
			return fmt.Errorf("RPC request failed: %s", string(errBytes))
		}
		return fmt.Errorf("RPC request failed with no error details")
	}

	// Unmarshal the actual data from the wrapper into the target response object
	if response != nil && wrapper.Data != nil {
		if err := json.Unmarshal(wrapper.Data, response); err != nil {
			return fmt.Errorf("error decoding RPC data: %w. Data: %s", err, string(wrapper.Data))
		}
	}

	return nil
}

// Admin methods implementation

// AdminGetInstances retrieves all instances
func (c *APIClient) AdminGetInstances(ctx context.Context) ([]models.Instance, error) {
	endpoint := routes.AdminInstancesURL()
	var response types.ListResponse[models.Instance]
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return []models.Instance{}, err
	}
	return response.Rows, nil
}

// AdminGetInstancesMetadata retrieves metadata for all instances
func (c *APIClient) AdminGetInstancesMetadata(ctx context.Context) ([]models.Instance, error) {
	endpoint := routes.AdminInstancesMetadataURL()
	var response types.ListResponse[models.Instance]
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return []models.Instance{}, err
	}
	return response.Rows, nil
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

	// Pagination params
	if opts.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}

	// Filtering params
	if opts.IncludeDeleted {
		q.Set("include_deleted", "true")
	}

	if opts.StatusFilter != "" {
		q.Set("status_filter", string(opts.StatusFilter))
	}

	// Instance status params
	if opts.InstanceStatus != nil {
		status := *opts.InstanceStatus
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
		q.Set("instance_status", statusStr)
	}

	return q, nil
}

// GetInstances lists instances with optional filtering
func (c *APIClient) GetInstances(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return []models.Instance{}, err
	}

	endpoint := routes.GetInstancesURL(q)
	var response types.ListResponse[models.Instance]
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return []models.Instance{}, err
	}
	return response.Rows, nil
}

// GetInstancesMetadata retrieves metadata for all instances
func (c *APIClient) GetInstancesMetadata(ctx context.Context, opts *models.ListOptions) ([]models.Instance, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return []models.Instance{}, err
	}

	endpoint := routes.GetInstanceMetadataURL(q)
	var response types.ListResponse[models.Instance]
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return []models.Instance{}, err
	}
	return response.Rows, nil
}

// GetInstancesPublicIPs retrieves public IPs for all instances
func (c *APIClient) GetInstancesPublicIPs(ctx context.Context, opts *models.ListOptions) (types.PublicIPsResponse, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return types.PublicIPsResponse{}, err
	}

	endpoint := routes.GetPublicIPsURL(q)
	var response types.PublicIPsResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return types.PublicIPsResponse{}, err
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
func (c *APIClient) CreateInstance(ctx context.Context, req types.InstancesRequest) error {
	endpoint := routes.CreateInstanceURL()
	return c.executeRequest(ctx, http.MethodPost, endpoint, req, nil)
}

// DeleteInstance deletes an instance by ID
func (c *APIClient) DeleteInstance(ctx context.Context, req types.DeleteInstanceRequest) (types.TaskResponse, error) {
	endpoint := routes.TerminateInstancesURL()
	var response types.TaskResponse
	if err := c.executeRequest(ctx, http.MethodDelete, endpoint, req, &response); err != nil {
		return types.TaskResponse{}, err
	}
	return response, nil
}

// User method implementation

// GetUserByID retrieves a user by id
func (c *APIClient) GetUserByID(ctx context.Context, id string) (types.UserResponse, error) {
	endpoint := routes.GetUserByIDURL(id)
	var response types.UserResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return types.UserResponse{}, err
	}
	return response, nil
}

// GetUsers retrieves a user by username
func (c *APIClient) GetUsers(ctx context.Context, opts *models.UserQueryOptions) (types.UserResponse, error) {
	q, err := getUsersQueryParams(opts)
	if err != nil {
		return types.UserResponse{}, err
	}
	endpoint := routes.GetUsersURL(q)
	var response types.UserResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return types.UserResponse{}, err
	}
	return response, nil
}

// CreateUser creates a new user
func (c *APIClient) CreateUser(ctx context.Context, req types.CreateUserRequest) (types.CreateUserResponse, error) {
	var response types.CreateUserResponse
	endpoint := routes.CreateUserURL()
	if err := c.executeRequest(ctx, http.MethodPost, endpoint, req, &response); err != nil {
		return types.CreateUserResponse{}, err
	}
	return response, nil
}

// DeleteUser user deletes a user
func (c *APIClient) DeleteUser(ctx context.Context, id string) error {
	endpoint := routes.DeleteUserURL(id)
	return c.executeRequest(ctx, http.MethodDelete, endpoint, nil, nil)
}

// CreateProject creates a new project
func (c *APIClient) CreateProject(ctx context.Context, params handlers.ProjectCreateParams) (models.Project, error) {
	var project models.Project
	if err := c.rpcRequest(ctx, handlers.ProjectCreate, params, &project); err != nil {
		return project, err
	}
	return project, nil
}

// GetProject retrieves a project by name
func (c *APIClient) GetProject(ctx context.Context, params handlers.ProjectGetParams) (models.Project, error) {
	var project models.Project
	if err := c.rpcRequest(ctx, handlers.ProjectGet, params, &project); err != nil {
		return models.Project{}, err
	}
	return project, nil
}

// ListProjects lists all projects
func (c *APIClient) ListProjects(ctx context.Context, params handlers.ProjectListParams) ([]models.Project, error) {
	var listResponse types.ListResponse[models.Project]
	if err := c.rpcRequest(ctx, handlers.ProjectList, params, &listResponse); err != nil {
		return nil, err
	}
	return listResponse.Rows, nil
}

// DeleteProject deletes a project by name
func (c *APIClient) DeleteProject(ctx context.Context, params handlers.ProjectDeleteParams) error {
	return c.rpcRequest(ctx, handlers.ProjectDelete, params, nil)
}

// ListProjectInstances lists all instances for a project
func (c *APIClient) ListProjectInstances(ctx context.Context, params handlers.ProjectListInstancesParams) ([]models.Instance, error) {
	var listResponse types.ListResponse[models.Instance]
	if err := c.rpcRequest(ctx, handlers.ProjectListInstances, params, &listResponse); err != nil {
		return nil, err
	}
	return listResponse.Rows, nil
}

// Task methods implementation

// GetTask retrieves a task by name
func (c *APIClient) GetTask(ctx context.Context, params handlers.TaskGetParams) (models.Task, error) {
	var task models.Task
	if err := c.rpcRequest(ctx, handlers.TaskGet, params, &task); err != nil {
		return models.Task{}, err
	}
	return task, nil
}

// ListTasks lists all tasks
func (c *APIClient) ListTasks(ctx context.Context, params handlers.TaskListParams) ([]models.Task, error) {
	var listResponse types.ListResponse[models.Task]
	if err := c.rpcRequest(ctx, handlers.TaskList, params, &listResponse); err != nil {
		return nil, err
	}
	return listResponse.Rows, nil
}

// TerminateTask terminates a task by name
func (c *APIClient) TerminateTask(ctx context.Context, params handlers.TaskTerminateParams) error {
	return c.rpcRequest(ctx, handlers.TaskTerminate, params, nil)
}

// UpdateTaskStatus updates the status of a task
func (c *APIClient) UpdateTaskStatus(ctx context.Context, params handlers.TaskUpdateStatusParams) error {
	return c.rpcRequest(ctx, handlers.TaskUpdateStatus, params, nil)
}
