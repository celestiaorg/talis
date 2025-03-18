package mocks

import (
	"context"

	"github.com/digitalocean/godo"

	"github.com/celestiaorg/talis/internal/compute/types"
)

// MockDOClient implements types.DOClient for testing
type MockDOClient struct {
	MockDropletService *MockDropletService
	MockKeyService     *MockKeyService
	StandardResponses  *StandardResponses
}

// NewMockDOClient creates a new MockDOClient with standard responses
func NewMockDOClient() *MockDOClient {
	std := newStandardResponses()

	client := &MockDOClient{
		StandardResponses: std,
	}

	client.MockDropletService = NewMockDropletService(std)
	client.MockKeyService = NewMockKeyService(std)

	return client
}

// ResetToStandard resets all mock services back to their standard success responses
func (c *MockDOClient) ResetToStandard() {
	c.MockDropletService.ResetToStandard()
	c.MockKeyService.ResetToStandard()
}

// Droplets returns the mock droplet service
func (c *MockDOClient) Droplets() types.DropletService {
	return c.MockDropletService
}

// Keys returns the mock key service
func (c *MockDOClient) Keys() types.KeyService {
	return c.MockKeyService
}

// SimulateAuthenticationFailure configures all services to return authentication errors
func (c *MockDOClient) SimulateAuthenticationFailure() {
	c.MockDropletService.SimulateAuthenticationFailure()
	c.MockKeyService.SimulateAuthenticationFailure()
}

// SimulateNotFound configures all services to return not found errors
func (c *MockDOClient) SimulateNotFound() {
	c.MockDropletService.SimulateNotFound()
	c.MockKeyService.SimulateNotFound()
}

// SimulateRateLimit configures all services to return rate limit errors
func (c *MockDOClient) SimulateRateLimit() {
	c.MockDropletService.SimulateRateLimit()
	c.MockKeyService.SimulateRateLimit()
}

// MockDropletService implements types.DropletService for testing
type MockDropletService struct {
	std                *StandardResponses
	CreateFunc         func(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultipleFunc func(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	GetFunc            func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	DeleteFunc         func(ctx context.Context, id int) (*godo.Response, error)
	ListFunc           func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	attemptCount       int // Track number of attempts for retry simulations
}

// setupStandardDropletResponses configures the standard success responses for droplet service
func setupStandardDropletResponses(s *MockDropletService) {
	s.CreateFunc = func(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		droplet := *s.std.Droplets.DefaultDroplet  // Create a copy
		droplet.Name = createRequest.Name          // Use requested name
		droplet.Region.Slug = createRequest.Region // Use requested region
		if createRequest.Size != "" {
			droplet.Size.Slug = createRequest.Size
		}
		return &droplet, nil, nil
	}

	s.CreateMultipleFunc = func(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
		droplets := make([]godo.Droplet, len(createRequest.Names))
		for i, name := range createRequest.Names {
			droplet := *s.std.Droplets.DefaultDroplet
			droplet.ID += i // Increment ID for each droplet
			droplet.Name = name
			droplet.Region.Slug = createRequest.Region
			if createRequest.Size != "" {
				droplet.Size.Slug = createRequest.Size
			}
			droplets[i] = droplet
		}
		return droplets, nil, nil
	}

	s.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		return s.std.Droplets.DefaultDroplet, nil, nil
	}

	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		return s.std.Droplets.DefaultDropletList, nil, nil
	}

	s.DeleteFunc = func(ctx context.Context, id int) (*godo.Response, error) {
		return nil, nil
	}
}

// NewMockDropletService creates a new MockDropletService with standard responses
func NewMockDropletService(std *StandardResponses) *MockDropletService {
	s := &MockDropletService{std: std}
	setupStandardDropletResponses(s)
	return s
}

// ResetToStandard resets the droplet service back to standard success responses
func (s *MockDropletService) ResetToStandard() {
	setupStandardDropletResponses(s)
}

// Create calls the mocked Create function
func (s *MockDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	return s.CreateFunc(ctx, createRequest)
}

// CreateMultiple calls the mocked CreateMultiple function
func (s *MockDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	return s.CreateMultipleFunc(ctx, createRequest)
}

// Get calls the mocked Get function
func (s *MockDropletService) Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
	return s.GetFunc(ctx, id)
}

