// Package client provides a comprehensive API client for interacting with the Talis API.
//
// This package offers a clean, idiomatic Go interface for all Talis API operations,
// including instance management, user administration, project operations, and task handling.
// The client handles authentication, request formatting, response parsing, and error management,
// allowing developers to focus on their application logic rather than API communication details.
//
// Basic usage:
//
//	// Create a client with default options
//	client, err := client.NewClient(nil)
//	if err != nil {
//	    log.Fatalf("Failed to create client: %v", err)
//	}
//
//	// Set API key if needed
//	client.SetAPIKey("your-api-key")
//
//	// Use the client to interact with the API
//	instances, err := client.GetInstances(context.Background(), nil)
//
// For more detailed examples, refer to the client_usage.md documentation.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	internalmodels "github.com/celestiaorg/talis/internal/db/models" // Import internal models
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/pkg/api/v1/routes"
	"github.com/celestiaorg/talis/pkg/db/models" // Import public models alias
	"github.com/celestiaorg/talis/pkg/types"     // Import public types alias
)

// DefaultTimeout is the default timeout for API requests
const DefaultTimeout = 30 * time.Second

// Client is the interface for the Talis API client. It provides methods for interacting
// with all aspects of the Talis API, organized into logical categories:
// - Admin operations (for administrative access)
// - Health checks (for monitoring API status)
// - Instance management (creating, listing, and deleting compute instances)
// - User management (creating, retrieving, and deleting users)
// - Project management (creating, retrieving, and deleting projects)
// - Task management (retrieving and managing long-running tasks)
// - SSH Key management (creating, listing, and deleting SSH keys)
//
// All methods accept a context.Context as their first parameter to support
// timeout and cancellation. Most methods return structured data and an error.
// A nil error indicates success.
type Client interface {
	// Admin Endpoints - These methods require administrative privileges

	// AdminGetInstances retrieves all instances across all projects.
	// This is an administrative endpoint that returns all instances regardless of owner.
	// Returns a slice of Instance pointers and any error encountered.
	AdminGetInstances(ctx context.Context) ([]*models.Instance, error)

	// AdminGetInstancesMetadata retrieves metadata for all instances across all projects.
	// This is an administrative endpoint that returns lightweight instance information.
	// Returns a slice of Instance pointers containing only metadata fields and any error encountered.
	AdminGetInstancesMetadata(ctx context.Context) ([]*models.Instance, error)

	// Health Check

	// HealthCheck performs a health check against the API.
	// Returns a map of component names to their status and any error encountered.
	// A successful response typically includes entries like {"status": "ok"}.
	HealthCheck(ctx context.Context) (map[string]string, error)

	// Instance Endpoints - Methods for managing compute instances

	// GetInstances retrieves instances with optional filtering via ListOptions.
	// The opts parameter can include pagination (limit/offset), status filtering,
	// and whether to include deleted instances.
	// Returns a slice of Instance pointers and any error encountered.
	GetInstances(ctx context.Context, opts *models.ListOptions) ([]*models.Instance, error)

	// GetInstancesMetadata retrieves metadata for instances with optional filtering.
	// Similar to GetInstances but returns only essential metadata fields for efficiency.
	// Returns a slice of Instance pointers with metadata fields and any error encountered.
	GetInstancesMetadata(ctx context.Context, opts *models.ListOptions) ([]*models.Instance, error)

	// GetInstancesPublicIPs retrieves public IP addresses for instances with optional filtering.
	// Returns a PublicIPsResponse containing a map of instance IDs to their public IPs,
	// and any error encountered.
	GetInstancesPublicIPs(ctx context.Context, opts *models.ListOptions) (types.PublicIPsResponse, error)

	// GetInstance retrieves a specific instance by its ID.
	// Returns the Instance and any error encountered.
	GetInstance(ctx context.Context, id string) (models.Instance, error)

	// CreateInstance creates new instances based on the provided specifications.
	// The req parameter is a slice of InstanceRequest objects, each describing
	// an instance to be created.
	// Returns a slice of the created Instance pointers and any error encountered.
	CreateInstance(ctx context.Context, req []types.InstanceRequest) ([]*models.Instance, error)

	// DeleteInstances terminates the specified instances for a project.
	// The req parameter contains the project name and instance IDs to delete.
	// Returns an error if the operation fails.
	DeleteInstances(ctx context.Context, req types.DeleteInstancesRequest) error

	// User Endpoints - Methods for managing users

	// GetUserByID retrieves a user by their ID.
	// Returns the User and any error encountered.
	GetUserByID(ctx context.Context, params handlers.UserGetByIDParams) (models.User, error)

	// GetUsers retrieves users based on the provided parameters.
	// Returns a UserResponse containing user information and any error encountered.
	GetUsers(ctx context.Context, params handlers.UserGetParams) (types.UserResponse, error)

	// CreateUser creates a new user with the provided parameters.
	// Returns a CreateUserResponse containing the created user information and any error encountered.
	CreateUser(ctx context.Context, params handlers.CreateUserParams) (types.CreateUserResponse, error)

	// DeleteUser deletes a user based on the provided parameters.
	// Returns an error if the operation fails.
	DeleteUser(ctx context.Context, params handlers.DeleteUserParams) error

	// Project methods - Methods for managing projects

	// CreateProject creates a new project with the provided parameters.
	// Returns the created Project and any error encountered.
	CreateProject(ctx context.Context, params handlers.ProjectCreateParams) (models.Project, error)

	// GetProject retrieves a project by name.
	// Returns the Project and any error encountered.
	GetProject(ctx context.Context, params handlers.ProjectGetParams) (models.Project, error)

	// ListProjects lists all projects based on the provided parameters.
	// Returns a slice of Project pointers and any error encountered.
	ListProjects(ctx context.Context, params handlers.ProjectListParams) ([]*models.Project, error)

	// DeleteProject deletes a project by name.
	// Returns an error if the operation fails.
	DeleteProject(ctx context.Context, params handlers.ProjectDeleteParams) error

	// ListProjectInstances lists all instances for a specific project.
	// Returns a slice of Instance pointers and any error encountered.
	ListProjectInstances(ctx context.Context, params handlers.ProjectListInstancesParams) ([]*models.Instance, error)

	// Task methods - Methods for managing tasks

	// GetTask retrieves a task by its identifier.
	// Returns the Task and any error encountered.
	GetTask(ctx context.Context, params handlers.TaskGetParams) (models.Task, error)

	// ListTasks lists all tasks based on the provided parameters.
	// Returns a slice of Task pointers and any error encountered.
	ListTasks(ctx context.Context, params handlers.TaskListParams) ([]*models.Task, error)

	// ListTasksByInstanceID retrieves tasks for a specific instance ID.
	// Parameters:
	// - ownerID: The ID of the user who owns the instance
	// - instanceID: The ID of the instance to list tasks for
	// - actionFilter: Optional filter for task action type (e.g., "create_instances")
	// - opts: Optional pagination parameters
	// Returns a slice of Task pointers and any error encountered.
	ListTasksByInstanceID(ctx context.Context, ownerID uint, instanceID uint, actionFilter string, opts *models.ListOptions) ([]*models.Task, error)

	// TerminateTask terminates a running task.
	// Returns an error if the operation fails.
	TerminateTask(ctx context.Context, params handlers.TaskTerminateParams) error

	// UpdateTaskStatus updates the status of a task.
	// Returns an error if the operation fails.
	UpdateTaskStatus(ctx context.Context, params handlers.TaskUpdateStatusParams) error

	// SSH Key methods - Methods for managing SSH keys

	// CreateSSHKey creates a new SSH key.
	// Returns the created SSH key and any error encountered.
	CreateSSHKey(ctx context.Context, params handlers.SSHKeyCreateParams) (internalmodels.SSHKey, error)

	// ListSSHKeys lists all SSH keys for a specific owner.
	// Returns a slice of SSH key pointers and any error encountered.
	ListSSHKeys(ctx context.Context, params handlers.SSHKeyListParams) ([]*internalmodels.SSHKey, error)

	// DeleteSSHKey deletes an SSH key.
	// Returns an error if the operation fails.
	DeleteSSHKey(ctx context.Context, params handlers.SSHKeyDeleteParams) error
}

