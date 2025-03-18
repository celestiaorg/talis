// Package compute provides mock implementations for compute providers
package compute

import (
	"context"

	"github.com/digitalocean/godo"
)

// MockDOClient is a mock implementation of the DigitalOcean client.
// This is a placeholder that will be expanded in Task 2.2.
type MockDOClient struct {
	// Mock services
	MockDropletService  MockDropletService
	MockKeyService      MockKeyService
	MockSnapshotService MockSnapshotService
}

// MockDropletService mocks the DigitalOcean Droplet service.
type MockDropletService struct {
	CreateFunc func(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	GetFunc    func(ctx context.Context, id int) (*godo.Droplet, *godo.Response, error)
}

// MockKeyService mocks the DigitalOcean Key service.
type MockKeyService struct {
	ListFunc func(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// MockSnapshotService mocks the DigitalOcean Snapshot service.
type MockSnapshotService struct {
	ListFunc func(ctx context.Context, opt *godo.ListOptions) ([]godo.Snapshot, *godo.Response, error)
}

// NewMockDOClient creates a new mock DigitalOcean client.
func NewMockDOClient() *MockDOClient {
	return &MockDOClient{}
}
