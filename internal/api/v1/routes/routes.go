package routes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/handlers"
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

	// Jobs routes
	GetJobs             = "GetJobs"
	GetJob              = "GetJob"
	GetMetadataByJobID  = "GetMetadataByJobID"
	GetInstancesByJobID = "GetInstancesByJobID"
	GetJobStatus        = "GetJobStatus"
	CreateJob           = "CreateJob"
	UpdateJob           = "UpdateJob"
	TerminateJob        = "TerminateJob"
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
	jobHandler *handlers.JobHandler,
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
	// Jobs endpoints
	// TODO: These should be filtered by OwnerID
	jobs := v1.Group("/jobs")
	jobs.Get("/", jobHandler.ListJobs).Name(GetJobs)
	jobs.Get("/:id", jobHandler.GetJob).Name(GetJob)
	jobs.Get("/:id/all-metadata", instanceHandler.GetAllMetadata).Name(GetMetadataByJobID)
	jobs.Get("/:id/instances", instanceHandler.GetInstancesByJobID).Name(GetInstancesByJobID)
	jobs.Get("/:id/status", jobHandler.GetJobStatus).Name(GetJobStatus)
	jobs.Post("/", jobHandler.CreateJob).Name(CreateJob)
	jobs.Put("/:id", jobHandler.UpdateJob).Name(UpdateJob)
	jobs.Delete("/:id", jobHandler.TerminateJob).Name(TerminateJob)
}

// initRouteCache initializes the route cache by creating a mock app and extracting routes
func initRouteCache() {
	routeCacheInit.Do(func() {
		routeCache = make(map[string]string)

		// Create a mock app
		app := fiber.New()

		// Create empty handlers for route registration
		mockInstanceHandler := &handlers.InstanceHandler{}
		mockJobHandler := &handlers.JobHandler{}

		// Register routes with mock handlers
		RegisterRoutes(app, mockInstanceHandler, mockJobHandler)

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
func BuildURL(routeName string, params map[string]string) string {
	route := GetRoute(routeName)
	if route == "" {
		return ""
	}

	// Replace parameters in the route
	for param, value := range params {
		route = strings.Replace(route, ":"+param, value, -1)
	}

	// Remove trailing slash if it's a base endpoint with no parameters
	if strings.HasSuffix(route, "/") && !strings.Contains(route, ":") {
		route = strings.TrimSuffix(route, "/")
	}

	return route
}

// Admin Routes

// AdminInstancesURL returns the URL for getting all instances
func AdminInstancesURL() string {
	return BuildURL(AdminGetInstances, nil)
}

// AdminInstancesMetadataURL returns the URL for getting all instances metadata
func AdminInstancesMetadataURL() string {
	return BuildURL(AdminGetInstancesMetadata, nil)
}

// Health check route helper

// HealthCheckURL returns the URL for the health check endpoint
func HealthCheckURL() string {
	return BuildURL(HealthCheck, nil)
}

// Instance route helpers

// GetInstancesURL returns the URL for getting instances
func GetInstancesURL() string {
	return BuildURL(GetInstances, nil)
}

// GetInstanceMetadataURL returns the URL for getting instance metadata
func GetInstanceMetadataURL() string {
	return BuildURL(GetMetadata, nil)
}

// GetPublicIPsURL returns the URL for getting public IPs
func GetPublicIPsURL() string {
	return BuildURL(GetPublicIPs, nil)
}

// GetInstanceURL returns the URL for getting an instance by ID
func GetInstanceURL(id string) string {
	return BuildURL(GetInstance, map[string]string{"id": id})
}

// CreateInstanceURL returns the URL for creating an instance
func CreateInstanceURL() string {
	return BuildURL(CreateInstance, nil)
}

// TerminateInstancesURL returns the URL for terminating instances
func TerminateInstancesURL() string {
	return BuildURL(TerminateInstances, nil)
}

// Job Routes

// GetJobsURL returns the URL for getting jobs
func GetJobsURL() string {
	return BuildURL(GetJobs, nil)
}

// GetJobURL returns the URL for getting a job by ID
func GetJobURL(id string) string {
	return BuildURL(GetJob, map[string]string{"id": id})
}

// GetJobMetadataURL returns the URL for getting job metadata by ID
func GetJobMetadataURL(id string) string {
	return BuildURL(GetMetadataByJobID, map[string]string{"id": id})
}

// GetJobInstancesURL returns the URL for getting instances by job ID
func GetJobInstancesURL(jobId string) string {
	return BuildURL(GetInstancesByJobID, map[string]string{"id": jobId})
}

// GetJobStatusURL returns the URL for getting job status by ID
func GetJobStatusURL(id string) string {
	return BuildURL(GetJobStatus, map[string]string{"id": id})
}

// CreateJobURL returns the URL for creating a job
func CreateJobURL() string {
	return BuildURL(CreateJob, nil)
}

// UpdateJobURL returns the URL for updating a job by ID
func UpdateJobURL(id string) string {
	return BuildURL(UpdateJob, map[string]string{"id": id})
}

// DeleteJobURL returns the URL for deleting a job by ID
func DeleteJobURL(id string) string {
	return BuildURL(TerminateJob, map[string]string{"id": id})
}