var _ Client = &APIClient{}

// Options contains configuration options for the API client.
// These options control the client's behavior when communicating with the API.
type Options struct {
	// BaseURL is the base URL of the API.
	// This should include the protocol (http:// or https://) and host,
	// but not the endpoint paths (e.g., "https://api.example.com").
	BaseURL string

	// APIKey is the API key for authentication.
	// If provided, this key will be included in all API requests.
	APIKey string

	// Timeout is the default request timeout.
	// This value is used when no deadline is set in the context.
	// If not specified, DefaultTimeout (30 seconds) is used.
	Timeout time.Duration
}

// DefaultOptions returns the default client options.
// The defaults are:
// - BaseURL: The value from routes.DefaultBaseURL
// - APIKey: Empty string (no authentication)
// - Timeout: DefaultTimeout (30 seconds)
//
// These defaults are suitable for local development but should be
// customized for production use.
func DefaultOptions() *Options {
	return &Options{
		BaseURL: routes.DefaultBaseURL,
		APIKey:  "",
		Timeout: DefaultTimeout,
	}
}

// APIClient implements the Client interface.
// It handles the actual HTTP communication with the Talis API,
// including request formatting, authentication, and response parsing.
type APIClient struct {
	// baseURL is the base URL for all API requests
	baseURL string

	// timeout is the default request timeout
	timeout time.Duration

	// AuthToken is a JWT token for authentication (if used)
	AuthToken string

	// APIKey is the API key for authentication
	APIKey string
}

