// Package routes defines the API routes and URL structure
package routes

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	fiber "github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
)

/*

To keep this file organized, routes should be organized in the following way:

1. Smallest scope first (i.e. instance routes before job routes)
2. For similar scopes, put the endpoints in alphabetical order
3. Order routes in GET, POST, PUT, DELETE order.
	a. Within this ordering, param urls (ie /:id) should go last, otherwise fiber will interpret the route slug as that param.
	b. After param considerations, order alphabetically.
4. For clarity, naming should match the action (i.e. GetJob, DeleteJob)

*/

// API base configuration
const (
	// DefaultPort is the default port for the API
	DefaultPort = "8080"
	// APIv1Prefix is the prefix for all API endpoints
	APIv1Prefix = "/api/v1"
)

// DefaultBaseURL is the default base URL for the API
var DefaultBaseURL = fmt.Sprintf("http://localhost:%s", DefaultPort)

// Route names for lookup
const (
	// Admin routes
	// No unique names, consider making Admin a private ownerID.
	AdminGetInstances         = "AdminGetInstances"
	AdminGetInstancesMetadata = "AdminGetInstancesMetadata"

	// Health check
	HealthCheck = "HealthCheck"

	// Instance routes
	GetInstances       = "GetInstances"
	GetMetadata        = "GetMetadata"
	GetPublicIPs       = "GetPublicIPs"
	GetInstance        = "GetInstance"
	CreateInstance     = "CreateInstance"
	TerminateInstances = "TerminateInstances"

	// User routes
	GetUsers    = "GetUsers"
	GetUserByID = "GetUserByID"
	CreateUser  = "CreateUser"
	DeleteUser  = "DeleteUser"

	// RPC routes
	RPC = "RPC"
)

// routeCache stores extracted routes for use prior to compilation
var (
	routeCache     map[string]string
	routeCacheMu   sync.RWMutex
	routeCacheInit sync.Once
)

// RegisterRoutes configures all the v1 routes
//
// NOTE: route ordering is important because routes will try and match in the order they are registered.
// For example, if we register GetInstance before GetInstanceMetadata, the /all-metadata will get interpreted as an instance ID.
func RegisterRoutes(
	app *fiber.App,
	instanceHandler *handlers.InstanceHandler,
	userHandler *handlers.UserHandler,
	rpcHandler *handlers.RPCHandler,
) {
	// API v1 routes
	v1 := app.Group(APIv1Prefix)

	// Admin endpoints for instances (all jobs)
	// TODO: We should create a private Admin OwnerID. Then these could probably just be a part of the instance endpoints.
	adminInstances := v1.Group("/admin/instances")
	adminInstances.Get("/", instanceHandler.ListInstances).Name(AdminGetInstances)
	adminInstances.Get("/all-metadata", instanceHandler.GetAllMetadata).Name(AdminGetInstancesMetadata)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	}).Name(HealthCheck)

	// Instances endpoints
	// TODO: These should be filtered by OwnerID
	instances := v1.Group("/instances")
	instances.Get("/", instanceHandler.ListInstances).Name(GetInstances)
	instances.Get("/all-metadata", instanceHandler.GetAllMetadata).Name(GetMetadata)
	instances.Get("/public-ips", instanceHandler.GetPublicIPs).Name(GetPublicIPs)
	instances.Get("/:id", instanceHandler.GetInstance).Name(GetInstance)
	instances.Post("/", instanceHandler.CreateInstance).Name(CreateInstance)
	instances.Delete("/", instanceHandler.TerminateInstances).Name(TerminateInstances)

	// ---------------------------
	// User endpoints
	users := v1.Group("/users")
	users.Get("/", userHandler.GetUsers).Name(GetUsers)
	users.Get("/:id", userHandler.GetUserByID).Name(GetUserByID)
	users.Post("/", userHandler.CreateUser).Name(CreateUser)
	users.Delete("/:id", userHandler.DeleteUser).Name(DeleteUser)

	// RPC endpoint as the root handler for all operations
	v1.Post("/", rpcHandler.HandleRPC).Name(RPC)
}

