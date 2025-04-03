package mocks

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"

	"github.com/celestiaorg/talis/internal/types"
)

const (
	defaultMaxRetries = 3
)

// MockDOClient is a mock implementation of both the ComputeProvider and DOClient interfaces
type MockDOClient struct {
	droplets           []types.InstanceInfo
	MockDropletService *MockDropletService
	MockKeyService     *MockKeyService
	MockStorageService *MockStorageService
	StandardResponses  *StandardResponses
}

// NewMockDOClient creates a new mock DO client
func NewMockDOClient() *MockDOClient {
	client := &MockDOClient{
		droplets:          make([]types.InstanceInfo, 0),
		StandardResponses: newStandardResponses(),
	}

	// Initialize services
	client.MockDropletService = NewMockDropletService(client.StandardResponses)
	client.MockKeyService = NewMockKeyService(client.StandardResponses)
	client.MockStorageService = NewMockStorageService(client.StandardResponses)

	// Reset to ensure clean state
	client.ResetToStandard()

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
	// Check for errors first
	if m.MockDropletService.std.Droplets.AuthenticationError != nil {
		return nil, m.MockDropletService.std.Droplets.AuthenticationError
	}
	if m.MockDropletService.std.Droplets.RateLimitError != nil {
		return nil, m.MockDropletService.std.Droplets.RateLimitError
	}
	if m.MockDropletService.std.Droplets.NotFoundError != nil {
		return nil, m.MockDropletService.std.Droplets.NotFoundError
	}

	var instances []types.InstanceInfo
	for i := 0; i < config.NumberOfInstances; i++ {
		instanceName := fmt.Sprintf("%s-%d", name, i)
		instance := types.InstanceInfo{
			ID:       fmt.Sprintf("%d", DefaultDropletID1+i),
			Name:     instanceName,
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
	// Check for errors first
	if m.MockDropletService.std.Droplets.AuthenticationError != nil {
		return m.MockDropletService.std.Droplets.AuthenticationError
	}
	if m.MockDropletService.std.Droplets.RateLimitError != nil {
		return m.MockDropletService.std.Droplets.RateLimitError
	}
	if m.MockDropletService.std.Droplets.NotFoundError != nil {
		return m.MockDropletService.std.Droplets.NotFoundError
	}

	// Buscar la instancia por nombre y región
	var remaining []types.InstanceInfo
	found := false
	for _, instance := range m.droplets {
		if instance.Name == name {
			found = true
			continue
		}
		remaining = append(remaining, instance)
	}

	if !found {
		return ErrDropletNotFound
	}

	m.droplets = remaining
	return nil
}

// SimulateAuthenticationFailure simulates an authentication failure
func (m *MockDOClient) SimulateAuthenticationFailure() {
	// Reset all errors first
	m.ResetToStandard()

	// Set authentication errors
	m.MockDropletService.std.Droplets.AuthenticationError = ErrAuthentication
	m.MockKeyService.std.Keys.AuthenticationError = ErrAuthentication
	m.MockStorageService.std.Volumes.AuthenticationError = ErrAuthentication
}

// SimulateRateLimit simulates a rate limit error
func (m *MockDOClient) SimulateRateLimit() {
	// Reset all errors first
	m.ResetToStandard()

	// Set rate limit errors
	m.MockDropletService.std.Droplets.RateLimitError = ErrRateLimit
	m.MockKeyService.std.Keys.RateLimitError = ErrRateLimit
	m.MockStorageService.std.Volumes.RateLimitError = ErrRateLimit

	// Set retries for all services
	m.MockDropletService.retriesRemaining = 3
	m.MockKeyService.retriesRemaining = 3
	m.MockStorageService.retriesRemaining = 3
	m.MockDropletService.maxRetries = 3
	m.MockKeyService.maxRetries = 3
	m.MockStorageService.maxRetries = 3
}

// SimulateNotFound simulates a not found error
func (m *MockDOClient) SimulateNotFound() {
	// Reset all errors first
	m.ResetToStandard()

	// Set not found errors
	m.MockDropletService.std.Droplets.NotFoundError = ErrDropletNotFound
	m.MockKeyService.std.Keys.NotFoundError = ErrKeyNotFound
	m.MockStorageService.std.Volumes.NotFoundError = ErrVolumeNotFound
}

// SimulateDelayedSuccess simulates a success after a number of retries for all services
func (m *MockDOClient) SimulateDelayedSuccess(retries int) {
	m.MockDropletService.SimulateDelayedSuccess(retries)
	m.MockKeyService.SimulateDelayedSuccess(retries)
	m.MockStorageService.SimulateDelayedSuccess(retries)
}

// SimulateMaxRetries simulates reaching the maximum number of retries for all services
func (m *MockDOClient) SimulateMaxRetries() {
	m.MockDropletService.SimulateMaxRetries()
	m.MockKeyService.SimulateMaxRetries()
	m.MockStorageService.SimulateMaxRetries()
}

// ResetToStandard resets the mock to standard responses
func (m *MockDOClient) ResetToStandard() {
	// Reset droplets list
	m.droplets = make([]types.InstanceInfo, 0)

	// Reset all errors to nil
	m.StandardResponses.Droplets.AuthenticationError = nil
	m.StandardResponses.Droplets.RateLimitError = nil
	m.StandardResponses.Droplets.NotFoundError = nil

	m.StandardResponses.Keys.AuthenticationError = nil
	m.StandardResponses.Keys.RateLimitError = nil
	m.StandardResponses.Keys.NotFoundError = nil

	m.StandardResponses.Volumes.AuthenticationError = nil
	m.StandardResponses.Volumes.RateLimitError = nil
	m.StandardResponses.Volumes.NotFoundError = nil

	// Reset retry counters
	m.MockDropletService.retriesRemaining = 0
	m.MockDropletService.maxRetries = 0
	m.MockKeyService.retriesRemaining = 0
	m.MockKeyService.maxRetries = 0
	m.MockStorageService.retriesRemaining = 0
	m.MockStorageService.maxRetries = 0
}

// MockDropletService implements types.DropletService for testing
type MockDropletService struct {
	std              *StandardResponses
	retriesRemaining int
	maxRetries       int
}

// SimulateDelayedSuccess simulates a success after a number of retries
func (s *MockDropletService) SimulateDelayedSuccess(retries int) {
	s.retriesRemaining = retries
	s.maxRetries = retries
}

// SimulateMaxRetries simulates reaching the maximum number of retries
func (s *MockDropletService) SimulateMaxRetries() {
	s.retriesRemaining = defaultMaxRetries
	s.maxRetries = defaultMaxRetries
}

// Create creates a new droplet
func (s *MockDropletService) Create(ctx context.Context, createRequest *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	if s.std.Droplets.AuthenticationError != nil {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
	if s.std.Droplets.RateLimitError != nil {
		return nil, nil, s.std.Droplets.RateLimitError
	}
	if s.std.Droplets.NotFoundError != nil {
		return nil, nil, s.std.Droplets.NotFoundError
	}

	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		return nil, nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	droplet := *s.std.Droplets.DefaultDroplet
	droplet.Name = createRequest.Name
	return &droplet, nil, nil
}

// CreateMultiple creates multiple droplets
func (s *MockDropletService) CreateMultiple(ctx context.Context, createRequest *godo.DropletMultiCreateRequest) ([]godo.Droplet, *godo.Response, error) {
	if s.std.Droplets.AuthenticationError != nil {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
	if s.std.Droplets.RateLimitError != nil {
		return nil, nil, s.std.Droplets.RateLimitError
	}
	if s.std.Droplets.NotFoundError != nil {
		return nil, nil, s.std.Droplets.NotFoundError
	}

	var droplets []godo.Droplet
	for i, name := range createRequest.Names {
		droplet := *s.std.Droplets.DefaultDroplet
		droplet.ID = DefaultDropletID1 + i
		droplet.Name = name
		droplets = append(droplets, droplet)
	}
	return droplets, nil, nil
}

// Get gets a droplet
func (s *MockDropletService) Get(ctx context.Context, dropletID int) (*godo.Droplet, *godo.Response, error) {
	if s.std.Droplets.AuthenticationError != nil {
		return nil, nil, s.std.Droplets.AuthenticationError
	}

	// If we have retries remaining, simulate a rate limit error
	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		if s.std.Droplets.RateLimitError != nil {
			return nil, nil, s.std.Droplets.RateLimitError
		}
		return nil, nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	// If not found error is set, return it
	if s.std.Droplets.NotFoundError != nil {
		return nil, nil, s.std.Droplets.NotFoundError
	}

	// If no errors, return success
	return s.std.Droplets.DefaultDroplet, nil, nil
}

// Delete deletes a droplet
func (s *MockDropletService) Delete(ctx context.Context, dropletID int) (*godo.Response, error) {
	if s.std.Droplets.AuthenticationError != nil {
		return nil, s.std.Droplets.AuthenticationError
	}

	// If we have retries remaining, simulate a rate limit error
	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		if s.std.Droplets.RateLimitError != nil {
			return nil, s.std.Droplets.RateLimitError
		}
		return nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	// If not found error is set, return it
	if s.std.Droplets.NotFoundError != nil {
		return nil, s.std.Droplets.NotFoundError
	}

	// If no errors, return success
	return nil, nil
}

// List lists all droplets
func (s *MockDropletService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
	if s.std.Droplets.AuthenticationError != nil {
		return nil, nil, s.std.Droplets.AuthenticationError
	}

	// If we have retries remaining, simulate a rate limit error
	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		if s.std.Droplets.RateLimitError != nil {
			return nil, nil, s.std.Droplets.RateLimitError
		}
		return nil, nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	// If not found error is set, return it
	if s.std.Droplets.NotFoundError != nil {
		return nil, nil, s.std.Droplets.NotFoundError
	}

	// If no errors, return success
	return s.std.Droplets.DefaultDropletList, nil, nil
}

// MockKeyService implements types.KeyService for testing
type MockKeyService struct {
	std              *StandardResponses
	retriesRemaining int
	maxRetries       int
}

// SimulateDelayedSuccess simulates a success after a number of retries
func (s *MockKeyService) SimulateDelayedSuccess(retries int) {
	s.retriesRemaining = retries
	s.maxRetries = retries
}

// SimulateMaxRetries simulates reaching the maximum number of retries
func (s *MockKeyService) SimulateMaxRetries() {
	s.retriesRemaining = defaultMaxRetries
	s.maxRetries = defaultMaxRetries
}

// List lists all SSH keys
func (s *MockKeyService) List(ctx context.Context, opt *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
	if s.std.Keys.AuthenticationError != nil {
		return nil, nil, s.std.Keys.AuthenticationError
	}

	// If we have retries remaining, simulate a rate limit error
	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		// Return rate limit error for all but the last retry
		if s.retriesRemaining > 0 {
			return nil, nil, s.std.Keys.RateLimitError
		}
	}

	if s.std.Keys.NotFoundError != nil {
		return nil, nil, s.std.Keys.NotFoundError
	}

	return s.std.Keys.DefaultKeyList, nil, nil
}

// MockStorageService implements types.StorageService for testing
type MockStorageService struct {
	std              *StandardResponses
	retriesRemaining int
	maxRetries       int
}

// SimulateDelayedSuccess simulates a success after a number of retries
func (s *MockStorageService) SimulateDelayedSuccess(retries int) {
	s.retriesRemaining = retries
	s.maxRetries = retries
}

// SimulateMaxRetries simulates reaching the maximum number of retries
func (s *MockStorageService) SimulateMaxRetries() {
	s.retriesRemaining = defaultMaxRetries
	s.maxRetries = defaultMaxRetries
}

// CreateVolume creates a new volume
func (s *MockStorageService) CreateVolume(ctx context.Context, request *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
	if s.std.Volumes.AuthenticationError != nil {
		return nil, nil, s.std.Volumes.AuthenticationError
	}
	if s.std.Volumes.RateLimitError != nil {
		return nil, nil, s.std.Volumes.RateLimitError
	}
	if s.std.Volumes.NotFoundError != nil {
		return nil, nil, s.std.Volumes.NotFoundError
	}

	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		return nil, nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	return s.std.Volumes.DefaultVolume, nil, nil
}

// DeleteVolume deletes a volume
func (s *MockStorageService) DeleteVolume(ctx context.Context, id string) (*godo.Response, error) {
	if s.std.Volumes.AuthenticationError != nil {
		return nil, s.std.Volumes.AuthenticationError
	}
	if s.std.Volumes.RateLimitError != nil {
		return nil, s.std.Volumes.RateLimitError
	}
	if s.std.Volumes.NotFoundError != nil {
		return nil, s.std.Volumes.NotFoundError
	}

	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		return nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	return nil, nil
}

// ListVolumes lists all volumes
func (s *MockStorageService) ListVolumes(ctx context.Context, opt *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
	if s.std.Volumes.AuthenticationError != nil {
		return nil, nil, s.std.Volumes.AuthenticationError
	}
	if s.std.Volumes.RateLimitError != nil {
		return nil, nil, s.std.Volumes.RateLimitError
	}
	if s.std.Volumes.NotFoundError != nil {
		return nil, nil, s.std.Volumes.NotFoundError
	}

	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		return nil, nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	return s.std.Volumes.DefaultVolumeList, nil, nil
}

// GetVolume gets a volume
func (s *MockStorageService) GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error) {
	if s.std.Volumes.AuthenticationError != nil {
		return nil, nil, s.std.Volumes.AuthenticationError
	}
	if s.std.Volumes.RateLimitError != nil {
		return nil, nil, s.std.Volumes.RateLimitError
	}
	if s.std.Volumes.NotFoundError != nil {
		return nil, nil, s.std.Volumes.NotFoundError
	}

	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		return nil, nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	return s.std.Volumes.DefaultVolume, nil, nil
}

// GetVolumeAction gets a volume action
func (s *MockStorageService) GetVolumeAction(ctx context.Context, volumeID string, actionID int) (*godo.Action, *godo.Response, error) {
	if s.std.Volumes.AuthenticationError != nil {
		return nil, nil, s.std.Volumes.AuthenticationError
	}
	if s.std.Volumes.RateLimitError != nil {
		return nil, nil, s.std.Volumes.RateLimitError
	}
	if s.std.Volumes.NotFoundError != nil {
		return nil, nil, s.std.Volumes.NotFoundError
	}

	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		return nil, nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	return &godo.Action{
		ID:     actionID,
		Status: "completed",
	}, nil, nil
}

// AttachVolume attaches a volume to a droplet
func (s *MockStorageService) AttachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error) {
	if s.std.Volumes.AuthenticationError != nil {
		return nil, s.std.Volumes.AuthenticationError
	}
	if s.std.Volumes.RateLimitError != nil {
		return nil, s.std.Volumes.RateLimitError
	}
	if s.std.Volumes.NotFoundError != nil {
		return nil, s.std.Volumes.NotFoundError
	}

	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		return nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	return nil, nil
}

// DetachVolume detaches a volume from a droplet
func (s *MockStorageService) DetachVolume(ctx context.Context, volumeID string, dropletID int) (*godo.Response, error) {
	if s.std.Volumes.AuthenticationError != nil {
		return nil, s.std.Volumes.AuthenticationError
	}
	if s.std.Volumes.RateLimitError != nil {
		return nil, s.std.Volumes.RateLimitError
	}
	if s.std.Volumes.NotFoundError != nil {
		return nil, s.std.Volumes.NotFoundError
	}

	if s.retriesRemaining > 0 {
		s.retriesRemaining--
		return nil, fmt.Errorf("simulated retry %d/%d", s.maxRetries-s.retriesRemaining, s.maxRetries)
	}

	return nil, nil
}

// NewMockDropletService creates a new mock droplet service
func NewMockDropletService(std *StandardResponses) *MockDropletService {
	return &MockDropletService{
		std:              std,
		retriesRemaining: 0,
		maxRetries:       0,
	}
}

// NewMockKeyService creates a new mock key service
func NewMockKeyService(std *StandardResponses) *MockKeyService {
	return &MockKeyService{
		std:              std,
		retriesRemaining: 0,
		maxRetries:       0,
	}
}

// NewMockStorageService creates a new mock storage service
func NewMockStorageService(std *StandardResponses) *MockStorageService {
	return &MockStorageService{
		std:              std,
		retriesRemaining: 0,
		maxRetries:       0,
	}
}