// NewClient creates a new API client with the given options.
//
// If opts is nil, default options are used (see DefaultOptions).
// The function validates the base URL and returns an error if it's invalid.
//
// Example:
//
//	// Create client with default options
//	client, err := client.NewClient(nil)
//
//	// Create client with custom options
//	client, err := client.NewClient(&client.Options{
//	    BaseURL: "https://api.example.com",
//	    APIKey: "your-api-key",
//	    Timeout: 60 * time.Second,
//	})
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
		APIKey:  opts.APIKey,
		timeout: opts.Timeout,
	}, nil
}

// SetAPIKey sets the API key for the client.
// This method can be used to update the API key after client creation.
func (c *APIClient) SetAPIKey(apiKey string) {
	c.APIKey = apiKey
}

// createAgent creates a new Fiber Agent for the given HTTP method and endpoint.
//
// This is an internal helper method that prepares an HTTP request with the appropriate
// headers, timeout, and body. It handles:
// - Constructing the full URL from the base URL and endpoint
// - Creating the appropriate HTTP method agent (GET, POST, etc.)
// - Setting the timeout based on context deadline or client default
// - Adding common headers (Content-Type, Accept)
// - Adding authentication headers if an API key is set
// - Adding the request body as JSON if provided
//
// Parameters:
// - ctx: Context for timeout and cancellation
// - method: HTTP method (GET, POST, PUT, DELETE, PATCH)
// - endpoint: API endpoint path (without base URL)
// - body: Optional request body to be serialized as JSON
//
// Returns a configured Fiber Agent or an error if the HTTP method is unsupported.
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

	// Add API key header if set
	if c.APIKey != "" {
		agent.Set("apikey", c.APIKey)
	}

	// Add body if provided
	if body != nil {
		agent.JSON(body)
	}

	return agent, nil
}

