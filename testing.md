# Talis Testing Strategy

## Current Testing Architecture

The current testing architecture in Talis isolates components from each other
during testing:

1. **API Client Testing**: The API client (`internal/api/v1/client`) is
currently mocked for testing using a mock implementation in
`internal/api/v1/client/mock/client.go`. Tests don't make actual HTTP calls to
an API server.

2. **Digital Ocean Provider Testing**: The Digital Ocean provider has its own
mocking system in `internal/compute/do_mocks.go` that's used for unit testing
the provider functionality.

3. **Isolation Limitation**: The current approach makes it difficult to catch
compatibility issues between the API client and the API server when the API
contract changes.

## Proposed Improvement: Integration Test Package

To address the limitations of the current isolated testing approach while still
maintaining control over external dependencies, we propose creating a dedicated
integration test package that sets up a complete test environment with a real
API server and mocked external providers.

### Key Components

1. **Test Package**: Create a dedicated `test` package to house all
integration testing infrastructure.

2. **Test Environment**: Implement a `TestEnvironment` struct that handles setting up and tearing down a complete test environment with:
   - In-memory database
   - Real API server
   - Real API client
   - Mocked external providers (e.g., Digital Ocean)

3. **Dependency Injection**: Allow tests to inject mocked providers into the test environment.

4. **Helper Functions**: Provide helper functions for common test scenarios.

### Implementation Details

#### 1. Test Environment Structure

```go
// test/environment.go
package test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/celestiaorg/talis/internal/api/v1/client"
	"github.com/celestiaorg/talis/internal/api/v1/handlers"
	"github.com/celestiaorg/talis/internal/api/v1/routes"
	"github.com/celestiaorg/talis/internal/api/v1/services"
	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/db"
	"github.com/celestiaorg/talis/internal/db/repos"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestEnvironment encapsulates all components needed for integration testing
type TestEnvironment struct {
	// Server components
	App            *fiber.App
	Server         *httptest.Server
	BaseURL        string
	
	// Client components
	APIClient      client.Client
	
	// Database components
	DB             *gorm.DB
	JobRepo        *repos.JobRepository
	InstanceRepo   *repos.InstanceRepository
	
	// Service components  
	JobService     *services.JobService
	InstanceService *services.InstanceService
	
	// Mock providers
	MockDOClient   *compute.MockDOClient
	
	// Cleanup function
	cleanup        func()
}

// NewTestEnvironment creates a complete test environment with a real API
// server that uses mocked external dependencies
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()
	
	env := &TestEnvironment{}
	
	// Setup database - in-memory SQLite for tests
	dbConn, err := db.NewInMemoryDB()
	require.NoError(t, err, "Failed to create in-memory database")
	env.DB = dbConn
	
	// Run migrations
	err = db.RunMigrations(env.DB)
	require.NoError(t, err, "Failed to run database migrations")
	
	// Create repositories
	env.JobRepo = repos.NewJobRepository(env.DB)
	env.InstanceRepo = repos.NewInstanceRepository(env.DB)
	
	// Create mocked DO client
	env.MockDOClient = compute.NewMockDOClient()
	
	// Setup Digital Ocean provider with mocked client
	provider := &compute.DigitalOceanProvider{}
	provider.SetClient(env.MockDOClient)
	
	// Create services
	env.JobService = services.NewJobService(env.JobRepo)
	env.InstanceService = services.NewInstanceService(env.InstanceRepo, env.JobService)
	
	// Create API handlers with injected services
	jobHandler := handlers.NewJobHandler(env.JobService, env.InstanceService, provider)
	instanceHandler := handlers.NewInstanceHandler(env.InstanceService)
	healthHandler := handlers.NewHealthHandler()
	
	// Create Fiber app
	env.App = fiber.New()
	
	// Register routes
	routes.RegisterRoutes(env.App, jobHandler, instanceHandler, healthHandler)
	
	// Create test server
	env.Server = httptest.NewServer(env.App.Handler())
	env.BaseURL = env.Server.URL
	
	// Create API client connected to test server
	apiClient, err := client.NewClient(&client.ClientOptions{
		BaseURL: env.BaseURL,
	})
	require.NoError(t, err, "Failed to create API client")
	env.APIClient = apiClient
	
	// Setup cleanup function
	env.cleanup = func() {
		env.Server.Close()
		// Any other cleanup needed
	}
	
	return env
}

// Cleanup tears down the test environment
func (e *TestEnvironment) Cleanup() {
	if e.cleanup != nil {
		e.cleanup()
	}
}

// Context returns a context for use in tests
func (e *TestEnvironment) Context() context.Context {
	return context.Background()
}
```

#### 2. Digital Ocean Provider Mock Helpers