// Delete calls the mocked Delete function
func (s *MockDropletService) Delete(ctx context.Context, id int) (*godo.Response, error) {
	return s.DeleteFunc(ctx, id)
}

// List calls the mocked List function
func (s *MockDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	return s.ListFunc(ctx, opt)
}

// SimulateNotFound configures the service to return not found errors
func (s *MockDropletService) SimulateNotFound() {
	s.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.NotFoundError
	}
	s.DeleteFunc = func(ctx context.Context, id int) (*godo.Response, error) {
		return nil, s.std.Droplets.NotFoundError
	}
}

// SimulateRateLimit configures the service to return rate limit errors
func (s *MockDropletService) SimulateRateLimit() {
	s.CreateFunc = func(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.RateLimitError
	}
	s.CreateMultipleFunc = func(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.RateLimitError
	}
	s.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.RateLimitError
	}
	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.RateLimitError
	}
}

// SimulateAuthenticationFailure configures the service to return authentication errors
func (s *MockDropletService) SimulateAuthenticationFailure() {
	s.CreateFunc = func(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
	s.CreateMultipleFunc = func(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
	s.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
}

// SimulateDelayedSuccess configures the service to succeed after a specific number of attempts
func (s *MockDropletService) SimulateDelayedSuccess(successAfterAttempts int) {
	s.attemptCount = 0

	// For waitForIP testing
	s.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		s.attemptCount++
		if s.attemptCount >= successAfterAttempts {
			return s.std.Droplets.DefaultDroplet, nil, nil
		}
		// Return a droplet with no IP before success
		droplet := *s.std.Droplets.DefaultDroplet
		droplet.Networks = &godo.Networks{}
		return &droplet, nil, nil
	}

	// For waitForDeletion testing
	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		s.attemptCount++
		if s.attemptCount >= successAfterAttempts {
			// Return empty list to simulate deletion
			return []godo.Droplet{}, nil, nil
		}
		// Return list with the droplet still present
		return s.std.Droplets.DefaultDropletList, nil, nil
	}
}

// SimulateMaxRetries configures the service to always fail until max retries are hit
func (s *MockDropletService) SimulateMaxRetries() {
	s.attemptCount = 0

	// For waitForIP testing
	s.GetFunc = func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
		s.attemptCount++
		droplet := *s.std.Droplets.DefaultDroplet
		droplet.Networks = &godo.Networks{} // Always return no IP
		return &droplet, nil, nil
	}

	// For waitForDeletion testing
	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		s.attemptCount++
		// Always return list with the droplet present
		return s.std.Droplets.DefaultDropletList, nil, nil
	}
}

// GetAttemptCount returns the current attempt count
func (s *MockDropletService) GetAttemptCount() int {
	return s.attemptCount
}

// ResetAttemptCount resets the attempt counter
func (s *MockDropletService) ResetAttemptCount() {
	s.attemptCount = 0
}

// MockKeyService implements types.KeyService for testing
type MockKeyService struct {
	std      *StandardResponses
	ListFunc func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// setupStandardKeyResponses configures the standard success responses for key service
func setupStandardKeyResponses(s *MockKeyService) {
	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return s.std.Keys.DefaultKeyList, nil, nil
	}
}

// NewMockKeyService creates a new MockKeyService with standard responses
func NewMockKeyService(std *StandardResponses) *MockKeyService {
	s := &MockKeyService{std: std}
	setupStandardKeyResponses(s)
	return s
}

// ResetToStandard resets the key service back to standard success responses
func (s *MockKeyService) ResetToStandard() {
	setupStandardKeyResponses(s)
}

// List calls the mocked List function
func (s *MockKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	return s.ListFunc(ctx, opt)
}

// SimulateNotFound configures the service to return not found errors
func (s *MockKeyService) SimulateNotFound() {
	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return nil, nil, s.std.Keys.NotFoundError
	}
}

// SimulateRateLimit configures the service to return rate limit errors
func (s *MockKeyService) SimulateRateLimit() {
	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return nil, nil, s.std.Keys.RateLimitError
	}
}

// SimulateAuthenticationFailure configures the service to return authentication errors
func (s *MockKeyService) SimulateAuthenticationFailure() {
	s.ListFunc = func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return nil, nil, s.std.Keys.AuthenticationError
	}
}
