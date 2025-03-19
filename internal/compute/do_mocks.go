package compute

import (
	"context"

	"github.com/digitalocean/godo"
)

// DropletService defines the interface for droplet operations
type DropletService interface {
	Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	Delete(ctx context.Context, id int) (*godo.Response, error)
}

// KeyService defines the interface for SSH key operations
type KeyService interface {
	List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
	Get(ctx context.Context, id int) (*godo.Key, *godo.Response, error)
	GetByFingerprint(ctx context.Context, fingerprint string) (*godo.Key, *godo.Response, error)
}

// DOClient defines the interface for DigitalOcean client operations
type DOClient interface {
	Droplets() DropletService
	Keys() KeyService
}

// mockDOClient implements the DOClient interface
type mockDOClient struct {
	mockDropletService *mockDropletService
	mockKeyService     *mockKeyService
}

// NewMockDOClient creates a new mock DigitalOcean client
func NewMockDOClient() DOClient {
	return &mockDOClient{
		mockDropletService: &mockDropletService{
			droplets: make(map[int]*godo.Droplet),
		},
		mockKeyService: &mockKeyService{
			keys: make(map[int]*godo.Key),
		},
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

// mockDropletService implements the DropletService interface
type mockDropletService struct {
	droplets map[int]*godo.Droplet
}

// mockKeyService implements the KeyService interface
type mockKeyService struct {
	keys map[int]*godo.Key
}

// Create mocks the Create method of DropletService
func (m *mockDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	// Mock implementation
	droplet := &godo.Droplet{
		ID:     1,
		Name:   createRequest.Name,
		Region: &godo.Region{Slug: createRequest.Region},
		Size:   &godo.Size{Slug: createRequest.Size},
		Image:  &godo.Image{Slug: createRequest.Image.Slug},
	}
	m.droplets[droplet.ID] = droplet
	return droplet, &godo.Response{}, nil
}

// CreateMultiple mocks the CreateMultiple method of DropletService
func (m *mockDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	// Mock implementation
	var droplets []godo.Droplet
	for i, name := range createRequest.Names {
		droplet := godo.Droplet{
			ID:     i + 1,
			Name:   name,
			Region: &godo.Region{Slug: createRequest.Region},
			Size:   &godo.Size{Slug: createRequest.Size},
			Image:  &godo.Image{Slug: createRequest.Image.Slug},
		}
		m.droplets[droplet.ID] = &droplet
		droplets = append(droplets, droplet)
	}
	return droplets, &godo.Response{}, nil
}

// Get mocks the Get method of DropletService
func (m *mockDropletService) Get(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error) {
	// Mock implementation
	if droplet, ok := m.droplets[id]; ok {
		return droplet, &godo.Response{}, nil
	}
	return nil, &godo.Response{}, &godo.ErrorResponse{
		Response: nil,
		Message:  "droplet not found",
	}
}

// List mocks the List method of DropletService
func (m *mockDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	// Mock implementation
	var droplets []godo.Droplet
	for _, droplet := range m.droplets {
		droplets = append(droplets, *droplet)
	}
	return droplets, &godo.Response{}, nil
}

// Delete mocks the Delete method of DropletService
func (m *mockDropletService) Delete(ctx context.Context, id int) (*godo.Response, error) {
	// Mock implementation
	delete(m.droplets, id)
	return &godo.Response{}, nil
}

// List mocks the List method of KeyService
func (m *mockKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	// Mock implementation
	var keys []godo.Key
	for _, key := range m.keys {
		keys = append(keys, *key)
	}
	return keys, &godo.Response{}, nil
}

// Get mocks the Get method of KeyService
func (m *mockKeyService) Get(ctx context.Context, id int) (*godo.Key, *godo.Response, error) {
	// Mock implementation
	if key, ok := m.keys[id]; ok {
		return key, &godo.Response{}, nil
	}
	return nil, &godo.Response{}, &godo.ErrorResponse{
		Response: nil,
		Message:  "key not found",
	}
}

// GetByFingerprint mocks the GetByFingerprint method of KeyService
func (m *mockKeyService) GetByFingerprint(ctx context.Context, fingerprint string) (*godo.Key, *godo.Response, error) {
	// Mock implementation
	for _, key := range m.keys {
		if key.Fingerprint == fingerprint {
			return key, &godo.Response{}, nil
		}
	}
	return nil, &godo.Response{}, &godo.ErrorResponse{
		Response: nil,
		Message:  "key not found",
	}
}
