package test

import (
	"net/http/httptest"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"

	"github.com/celestiaorg/talis/internal/api/v1/client"
	"github.com/celestiaorg/talis/internal/api/v1/handlers"
	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/logger"
)

// testClientTimeout is the timeout for test API client requests
const testClientTimeout = 5 * time.Second

// SetupServer configures the test suite with a real API server
func SetupServer(suite *TestSuite) {
	// Create Fiber app with default config
	suite.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	// Add logger
	suite.App.Use(logger.APILogger())

	// Create services
	jobService := services.NewJobService(suite.JobRepo, suite.InstanceRepo)
	instanceService := services.NewInstanceService(suite.InstanceRepo, jobService)

	// Create handlers
	jobHandler := handlers.NewJobHandler(jobService, instanceService)
	instanceHandler := handlers.NewInstanceHandler(instanceService)
	projectHandler := handlers.NewProjectHandler(nil) // TODO: Add project service
	taskHandler := handlers.NewTaskHandler(nil)       // TODO: Add task service

	// Register routes
	routes.RegisterRoutes(suite.App, instanceHandler, jobHandler, projectHandler, taskHandler)

	// Create test server using adaptor to convert Fiber app to http.Handler
	suite.Server = httptest.NewServer(adaptor.FiberApp(suite.App))

	// Create API client with test configuration
	client, err := client.NewClient(&client.ClientOptions{
		BaseURL: suite.Server.URL,
		Timeout: testClientTimeout,
	})
	suite.Require().NoError(err, "Failed to create API client")
	suite.APIClient = client

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
