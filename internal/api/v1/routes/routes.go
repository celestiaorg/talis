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
	GetJob    = "GetJobStatus"
	CreateJob = "CreateJob"
	ListJobs  = "ListJobs"

	// Job instance routes
	GetJobInstances   = "GetJobInstances"
	GetJobPublicIPs   = "GetJobPublicIPs"
	GetJobInstance    = "GetJobInstance"
	CreateJobInstance = "CreateJobInstance"
	DeleteJobInstance = "DeleteJobInstance"

	// Instance routes
	ListInstances       = "ListInstances"
	GetInstanceMetadata = "GetInstanceMetadata"
	GetInstance         = "GetInstance"

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

	// Jobs endpoints
	jobs := v1.Group("/jobs")
	jobs.Get("/:id", jobHandler.GetJobStatus).Name(GetJob)
	jobs.Post("/", jobHandler.CreateJob).Name(CreateJob)
	jobs.Get("/", jobHandler.ListJobs).Name(ListJobs)

	// Job instances endpoints
	jobInstances := jobs.Group("/:jobId/instances")
	jobInstances.Get("/", instanceHandler.GetInstancesByJobID).Name(GetJobInstances)
	jobInstances.Get("/public-ips", instanceHandler.GetPublicIPs).Name(GetJobPublicIPs)
	jobInstances.Get("/:instanceId", instanceHandler.GetInstance).Name(GetJobInstance)
	jobInstances.Post("/", instanceHandler.CreateInstance).Name(CreateJobInstance)
	jobInstances.Delete("/", instanceHandler.DeleteInstance).Name(DeleteJobInstance)

	// Instance endpoints for instances (all jobs)
	instances := v1.Group("/instances")
	instances.Get("/", instanceHandler.ListInstances).Name(ListInstances)
	instances.Get("/all-metadata", instanceHandler.GetAllMetadata).Name(GetInstanceMetadata)
	instances.Get("/:id", instanceHandler.GetInstance).Name(GetInstance)

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

// Job route helpers

// GetJobURL returns the URL for getting a job by ID
func GetJobURL(id string) string {
	return BuildURL(GetJob, map[string]string{"id": id})
}

// CreateJobURL returns the URL for creating a job
func CreateJobURL() string {
	return BuildURL(CreateJob, nil)
}

// ListJobsURL returns the URL for listing jobs
func ListJobsURL() string {
	return BuildURL(ListJobs, nil)
}

// Job instance route helpers

// GetJobInstancesURL returns the URL for getting instances by job ID
func GetJobInstancesURL(jobId string) string {
	return BuildURL(GetJobInstances, map[string]string{"jobId": jobId})
}

// GetJobPublicIPsURL returns the URL for getting public IPs for a job
func GetJobPublicIPsURL(jobId string) string {
	return BuildURL(GetJobPublicIPs, map[string]string{"jobId": jobId})
}

// GetJobInstanceURL returns the URL for getting a specific job instance
func GetJobInstanceURL(jobId, instanceId string) string {
	return BuildURL(GetJobInstance, map[string]string{
		"jobId":      jobId,
		"instanceId": instanceId,
	})
}

// CreateJobInstanceURL returns the URL for creating a job instance
func CreateJobInstanceURL(jobId string) string {
	return BuildURL(CreateJobInstance, map[string]string{"jobId": jobId})
}

// DeleteJobInstanceURL returns the URL for deleting a job instance
func DeleteJobInstanceURL(jobId string) string {
	return BuildURL(DeleteJobInstance, map[string]string{"jobId": jobId})
}

// Instance route helpers

// ListInstancesURL returns the URL for listing instances
func ListInstancesURL() string {
	return BuildURL(ListInstances, nil)
}

// GetInstanceMetadataURL returns the URL for getting instance metadata
func GetInstanceMetadataURL() string {
	return BuildURL(GetInstanceMetadata, nil)
}

// GetInstanceURL returns the URL for getting an instance by ID
func GetInstanceURL(id string) string {
	return BuildURL(GetInstance, map[string]string{"id": id})
}

// Health check route helper

// HealthCheckURL returns the URL for the health check endpoint
func HealthCheckURL() string {
	return BuildURL(HealthCheck, nil)
}
