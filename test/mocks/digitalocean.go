package mocks

import (
	"context"
	"fmt"
	"os"

	"github.com/celestiaorg/talis/internal/types"
	"github.com/digitalocean/godo"
)

// MockDOClient is a mock implementation of both the ComputeProvider and DOClient interfaces
type MockDOClient struct {
	droplets           []types.InstanceInfo
	MockDropletService *MockDropletService
	MockKeyService     *MockKeyService
	MockStorageService *MockStorageService
}

// NewMockDOClient creates a new mock DO client
func NewMockDOClient() *MockDOClient {
	std := newStandardResponses()
	client := &MockDOClient{
		droplets:           make([]types.InstanceInfo, 0),
		MockDropletService: NewMockDropletService(std),
		MockKeyService:     NewMockKeyService(std),
		MockStorageService: NewMockStorageService(std),
	}
	return client
}

// Droplets returns the mock droplet service
func (m *MockDOClient) Droplets() types.DropletService {
	return m.MockDropletService
}

// Keys returns the mock key service
func (m *MockDOClient) Keys() types.KeyService {
	return m.MockKeyService
}

// Storage returns the mock storage service
func (m *MockDOClient) Storage() types.StorageService {
	return m.MockStorageService
}

// ValidateCredentials validates the provider credentials
func (m *MockDOClient) ValidateCredentials() error {
	return nil
}

// GetEnvironmentVars returns the environment variables needed for the provider
func (m *MockDOClient) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"DIGITALOCEAN_TOKEN": os.Getenv("DIGITALOCEAN_TOKEN"),
	}
}

// ConfigureProvider configures the provider with the given stack
func (m *MockDOClient) ConfigureProvider(stack interface{}) error {
	return nil
}

// CreateInstance creates a new instance
func (m *MockDOClient) CreateInstance(ctx context.Context, name string, config types.InstanceConfig) ([]types.InstanceInfo, error) {
	var instances []types.InstanceInfo
	for i := 0; i < config.NumberOfInstances; i++ {
		instance := types.InstanceInfo{
			ID:       fmt.Sprintf("mock-instance-%d", i),
			Name:     fmt.Sprintf("%s-%d", name, i),
			Provider: "digitalocean-mock",
			Region:   config.Region,
			Size:     config.Size,
			PublicIP: fmt.Sprintf("192.168.1.%d", i+100),
		}
		instances = append(instances, instance)
		m.droplets = append(m.droplets, instance)
	}
	return instances, nil
}

// DeleteInstance deletes an instance
func (m *MockDOClient) DeleteInstance(ctx context.Context, name string, region string) error {
	var remaining []types.InstanceInfo
	for _, instance := range m.droplets {
		if instance.Name != name || instance.Region != region {
			remaining = append(remaining, instance)
		}
	}
	m.droplets = remaining
	return nil
}

// MockDropletService implements types.DropletService for testing
type MockDropletService struct {
	std *StandardResponses
}

// Create creates a new droplet
func (s *MockDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	return s.std.Droplets.DefaultDroplet, nil, nil
}

// CreateMultiple creates multiple droplets
func (s *MockDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	return s.std.Droplets.DefaultDropletList, nil, nil
}

// Delete deletes a droplet
func (s *MockDropletService) Delete(ctx context.Context, dropletID int) (*godo.Response, error) {
	return nil, nil
}

// Get gets a droplet
func (s *MockDropletService) Get(ctx context.Context, dropletID int) (*godo.Droplet, *godo.Response, error) {
	return s.std.Droplets.DefaultDroplet, nil, nil
}

// List lists all droplets
func (s *MockDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	return s.std.Droplets.DefaultDropletList, nil, nil
}

// MockKeyService implements types.KeyService for testing
type MockKeyService struct {
	std *StandardResponses
}

// List lists all SSH keys
func (s *MockKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	return s.std.Keys.DefaultKeyList, nil, nil
}

// MockStorageService implements types.StorageService for testing
type MockStorageService struct {
	std *StandardResponses
}

// CreateVolume creates a new volume
func (s *MockStorageService) CreateVolume(ctx context.Context, request *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
	return s.std.Volumes.DefaultVolume, nil, nil
}

// DeleteVolume deletes a volume
func (s *MockStorageService) DeleteVolume(ctx context.Context, id string) (*godo.Response, error) {
	return nil, nil
}

// ListVolumes lists all volumes
func (s *MockStorageService) ListVolumes(ctx context.Context, opt *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
	return s.std.Volumes.DefaultVolumeList, nil, nil
}

// GetVolume gets a volume
func (s *MockStorageService) GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error) {
	return s.std.Volumes.DefaultVolume, nil, nil
}

// GetVolumeAction gets a volume action
func (s *MockStorageService) GetVolumeAction(ctx context.Context, volumeID string, actionID int) (*godo.Action, *godo.Response, error) {
	return &godo.Action{
		ID:     actionID,
		Status: "completed",
	}, nil, nil
}

// AttachVolume attaches a volume to a droplet
func (s *MockStorageService) AttachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error) {
	return nil, nil
}

// DetachVolume detaches a volume from a droplet
func (s *MockStorageService) DetachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error) {
	return nil, nil
}

// NewMockDropletService creates a new mock droplet service
func NewMockDropletService(std *StandardResponses) *MockDropletService {
	return &MockDropletService{std: std}
}

// NewMockKeyService creates a new mock key service
func NewMockKeyService(std *StandardResponses) *MockKeyService {
	return &MockKeyService{std: std}
}

// NewMockStorageService creates a new mock storage service
func NewMockStorageService(std *StandardResponses) *MockStorageService {
	return &MockStorageService{std: std}
}