// initRouteCache initializes the route cache by creating a mock app and extracting routes
func initRouteCache() {
	routeCacheInit.Do(func() {
		routeCache = make(map[string]string)

		// Create a mock app
		app := fiber.New()

		// Create empty handlers for route registration
		mockInstanceHandler := &handlers.InstanceHandler{}
		mockUserHandler := &handlers.UserHandler{}
		mockRPCHandler := &handlers.RPCHandler{}

		// Register routes with mock handlers - project and task handlers are handled via RPC
		RegisterRoutes(app, mockInstanceHandler, mockUserHandler, mockRPCHandler)

		// Extract routes from the app
		for _, route := range app.GetRoutes() {
			if route.Name != "" {
				routeCache[route.Name] = route.Path
			}
		}
	})
}

// GetRoute returns the route pattern for the given route name
func GetRoute(name string) string {
	routeCacheMu.RLock()
	defer routeCacheMu.RUnlock()

	// Initialize cache if needed
	if routeCache == nil {
		routeCacheMu.RUnlock()
		initRouteCache()
		routeCacheMu.RLock()
	}

	return routeCache[name]
}

// BuildURL builds a URL for the given route name and parameters
func BuildURL(routeName string, params map[string]string, queryParams url.Values) string {
	route := GetRoute(routeName)
	if route == "" {
		return ""
	}

	// Replace parameters in the route
	for param, value := range params {
		route = strings.ReplaceAll(route, ":"+param, value)
	}

	// Remove trailing slash if it's a base endpoint with no parameters
	if strings.HasSuffix(route, "/") && !strings.Contains(route, ":") {
		route = strings.TrimSuffix(route, "/")
	}

	// Add query parameters if any
	if len(queryParams) > 0 {
		route = fmt.Sprintf("%s?%s", route, queryParams.Encode())
	}

	return route
}

// Admin Routes

// AdminInstancesURL returns the URL for getting all instances
func AdminInstancesURL() string {
	return BuildURL(AdminGetInstances, nil, nil)
}

// AdminInstancesMetadataURL returns the URL for getting all instances metadata
func AdminInstancesMetadataURL() string {
	return BuildURL(AdminGetInstancesMetadata, nil, nil)
}

// Health check route helper

// HealthCheckURL returns the URL for the health check endpoint
func HealthCheckURL() string {
	return BuildURL(HealthCheck, nil, nil)
}

// Instance route helpers

// GetInstancesURL returns the URL for getting instances
func GetInstancesURL(queryParams url.Values) string {
	return BuildURL(GetInstances, nil, queryParams)
}

// GetInstanceMetadataURL returns the URL for getting instance metadata
func GetInstanceMetadataURL(queryParams url.Values) string {
	return BuildURL(GetMetadata, nil, queryParams)
}

// GetPublicIPsURL returns the URL for getting public IPs
func GetPublicIPsURL(queryParams url.Values) string {
	return BuildURL(GetPublicIPs, nil, queryParams)
}

// GetInstanceURL returns the URL for getting an instance by ID
func GetInstanceURL(id string) string {
	return BuildURL(GetInstance, map[string]string{"id": id}, nil)
}

// CreateInstanceURL returns the URL for creating an instance
func CreateInstanceURL() string {
	return BuildURL(CreateInstance, nil, nil)
}

// TerminateInstancesURL returns the URL for terminating instances
func TerminateInstancesURL() string {
	return BuildURL(TerminateInstances, nil, nil)
}

// User route helpers

// GetUsersURL returns the URL for getting users
func GetUsersURL(queryParams url.Values) string {
	return BuildURL(GetUsers, nil, queryParams)
}

// GetUserByIDURL returns the URL for getting a user by ID
func GetUserByIDURL(id string) string {
	return BuildURL(GetUserByID, map[string]string{"id": id}, nil)
}

// CreateUserURL returns the URL for creating a user
func CreateUserURL() string {
	return BuildURL(CreateUser, nil, nil)
}

// DeleteUserURL returns the URL for deleting a user
func DeleteUserURL(id string) string {
	return BuildURL(DeleteUser, map[string]string{"id": id}, nil)
}

// RPC route helper

// RPCURL returns the URL for the RPC endpoint
func RPCURL() string {
	return BuildURL(RPC, nil, nil)
}
