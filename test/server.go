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
)

// testClientTimeout is the timeout for test API client requests
const testClientTimeout = 5 * time.Second

// WithServer is an option that configures the test environment with a real API server
func WithServer() Option {
	return func(env *TestEnvironment) {
		// Create Fiber app with default config
		env.App = fiber.New(fiber.Config{
			DisableStartupMessage: true,
		})

		// Create services
		jobService := services.NewJobService(env.JobRepo)
		instanceService := services.NewInstanceService(env.InstanceRepo, jobService)

		// Create handlers
		jobHandler := handlers.NewJobHandler(jobService)
		instanceHandler := handlers.NewInstanceHandler(instanceService, jobService)

		// Register routes
		routes.RegisterRoutes(env.App, instanceHandler, jobHandler)

		// Create test server using adaptor to convert Fiber app to http.Handler
		env.Server = httptest.NewServer(adaptor.FiberApp(env.App))

		// Create API client with test configuration
		client, err := client.NewClient(&client.ClientOptions{
			BaseURL: env.Server.URL,
			Timeout: testClientTimeout,
		})
		env.Require().NoError(err, "Failed to create API client")
		env.APIClient = client

		// Update cleanup to close server
		originalCleanup := env.cleanup
		env.cleanup = func() {
			if env.Server != nil {
				env.Server.Close()
			}
			if originalCleanup != nil {
				originalCleanup()
			}
		}
	}
}