// doRequest sends the HTTP request and processes the response.
//
// This internal helper method executes the HTTP request using the provided Fiber Agent
// and handles the response processing, including:
// - Checking for network errors
// - Handling non-success HTTP status codes
// - Special handling for SlugResponse types
// - JSON unmarshaling of the response body into the provided target
//
// Parameters:
// - agent: The configured Fiber Agent ready to execute the request
// - v: Optional target to unmarshal the response into (can be nil for requests without response bodies)
//
// Returns an error if the request fails, the status code indicates an error,
// or if the response cannot be properly decoded.
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

// executeRequest creates an agent, sends the request, and processes the response.
//
// This is a convenience method that combines createAgent and doRequest into a single call.
// It handles the complete HTTP request lifecycle:
// - Creating and configuring the HTTP agent
// - Sending the request
// - Processing the response
//
// Parameters:
// - ctx: Context for timeout and cancellation
// - method: HTTP method (GET, POST, PUT, DELETE, PATCH)
// - endpoint: API endpoint path (without base URL)
// - body: Optional request body to be serialized as JSON
// - response: Optional target to unmarshal the response into
//
// Returns an error if any part of the request process fails.
func (c *APIClient) executeRequest(ctx context.Context, method, endpoint string, body, response interface{}) error {
	agent, err := c.createAgent(ctx, method, endpoint, body)
	if err != nil {
		return err
	}

	return c.doRequest(agent, response)
}

// executeRPC performs an RPC call to the Talis API.
//
// This method handles the complete RPC request lifecycle, including:
// - Constructing the RPC request with method name and parameters
// - Sending the request to the RPC endpoint
// - Processing the response and checking for application-level errors
// - Unmarshaling the response data into the provided result
//
// The Talis API uses a custom RPC protocol where:
// - All RPC calls are made to a single endpoint
// - The method name and parameters are sent in the request body
// - Responses include success/error flags and a data field
//
// Parameters:
// - ctx: Context for timeout and cancellation
// - method: The RPC method name to call (e.g., "user.get")
// - params: The parameters for the RPC method (will be serialized to JSON)
// - result: Optional target to unmarshal the response data into (can be nil for methods without return values)
//
// Returns an error if the request fails, the RPC call returns an error,
// or if the response data cannot be properly decoded.
func (c *APIClient) executeRPC(ctx context.Context, method string, params interface{}, result interface{}) error {
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

	// Unmarshal the response into the handlers.RPCResponse struct
	var rpcResp handlers.RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return fmt.Errorf("failed to unmarshal RPC response body: %w", err)
	}

	// Check for application-level errors
	if rpcResp.Error != nil {
		return fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	if !rpcResp.Success {
		return fmt.Errorf("RPC call failed without specific error details")
	}

	// If result is nil, we don't need to unmarshal data (e.g., for notification-style calls)
	if result == nil {
		return nil
	}

	// Unmarshal the Data field into the provided result interface{}
	// Since rpcResp.Data is interface{}, we need to marshal it back to JSON
	// and then unmarshal it into the target result struct.
	dataBytes, err := json.Marshal(rpcResp.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal RPC data field: %w", err)
	}

	if err := json.Unmarshal(dataBytes, result); err != nil {
		return fmt.Errorf("failed to unmarshal RPC data into result: %w", err)
	}

	return nil
}

// Admin methods implementation

// AdminGetInstances retrieves all instances
func (c *APIClient) AdminGetInstances(ctx context.Context) ([]*models.Instance, error) {
	endpoint := routes.AdminInstancesURL()
	var response types.ListResponse[models.Instance] // Use pkg/types
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return []*models.Instance{}, err
	}
	return response.Rows, nil
}