```go
// test/mocks.go
package test

import (
	"context"
	
	"github.com/celestiaorg/talis/internal/compute"
	"github.com/digitalocean/godo"
)

// SetupMockDOForInstanceCreation configures the mock Digital Ocean client
// to respond appropriately to instance creation requests
func SetupMockDOForInstanceCreation(mockClient *compute.MockDOClient, config InstanceCreationConfig) {
	// Setup SSH key mocking
	mockClient.MockKeyService.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return []godo.Key{
			{
				ID:   config.SSHKeyID,
				Name: config.SSHKeyName,
			},
		}, nil, nil
	}
	
	// Setup droplet creation mocking
	mockClient.MockDropletService.CreateFunc = func(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		return &godo.Droplet{
			ID:   config.DropletID,
			Name: req.Name,
			Networks: &godo.Networks{
				V4: []godo.NetworkV4{
					{
						Type:      "public",
						IPAddress: config.IPAddress,
					},
				},
			},
			Region: &godo.Region{
				Slug: req.Region,
			},
		}, nil, nil
	}
	
	// Setup droplet getter mocking
	mockClient.MockDropletService.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		if id == config.DropletID {
			return &godo.Droplet{
				ID:   config.DropletID,
				Name: config.DropletName,
				Networks: &godo.Networks{
					V4: []godo.NetworkV4{
						{
							Type:      "public",
							IPAddress: config.IPAddress,
						},
					},
				},
				Region: &godo.Region{
					Slug: config.Region,
				},
			}, nil, nil
		}
		return nil, nil, nil
	}
}

// InstanceCreationConfig holds configuration for mock instance creation
type InstanceCreationConfig struct {
	SSHKeyID     int
	SSHKeyName   string
	DropletID    int
	DropletName  string
	IPAddress    string
	Region       string
}

// DefaultInstanceCreationConfig returns a default configuration for mocking instance creation
func DefaultInstanceCreationConfig() InstanceCreationConfig {
	return InstanceCreationConfig{
		SSHKeyID:    12345,
		SSHKeyName:  "test-key",
		DropletID:   54321,
		DropletName: "test-instance",
		IPAddress:   "192.0.2.1",
		Region:      "nyc1",
	}
}
```

#### 3. Test Utilities

```go
// test/utils.go
package test

import (
	"testing"
	
	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/stretchr/testify/assert"
)

// AssertInstanceEquals checks if two instances match
func AssertInstanceEquals(t *testing.T, expected, actual infrastructure.InstanceInfo) {
	t.Helper()
	assert.Equal(t, expected.Name, actual.Name, "Instance name mismatch")
	assert.Equal(t, expected.IP, actual.IP, "Instance IP mismatch")
	assert.Equal(t, expected.Provider, actual.Provider, "Instance provider mismatch")
	assert.Equal(t, expected.Region, actual.Region, "Instance region mismatch")
	assert.Equal(t, expected.Size, actual.Size, "Instance size mismatch")
}

// CreateTestInstanceRequest creates a test instance request
func CreateTestInstanceRequest(provider, region, size string) infrastructure.InstanceRequest {
	return infrastructure.InstanceRequest{
		Provider:          provider,
		NumberOfInstances: 1,
		Provision:         true,
		Region:            region,
		Size:              size,
		Image:             "ubuntu-20-04-x64",
		Tags:              []string{"test", "integration"},
		SSHKeyName:        "test-key",
	}
}
```

### Example Usage

Here's how you would use the integration test package to write a test that verifies the API client can correctly call the API server, which then uses a mocked Digital Ocean provider:

```go
// test/example_test.go
package test_test

import (
	"testing"
	
	"github.com/celestiaorg/talis/test"
	"github.com/celestiaorg/talis/internal/types/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateJobInstance(t *testing.T) {
	// Setup test environment
	env := test.NewTestEnvironment(t)
	defer env.Cleanup()
	
	// Configure Digital Ocean mock
	mockConfig := test.DefaultInstanceCreationConfig()
	mockConfig.Region = "sfo3"
	mockConfig.DropletName = "test-job-instance"
	test.SetupMockDOForInstanceCreation(env.MockDOClient, mockConfig)
	
	// Create a test job first
	jobReq := infrastructure.CreateRequest{
		JobName:      "test-job",
		InstanceName: "test-instance",
		ProjectName:  "test-project",
		WebhookURL:   "https://example.com/webhook",
		Instances: []infrastructure.InstanceRequest{
			test.CreateTestInstanceRequest("digitalocean", "sfo3", "s-1vcpu-1gb"),
		},
	}
	
	jobResp, err := env.APIClient.CreateJob(env.Context(), jobReq)
	require.NoError(t, err, "Failed to create job")
	require.NotNil(t, jobResp, "Job response should not be nil")
	
	// Now add an instance to the job
	instanceReq := test.CreateTestInstanceRequest("digitalocean", "sfo3", "s-1vcpu-1gb")
	instanceResp, err := env.APIClient.CreateJobInstance(env.Context(), jobResp.ID, instanceReq)
	
	// Verify the response
	require.NoError(t, err, "Failed to create job instance")
	require.NotNil(t, instanceResp, "Instance response should not be nil")
	
	// Check instance details
	assert.Equal(t, "digitalocean", instanceResp.Provider)
	assert.Equal(t, "sfo3", instanceResp.Region)
	assert.Equal(t, "s-1vcpu-1gb", instanceResp.Size)
	assert.NotEmpty(t, instanceResp.IP)
	
	// Verify that our mock was called with the right parameters
	// This can be done by checking the mock's call history
}
```

### Benefits of This Approach

1. **API Contract Testing**: Tests the entire API call path including request/response serialization, URL construction, and error handling.

2. **Isolation from External Services**: Despite using real API and client components, tests remain isolated from actual cloud providers.

3. **Realistic Test Environment**: Tests are run against a setup that mimics the production environment closely.

4. **Centralized Test Infrastructure**: All the complexity of setting up test environments is centralized in one package.

5. **Easy to Use in CI/CD**: These tests can be run in CI/CD pipelines without any external dependencies.

6. **Flexible Mocking**: The approach allows for fine-grained control over mock behavior to test different scenarios.

### Implementation Steps

1. Create the `test` package structure
2. Implement the test environment setup and tear down
3. Create mock helpers for external providers
4. Add test utilities
5. Write example tests
6. Integrate with the CI/CD pipeline

### Integration with Existing Tests

This new approach can coexist with the current unit testing approach. Unit tests can continue to focus on testing isolated components, while integration tests can use this new infrastructure to test how components work together. Over time, some unit tests might be replaced with integration tests if they provide better coverage and reliability. 