package compute

import (
	"context"

	"github.com/digitalocean/godo"
)

// MockDOClient implements DOClient for testing
type MockDOClient struct {
	MockDropletService *MockDropletService
	MockKeyService     *MockKeyService
}

// NewMockDOClient creates a new MockDOClient
func NewMockDOClient() *MockDOClient {
	return &MockDOClient{
		MockDropletService: NewMockDropletService(),
		MockKeyService:     NewMockKeyService(),
	}
}

// Droplets returns the mock droplet service
func (c *MockDOClient) Droplets() DropletService {
	return c.MockDropletService
}

// Keys returns the mock key service
func (c *MockDOClient) Keys() KeyService {
	return c.MockKeyService
}

// MockDropletService implements DropletService for testing
type MockDropletService struct {
	CreateFunc         func(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultipleFunc func(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	GetFunc            func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	DeleteFunc         func(ctx context.Context, id int) (*godo.Response, error)
	ListFunc           func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

// NewMockDropletService creates a new MockDropletService
func NewMockDropletService() *MockDropletService {
	return &MockDropletService{}
}

// Create calls the mocked Create function
func (s *MockDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	if s.CreateFunc != nil {
		return s.CreateFunc(ctx, createRequest)
	}
	return nil, nil, nil
}

// CreateMultiple calls the mocked CreateMultiple function
func (s *MockDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	if s.CreateMultipleFunc != nil {
		return s.CreateMultipleFunc(ctx, createRequest)
	}
	return nil, nil, nil
}

// Get calls the mocked Get function
func (s *MockDropletService) Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
	if s.GetFunc != nil {
		return s.GetFunc(ctx, id)
	}
	return nil, nil, nil
}

// Delete calls the mocked Delete function
func (s *MockDropletService) Delete(ctx context.Context, id int) (*godo.Response, error) {
	if s.DeleteFunc != nil {
		return s.DeleteFunc(ctx, id)
	}
	return nil, nil
}

// List calls the mocked List function
func (s *MockDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	if s.ListFunc != nil {
		return s.ListFunc(ctx, opt)
	}
	return nil, nil, nil
}

// MockKeyService implements KeyService for testing
type MockKeyService struct {
	ListFunc func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// NewMockKeyService creates a new MockKeyService
func NewMockKeyService() *MockKeyService {
	return &MockKeyService{}
}

// List calls the mocked List function
func (s *MockKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	if s.ListFunc != nil {
		return s.ListFunc(ctx, opt)
	}
	return nil, nil, nil
}