// AdminGetInstancesMetadata retrieves metadata for all instances
func (c *APIClient) AdminGetInstancesMetadata(ctx context.Context) ([]*models.Instance, error) {
	endpoint := routes.AdminInstancesMetadataURL()
	var response types.ListResponse[models.Instance] // Use pkg/types
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return []*models.Instance{}, err
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

// getQueryParams converts ListOptions into URL query parameters.
//
// This helper function transforms the structured ListOptions object into
// URL query parameters (url.Values) for use in API requests. It handles:
// - Pagination parameters (limit, offset)
// - Filtering options (include_deleted)
// - Status filtering (status_filter)
// - Instance status filtering (instance_status)
//
// The function properly converts each option to its string representation
// and handles special cases like the InstanceStatus enum conversion.
//
// Parameters:
//   - opts: A pointer to ListOptions containing filtering and pagination options.
//     If nil, an empty url.Values is returned.
//
// Returns:
// - url.Values: The constructed query parameters
// - error: An error if the conversion fails, particularly for invalid enum values
func getQueryParams(opts *models.ListOptions) (url.Values, error) {
	q := url.Values{}
	if opts == nil {
		return q, nil
	}

	// Pagination params
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", strconv.Itoa(opts.Offset))
	}

	// Filtering params
	if opts.IncludeDeleted {
		q.Set("include_deleted", "true")
	}

	// StatusFilter is string-based (public alias)
	if opts.StatusFilter != "" {
		q.Set("status_filter", string(opts.StatusFilter))
	}

	// InstanceStatus is pointer-based in the underlying internal struct
	if opts.InstanceStatus != nil { // Check if the pointer is non-nil
		status := *opts.InstanceStatus // Dereference to get the value
		var statusStr string
		switch status {
		case models.InstanceStatusUnknown:
			statusStr = "unknown"
		case models.InstanceStatusPending:
			statusStr = "pending"
		case models.InstanceStatusProvisioning:
			statusStr = "provisioning"
		case models.InstanceStatusCreated:
			statusStr = "created"
		case models.InstanceStatusReady:
			statusStr = "ready"
		case models.InstanceStatusTerminated:
			statusStr = "terminated"
		default:
			// Use %v for the underlying int type
			return nil, fmt.Errorf("invalid instance status: %v", status)
		}
		q.Set("instance_status", statusStr)
	}

	return q, nil
}

// GetInstances lists instances with optional filtering
func (c *APIClient) GetInstances(ctx context.Context, opts *models.ListOptions) ([]*models.Instance, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return []*models.Instance{}, err
	}

	endpoint := routes.GetInstancesURL(q)
	var response types.ListResponse[models.Instance] // Use pkg/types
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return []*models.Instance{}, err
	}
	return response.Rows, nil
}

// GetInstancesMetadata retrieves metadata for all instances
func (c *APIClient) GetInstancesMetadata(ctx context.Context, opts *models.ListOptions) ([]*models.Instance, error) {
	q, err := getQueryParams(opts)
	if err != nil {
		return []*models.Instance{}, err
	}

	endpoint := routes.GetInstanceMetadataURL(q)
	var response types.ListResponse[models.Instance] // Use pkg/types
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return []*models.Instance{}, err
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
	var response models.Instance // Use pkg/models.Instance
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return models.Instance{}, err
	}
	return response, nil
}

// CreateInstance creates new instances
func (c *APIClient) CreateInstance(ctx context.Context, req []types.InstanceRequest) ([]*models.Instance, error) {
	endpoint := routes.CreateInstanceURL()
	var slugResp types.SlugResponse

	if err := c.executeRequest(ctx, fiber.MethodPost, endpoint, req, &slugResp); err != nil {
		return nil, err
	}

	if slugResp.Slug != types.SuccessSlug {
		return nil, fmt.Errorf("API error (%s): %s", slugResp.Slug, slugResp.Error)
	}

	// Data should be a slice of instances
	var createdInstances []*models.Instance
	jsonData, err := json.Marshal(slugResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal slugResp.Data for CreateInstance: %w", err)
	}

	if err := json.Unmarshal(jsonData, &createdInstances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created instances from slugResp.Data: %w", err)
	}

	return createdInstances, nil
}

