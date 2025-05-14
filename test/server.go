package test

import (
	"net/http/httptest"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"

	"github.com/celestiaorg/talis/internal/logger"
	"github.com/celestiaorg/talis/internal/services"
	"github.com/celestiaorg/talis/pkg/api/v1/client"
	"github.com/celestiaorg/talis/pkg/api/v1/handlers"
	"github.com/celestiaorg/talis/pkg/api/v1/routes"
)

// testClientTimeout is the timeout for test API client requests
const testClientTimeout = 5 * time.Second

// SetupServer configures the test suite with a real API server
func SetupServer(suite *Suite) {
	// Create Fiber app with default config
	suite.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	// Add logger
	suite.App.Use(logger.APILogger())

	// Create services
	userService := services.NewUserService(suite.UserRepo)
	projectService := services.NewProjectService(suite.ProjectRepo)
	taskService := services.NewTaskService(suite.TaskRepo, projectService)
	instanceService := services.NewInstanceService(suite.InstanceRepo, taskService, projectService)

	// Create handlers
	apiHandler := handlers.NewAPIHandler(instanceService, projectService, taskService, userService)
	instanceHandler := handlers.NewInstanceHandler(apiHandler)
	projectHandler := handlers.NewProjectHandlers(apiHandler)
	taskHandler := handlers.NewTaskHandlers(apiHandler)
	userHandler := handlers.NewUserHandler(apiHandler)
	rpcHandler := &handlers.RPCHandler{
		ProjectHandlers: projectHandler,
		TaskHandlers:    taskHandler,
		UserHandlers:    userHandler,
	}

	// Register routes
	routes.RegisterRoutes(suite.App, instanceHandler, rpcHandler, taskHandler)

	// Create test server using adaptor to convert Fiber app to http.Handler
	suite.Server = httptest.NewServer(adaptor.FiberApp(suite.App))

	// Create API client with test configuration
	client, err := client.NewClient(&client.Options{
		BaseURL: suite.Server.URL,
		Timeout: testClientTimeout,
	})
	suite.Require().NoError(err, "Failed to create API client")
	suite.APIClient = client

	// Launch worker
	var wg sync.WaitGroup
	wg.Add(1)
	suite.workerWG = &wg
	worker := services.NewWorker(instanceService, projectService, taskService, userService, 100*time.Millisecond)
	go worker.LaunchWorker(suite.ctx, &wg)

	// Update cleanup to close server
	originalCleanup := suite.cleanup
	suite.cleanup = func() {
		if suite.Server != nil {
			suite.Server.Close()
		}
		if originalCleanup != nil {
			originalCleanup()
		}
	}
}
