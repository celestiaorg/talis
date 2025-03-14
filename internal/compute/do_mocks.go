package compute

import (
	"context"

	"github.com/digitalocean/godo"
)

// mockDOClient implements DOClient for testing
type mockDOClient struct {
	mockDropletService *mockDropletService
	mockKeyService     *mockKeyService
}

// newMockDOClient creates a new mockDOClient
func newMockDOClient() *mockDOClient {
	return &mockDOClient{
		mockDropletService: newMockDropletService(),
		mockKeyService:     newMockKeyService(),
	}
}

// Droplets returns the mock droplet service
func (c *mockDOClient) Droplets() DropletService {
	return c.mockDropletService
}

// Keys returns the mock key service
func (c *mockDOClient) Keys() KeyService {
	return c.mockKeyService
}

// mockDropletService implements DropletService for testing
type mockDropletService struct {
	CreateFunc         func(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultipleFunc func(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	GetFunc            func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	DeleteFunc         func(ctx context.Context, id int) (*godo.Response, error)
	ListFunc           func(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
}

// newMockDropletService creates a new mockDropletService
func newMockDropletService() *mockDropletService {
	return &mockDropletService{}
}

// Create calls the mocked Create function
func (s *mockDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	if s.CreateFunc != nil {
		return s.CreateFunc(ctx, createRequest)
	}
	return nil, nil, nil
}

// CreateMultiple calls the mocked CreateMultiple function
func (s *mockDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	if s.CreateMultipleFunc != nil {
		return s.CreateMultipleFunc(ctx, createRequest)
	}
	return nil, nil, nil
}

// Get calls the mocked Get function
func (s *mockDropletService) Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
	if s.GetFunc != nil {
		return s.GetFunc(ctx, id)
	}
	return nil, nil, nil
}

// Delete calls the mocked Delete function
func (s *mockDropletService) Delete(ctx context.Context, id int) (*godo.Response, error) {
	if s.DeleteFunc != nil {
		return s.DeleteFunc(ctx, id)
	}
	return nil, nil
}

// List calls the mocked List function
func (s *mockDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	if s.ListFunc != nil {
		return s.ListFunc(ctx, opt)
	}
	return nil, nil, nil
}

// mockKeyService implements KeyService for testing
type mockKeyService struct {
	ListFunc func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// newMockKeyService creates a new mockKeyService
func newMockKeyService() *mockKeyService {
	return &mockKeyService{}
}

// List calls the mocked List function
func (s *mockKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	if s.ListFunc != nil {
		return s.ListFunc(ctx, opt)
	}
	return nil, nil, nil
}