// DeleteInstances deletes specified instances for a project
func (c *APIClient) DeleteInstances(ctx context.Context, req types.DeleteInstancesRequest) error {
	endpoint := routes.TerminateInstancesURL()
	return c.executeRequest(ctx, http.MethodDelete, endpoint, req, nil)
}

// User method implementation

// GetUserByID retrieves a user by id
func (c *APIClient) GetUserByID(ctx context.Context, params handlers.UserGetByIDParams) (models.User, error) {
	var response models.User // Use pkg/models.User
	if err := c.executeRPC(ctx, handlers.UserGetByID, params, &response); err != nil {
		return models.User{}, err
	}
	return response, nil
}

// GetUsers retrieves a user by username
func (c *APIClient) GetUsers(ctx context.Context, params handlers.UserGetParams) (types.UserResponse, error) {
	var response types.UserResponse
	if err := c.executeRPC(ctx, handlers.UserGet, params, &response); err != nil {
		return types.UserResponse{}, err
	}
	return response, nil
}

// CreateUser creates a new user
func (c *APIClient) CreateUser(ctx context.Context, params handlers.CreateUserParams) (types.CreateUserResponse, error) {
	var response types.CreateUserResponse
	if err := c.executeRPC(ctx, handlers.UserCreate, params, &response); err != nil {
		return types.CreateUserResponse{}, err
	}
	return response, nil
}

// DeleteUser user deletes a user
func (c *APIClient) DeleteUser(ctx context.Context, params handlers.DeleteUserParams) error {
	return c.executeRPC(ctx, handlers.UserDelete, params, nil)
}

// CreateProject creates a new project
func (c *APIClient) CreateProject(ctx context.Context, params handlers.ProjectCreateParams) (models.Project, error) {
	var project models.Project // Use pkg/models.Project
	if err := c.executeRPC(ctx, handlers.ProjectCreate, params, &project); err != nil {
		return project, err
	}
	return project, nil
}

// GetProject retrieves a project by name
func (c *APIClient) GetProject(ctx context.Context, params handlers.ProjectGetParams) (models.Project, error) {
	var project models.Project // Use pkg/models.Project
	if err := c.executeRPC(ctx, handlers.ProjectGet, params, &project); err != nil {
		return models.Project{}, err
	}
	return project, nil
}

// ListProjects lists all projects
func (c *APIClient) ListProjects(ctx context.Context, params handlers.ProjectListParams) ([]*models.Project, error) {
	var listResponse types.ListResponse[models.Project] // Use pkg/types
	if err := c.executeRPC(ctx, handlers.ProjectList, params, &listResponse); err != nil {
		return nil, err
	}
	return listResponse.Rows, nil
}

// DeleteProject deletes a project by name
func (c *APIClient) DeleteProject(ctx context.Context, params handlers.ProjectDeleteParams) error {
	return c.executeRPC(ctx, handlers.ProjectDelete, params, nil)
}

// ListProjectInstances lists all instances for a project
func (c *APIClient) ListProjectInstances(ctx context.Context, params handlers.ProjectListInstancesParams) ([]*models.Instance, error) {
	var listResponse types.ListResponse[models.Instance] // Use pkg/types
	if err := c.executeRPC(ctx, handlers.ProjectListInstances, params, &listResponse); err != nil {
		return nil, err
	}
	return listResponse.Rows, nil
}

// Task methods implementation

// GetTask retrieves a task by name
func (c *APIClient) GetTask(ctx context.Context, params handlers.TaskGetParams) (models.Task, error) {
	var task models.Task // Use pkg/models.Task
	if err := c.executeRPC(ctx, handlers.TaskGet, params, &task); err != nil {
		return models.Task{}, err
	}
	return task, nil
}

