package routes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"

	"github.com/celestiaorg/talis/internal/api/v1/handlers"
)

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
	// Jobs routes
	CreateJob    = "CreateJob"
	GetJobStatus = "GetJobStatus"
	ListJobs     = "ListJobs"
	SearchJobs   = "SearchJobs"
	UpdateJob    = "UpdateJob"
	TerminateJob = "TerminateJob"
	GetJob       = "GetJob"

	// Job instance routes
	CreateJobInstance   = "CreateJobInstance"
	DeleteJobInstance   = "DeleteJobInstance"
	GetJobInstance      = "GetJobInstance"
	GetJobInstances     = "GetJobInstances"
	GetJobPublicIPs     = "GetJobPublicIPs"
	GetInstancesByJobID = "GetInstancesByJobID"

	// Instance routes
	CreateInstance      = "CreateInstance"
	GetInstance         = "GetInstance"
	GetInstanceMetadata = "GetInstanceMetadata"
	ListInstances       = "ListInstances"

	// Health check
	HealthCheck = "HealthCheck"
)

// routeCache stores extracted routes for use prior to compilation
var (
	routeCache     map[string]string
	routeCacheMu   sync.RWMutex
	routeCacheInit sync.Once
)

// RegisterRoutes configures all the v1 routes
func RegisterRoutes(
	app *fiber.App,
	instanceHandler *handlers.InstanceHandler,
	jobHandler *handlers.JobHandler,
) {
	// API v1 routes
	v1 := app.Group(APIv1Prefix)

	// Instances endpoints
	instances := v1.Group("/instances")
	instances.Get("/", instanceHandler.ListInstances).Name(ListInstances)
	instances.Post("/", instanceHandler.CreateInstance).Name(CreateInstance)
	instances.Get("/:id", instanceHandler.GetInstance).Name(GetInstance)
	instances.Get("/public-ips", instanceHandler.GetPublicIPs).Name(GetJobPublicIPs)
	instances.Get("/all-metadata", instanceHandler.GetAllMetadata).Name(GetInstanceMetadata)

	// ---------------------------
	// Jobs endpoints
	jobs := v1.Group("/jobs")
	jobs.Get("/:id", jobHandler.GetJobStatus).Name(GetJobStatus)
	jobs.Post("/", jobHandler.CreateJob).Name(CreateJob)
	jobs.Delete("/:id", jobHandler.TerminateJob).Name(TerminateJob)
	jobs.Put("/:id", jobHandler.UpdateJob).Name(UpdateJob)
	jobs.Get("/search", jobHandler.SearchJobs).Name(SearchJobs)
	jobs.Get("/:jobId/instances", instanceHandler.GetInstancesByJobID).Name(GetInstancesByJobID)

	// Admin endpoints for instances (all jobs)
	adminInstances := v1.Group("/admin/instances")
	adminInstances.Get("/", instanceHandler.ListInstances).Name(ListInstances)
	adminInstances.Get("/all-metadata", instanceHandler.GetAllMetadata).Name(GetInstanceMetadata)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	}).Name(HealthCheck)
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

// CreateJobURL returns the URL for creating a job
func CreateJobURL() string {
	return BuildURL(CreateJob, nil)
}

// GetJobURL returns the URL for getting a job by ID
func GetJobURL(id string) string {
	return BuildURL(GetJob, map[string]string{"id": id})
}

// ListJobsURL returns the URL for listing jobs
func ListJobsURL() string {
	return BuildURL(ListJobs, nil)
}

// Job instance route helpers

// CreateJobInstanceURL returns the URL for creating a job instance
func CreateJobInstanceURL(jobId string) string {
	return BuildURL(CreateJobInstance, map[string]string{"jobId": jobId})
}

// DeleteJobInstanceURL returns the URL for deleting a job instance
func DeleteJobInstanceURL(jobId string) string {
	return BuildURL(DeleteJobInstance, map[string]string{"jobId": jobId})
}

// GetJobInstanceURL returns the URL for getting a specific job instance
func GetJobInstanceURL(jobId, instanceId string) string {
	return BuildURL(GetJobInstance, map[string]string{
		"jobId":      jobId,
		"instanceId": instanceId,
	})
}

// GetJobInstancesURL returns the URL for getting instances by job ID
func GetJobInstancesURL(jobId string) string {
	return BuildURL(GetJobInstances, map[string]string{"jobId": jobId})
}

// GetJobPublicIPsURL returns the URL for getting public IPs for a job
func GetJobPublicIPsURL(jobId string) string {
	return BuildURL(GetJobPublicIPs, map[string]string{"jobId": jobId})
}

// Instance route helpers

// GetInstanceURL returns the URL for getting an instance by ID
func GetInstanceURL(id string) string {
	return BuildURL(GetInstance, map[string]string{"id": id})
}

// GetInstanceMetadataURL returns the URL for getting instance metadata
func GetInstanceMetadataURL() string {
	return BuildURL(GetInstanceMetadata, nil)
}

// ListInstancesURL returns the URL for listing instances
func ListInstancesURL() string {
	return BuildURL(ListInstances, nil)
}

// Health check route helper

// HealthCheckURL returns the URL for the health check endpoint
func HealthCheckURL() string {
	return BuildURL(HealthCheck, nil)
}
