// Package routes defines the API routes and URL structure
package routes

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	fiber "github.com/gofiber/fiber/v2"

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

	// User routes
	GetUsers    = "GetUsers"
	GetUserByID = "GetUserByID"
	CreateUser  = "CreateUser"
	DeleteUser  = "DeleteUser"

	// Project routes
	CreateProject       = "CreateProject"
	ListProjects        = "ListProjects"
	GetProject          = "GetProject"
	DeleteProject       = "DeleteProject"
	GetProjectInstances = "GetProjectInstances"

	// Task routes
	CreateTask       = "CreateTask"
	ListProjectTasks = "ListProjectTasks"
	GetTask          = "GetTask"
	UpdateTaskStatus = "UpdateTaskStatus"
	DeleteTask       = "DeleteTask"
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
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	taskHandler *handlers.TaskHandler,
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

	// ---------------------------
	// User endpoints
	users := v1.Group("/users")
	users.Get("/", userHandler.GetUsers).Name(GetUsers)
	users.Get("/:id", userHandler.GetUserByID).Name(GetUserByID)
	users.Post("/", userHandler.CreateUser).Name(CreateUser)
	users.Delete("/:id", userHandler.DeleteUser).Name(DeleteUser)

	// Project routes
	projects := v1.Group("/projects")
	projects.Post("/", projectHandler.CreateProject).Name(CreateProject)
	projects.Get("/", projectHandler.ListProjects).Name(ListProjects)
	projects.Get("/:name", projectHandler.GetProject).Name(GetProject)
	projects.Get("/:name/instances", projectHandler.ListProjectInstances).Name(GetProjectInstances)
	projects.Delete("/:name", projectHandler.DeleteProject).Name(DeleteProject)

	// Task routes
	projects.Get("/:name/tasks", taskHandler.ListProjectTasks).Name(ListProjectTasks)
	projects.Get("/:name/tasks/:taskName", taskHandler.GetTask).Name(GetTask)
	projects.Put("/:name/tasks/:taskName/status", taskHandler.UpdateTaskStatus).Name(UpdateTaskStatus)
	projects.Delete("/:name/tasks/:taskName", taskHandler.DeleteTask).Name(DeleteTask)
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
		mockUserHandler := &handlers.UserHandler{}
		mockProjectHandler := &handlers.ProjectHandler{}
		mockTaskHandler := &handlers.TaskHandler{}

		// Register routes with mock handlers
		RegisterRoutes(app, mockInstanceHandler, mockJobHandler, mockUserHandler, mockProjectHandler, mockTaskHandler)

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

// Job Routes

// GetJobsURL returns the URL for getting jobs
func GetJobsURL(queryParams url.Values) string {
	return BuildURL(GetJobs, nil, queryParams)
}

// GetJobURL returns the URL for getting a job by ID
func GetJobURL(id string) string {
	return BuildURL(GetJob, map[string]string{"id": id}, nil)
}

// GetJobMetadataURL returns the URL for getting job metadata
func GetJobMetadataURL(id string, queryParams url.Values) string {
	return BuildURL(GetMetadataByJobID, map[string]string{"id": id}, queryParams)
}

// GetJobInstancesURL returns the URL for getting job instances
func GetJobInstancesURL(jobID string, queryParams url.Values) string {
	return BuildURL(GetInstancesByJobID, map[string]string{"id": jobID}, queryParams)
}

// GetJobStatusURL returns the URL for getting job status by ID
func GetJobStatusURL(id string) string {
	return BuildURL(GetJobStatus, map[string]string{"id": id}, nil)
}

// CreateJobURL returns the URL for creating a job
func CreateJobURL() string {
	return BuildURL(CreateJob, nil, nil)
}

// UpdateJobURL returns the URL for updating a job by ID
func UpdateJobURL(id string) string {
	return BuildURL(UpdateJob, map[string]string{"id": id}, nil)
}

// DeleteJobURL returns the URL for deleting a job by ID
func DeleteJobURL(id string) string {
	return BuildURL(TerminateJob, map[string]string{"id": id}, nil)
}

// User Routes

// GetUserByIDURL returns the URL for getting a user by ID
func GetUserByIDURL(id string) string {
	return BuildURL(GetUserByID, map[string]string{"id": id}, nil)
}

// GetUsersURL returns the URL for getting a user
func GetUsersURL(queryParams url.Values) string {
	return BuildURL(GetUsers, nil, queryParams)
}

// CreateUserURL returns the URL for creating a user
func CreateUserURL() string {
	return BuildURL(CreateUser, nil, nil)
}

// DeleteUserURL returns the URL for deleting a user by ID
func DeleteUserURL(id string) string {
	return BuildURL(DeleteUser, map[string]string{"id": id}, nil)
}

// Project Routes

// CreateProjectURL returns the URL for creating a project
func CreateProjectURL() string {
	return BuildURL(CreateProject, nil, nil)
}

// ListProjectsURL returns the URL for listing projects
func ListProjectsURL() string {
	return BuildURL(ListProjects, nil, nil)
}

// GetProjectURL returns the URL for getting a project by ID
func GetProjectURL(id string) string {
	return BuildURL(GetProject, map[string]string{"name": id}, nil)
}

// DeleteProjectURL returns the URL for deleting a project by ID
func DeleteProjectURL(id string) string {
	return BuildURL(DeleteProject, map[string]string{"name": id}, nil)
}

// Task Routes

// ListProjectTasksURL returns the URL for listing tasks in a project
func ListProjectTasksURL(projectName string) string {
	return BuildURL(ListProjectTasks, map[string]string{"name": projectName}, nil)
}

// GetTaskURL returns the URL for getting a task by name
func GetTaskURL(projectName string, taskName string) string {
	return BuildURL(GetTask, map[string]string{"name": projectName, "taskName": taskName}, nil)
}

// UpdateTaskStatusURL returns the URL for updating a task status
func UpdateTaskStatusURL(projectName string, taskName string) string {
	return BuildURL(UpdateTaskStatus, map[string]string{"name": projectName, "taskName": taskName}, nil)
}

// DeleteTaskURL returns the URL for deleting a task
func DeleteTaskURL(projectName string, taskName string) string {
	return BuildURL(DeleteTask, map[string]string{"name": projectName, "taskName": taskName}, nil)
}