// ListTasks lists all tasks
func (c *APIClient) ListTasks(ctx context.Context, params handlers.TaskListParams) ([]*models.Task, error) {
	var listResponse types.ListResponse[models.Task] // Use pkg/types
	if err := c.executeRPC(ctx, handlers.TaskList, params, &listResponse); err != nil {
		return nil, err
	}
	return listResponse.Rows, nil
}

// ListTasksByInstanceID retrieves tasks for a specific instance ID, with optional action and pagination.
func (c *APIClient) ListTasksByInstanceID(ctx context.Context, ownerID uint, instanceID uint, actionFilter string, opts *models.ListOptions) ([]*models.Task, error) {
	queryParams := url.Values{}
	// Add owner_id to query parameters
	queryParams.Set("owner_id", strconv.FormatUint(uint64(ownerID), 10))

	if actionFilter != "" {
		queryParams.Set("action", actionFilter)
	}
	if opts != nil {
		if opts.Limit > 0 {
			queryParams.Set("limit", strconv.Itoa(opts.Limit))
		}
		if opts.Offset > 0 {
			queryParams.Set("offset", strconv.Itoa(opts.Offset))
		}
	}

	endpoint := routes.ListInstanceTasksURL(strconv.FormatUint(uint64(instanceID), 10), queryParams)

	// The response from the server is expected to be types.SlugResponse
	// where Data contains types.ListResponse[models.Task]
	var slugResp types.SlugResponse
	if err := c.executeRequest(ctx, http.MethodGet, endpoint, nil, &slugResp); err != nil {
		return nil, fmt.Errorf("failed to execute request for ListTasksByInstanceID: %w", err)
	}

	if slugResp.Slug != types.SuccessSlug {
		return nil, fmt.Errorf("API error on ListTasksByInstanceID (%s): %s", slugResp.Slug, slugResp.Error)
	}

	// Ensure Data is not nil
	if slugResp.Data == nil {
		return nil, fmt.Errorf("API response for ListTasksByInstanceID missing data")
	}

	// Manually unmarshal slugResp.Data into types.ListResponse[models.Task]
	var listResponse types.ListResponse[models.Task]
	jsonData, err := json.Marshal(slugResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal slugResp.Data for ListTasksByInstanceID: %w", err)
	}

	if err := json.Unmarshal(jsonData, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ListResponse from slugResp.Data for ListTasksByInstanceID: %w", err)
	}

	return listResponse.Rows, nil
}

// TerminateTask terminates a task by name
func (c *APIClient) TerminateTask(ctx context.Context, params handlers.TaskTerminateParams) error {
	return c.executeRPC(ctx, handlers.TaskTerminate, params, nil)
}

// UpdateTaskStatus updates the status of a task
func (c *APIClient) UpdateTaskStatus(ctx context.Context, params handlers.TaskUpdateStatusParams) error {
	return c.executeRPC(ctx, handlers.TaskUpdateStatus, params, nil)
}

// SSH Key methods implementation

// CreateSSHKey creates a new SSH key
func (c *APIClient) CreateSSHKey(ctx context.Context, params handlers.SSHKeyCreateParams) (internalmodels.SSHKey, error) {
	var key internalmodels.SSHKey
	err := c.executeRPC(ctx, handlers.SSHKeyCreate, params, &key)
	return key, err
}

// ListSSHKeys lists all SSH keys for a specific owner
func (c *APIClient) ListSSHKeys(ctx context.Context, params handlers.SSHKeyListParams) ([]*internalmodels.SSHKey, error) {
	var keys []*internalmodels.SSHKey
	err := c.executeRPC(ctx, handlers.SSHKeyList, params, &keys)
	return keys, err
}

// DeleteSSHKey deletes an SSH key
func (c *APIClient) DeleteSSHKey(ctx context.Context, params handlers.SSHKeyDeleteParams) error {
	return c.executeRPC(ctx, handlers.SSHKeyDelete, params, nil)
}
