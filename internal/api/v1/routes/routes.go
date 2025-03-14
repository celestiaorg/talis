package routes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/celestiaorg/talis/internal/api/v1/handlers"

	"github.com/gofiber/fiber/v2"
)

// API base configuration
const (
	// DefaultPort is the default port for the API
	DefaultPort = "8080"
	// APIPrefix is the prefix for all API endpoints
	APIPrefix = "/api/v1"
)

// DefaultBaseURL is the default base URL for the API
var DefaultBaseURL = fmt.Sprintf("http://localhost:%s", DefaultPort)

// Route names for lookup
const (
	// Jobs routes
	RouteGetJob    = "GetJob"
	RouteCreateJob = "CreateJob"
	RouteListJobs  = "ListJobs"

	// Instance routes
	RouteGetInstance         = "GetInstance"
	RouteListInstances       = "ListInstances"
	RouteGetInstanceMetadata = "GetInstanceMetadata"

	// Job instance routes
	RouteGetJobInstances   = "GetJobInstances"
	RouteGetJobPublicIPs   = "GetJobPublicIPs"
	RouteGetJobInstance    = "GetJobInstance"
	RouteCreateJobInstance = "CreateJobInstance"
	RouteDeleteJobInstance = "DeleteJobInstance"

	// Health check
	RouteHealthCheck = "HealthCheck"
)

// routeCache stores extracted routes
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
	v1 := app.Group(APIPrefix)

	// Jobs endpoints
	jobs := v1.Group("/jobs")
	jobs.Get("/:id", jobHandler.GetJobStatus).Name(RouteGetJob)
	jobs.Post("/", jobHandler.CreateJob).Name(RouteCreateJob)
	jobs.Get("/", jobHandler.ListJobs).Name(RouteListJobs)

	// Job instances endpoints
	jobInstances := jobs.Group("/:jobId/instances")
	jobInstances.Get("/", instanceHandler.GetInstancesByJobID).Name(RouteGetJobInstances)
	jobInstances.Get("/public-ips", instanceHandler.GetPublicIPs).Name(RouteGetJobPublicIPs)
	jobInstances.Get("/:instanceId", instanceHandler.GetInstance).Name(RouteGetJobInstance)
	jobInstances.Post("/", instanceHandler.CreateInstance).Name(RouteCreateJobInstance)
	jobInstances.Delete("/", instanceHandler.DeleteInstance).Name(RouteDeleteJobInstance)

	// Admin endpoints for instances (all jobs)
	instances := v1.Group("/instances")
	instances.Get("/", instanceHandler.ListInstances).Name(RouteListInstances)
	instances.Get("/all-metadata", instanceHandler.GetAllMetadata).Name(RouteGetInstanceMetadata)
	instances.Get("/:id", instanceHandler.GetInstance).Name(RouteGetInstance)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	}).Name(RouteHealthCheck)
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

// Helper functions for common routes

// GetJobURL returns the URL for getting a job by ID
func GetJobURL(id string) string {
	return BuildURL(RouteGetJob, map[string]string{"id": id})
}

// ListJobsURL returns the URL for listing jobs
func ListJobsURL() string {
	return BuildURL(RouteListJobs, nil)
}

// GetInstanceURL returns the URL for getting an instance by ID
func GetInstanceURL(id string) string {
	return BuildURL(RouteGetInstance, map[string]string{"id": id})
}

// ListInstancesURL returns the URL for listing instances
func ListInstancesURL() string {
	return BuildURL(RouteListInstances, nil)
}

// CreateInstanceURL returns the URL for creating an instance
func CreateInstanceURL() string {
	return BuildURL(RouteListInstances, nil) // Uses the same endpoint as list but with POST
}

// DeleteInstanceURL returns the URL for deleting an instance
func DeleteInstanceURL() string {
	return BuildURL(RouteListInstances, nil) // Uses the same endpoint as list but with DELETE
}
