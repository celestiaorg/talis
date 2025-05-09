// Package mocks provides mock implementations for external services and APIs used in testing
package mocks

import (
	"context"
	"fmt"
	"os"

	"github.com/digitalocean/godo"

	computeTypes "github.com/celestiaorg/talis/internal/compute/types"
	talisTypes "github.com/celestiaorg/talis/internal/types"
)

// This file contains all the mock implementations for the DigitalOcean API and helper methods

// MockDOClient implements types.DOClient for testing
type MockDOClient struct {
	droplets           []talisTypes.InstanceRequest
	MockDropletService *MockDropletService
	MockKeyService     *MockKeyService
	MockStorageService *MockStorageService
	StandardResponses  *StandardResponses
}

// ConfigureProvider is a no-op to satisfy the ComputeProvider interface
func (c *MockDOClient) ConfigureProvider(_ interface{}) error {
	return nil
}

// CreateInstance is a mock implementation of the CreateInstance method
func (c *MockDOClient) CreateInstance(ctx context.Context, config *talisTypes.InstanceRequest) error {
	// config.Name is no longer available. Generate a mock droplet name using ProjectName.
	// For actual DO, uniqueness per account/region is needed. For mock, this might be sufficient.
	dropletName := fmt.Sprintf("%s-mock-droplet-test-0", config.ProjectName)
	createRequest := createDropletRequest(dropletName, *config, DefaultKeyID1)
	_, _, err := c.MockDropletService.Create(ctx, createRequest)
	if err != nil {
		return err
	}
	return nil
}

// DeleteInstance is a mock implementation of the DeleteInstance method
func (c *MockDOClient) DeleteInstance(ctx context.Context, dropletID int) error {
	_, err := c.MockDropletService.Delete(ctx, dropletID)
	return err
}

// GetEnvironmentVars is a no-op to satisfy the ComputeProvider interface
func (c *MockDOClient) GetEnvironmentVars() map[string]string {
	return map[string]string{
		"DIGITALOCEAN_TOKEN": os.Getenv("DIGITALOCEAN_TOKEN"),
	}
}

// ValidateCredentials is a no-op to satisfy the ComputeProvider interface
func (c *MockDOClient) ValidateCredentials() error {
	return nil
}

// createDropletRequest is a helper function to create a DropletCreateRequest
func createDropletRequest(
	name string,
	config talisTypes.InstanceRequest,
	sshKeyID int,
) *godo.DropletCreateRequest {
	return &godo.DropletCreateRequest{
		Name:   name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: config.Image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{ID: sshKeyID},
		},
		Tags: append([]string{name}, config.Tags...),
	}
}

// NewMockDOClient creates a new MockDOClient with standard responses
func NewMockDOClient() *MockDOClient {
	client := &MockDOClient{
		droplets:          make([]talisTypes.InstanceRequest, 0),
		StandardResponses: newStandardResponses(),
	}

	client.MockDropletService = NewMockDropletService(client.StandardResponses)
	client.MockKeyService = NewMockKeyService(client.StandardResponses)
	client.MockStorageService = NewMockStorageService(client.StandardResponses)

	return client
}

// ResetToStandard resets all mock services back to their standard success responses
func (c *MockDOClient) ResetToStandard() {
	c.MockDropletService.ResetToStandard()
	c.MockKeyService.ResetToStandard()
	c.MockStorageService.ResetToStandard()
}

// Droplets returns the mock droplet service
func (c *MockDOClient) Droplets() computeTypes.DropletService {
	return c.MockDropletService
}

// Keys returns the mock key service
func (c *MockDOClient) Keys() computeTypes.KeyService {
	return c.MockKeyService
}

// Storage returns the mock storage service
func (c *MockDOClient) Storage() computeTypes.StorageService {
	return c.MockStorageService
}

// SimulateAuthenticationFailure configures all services to return authentication errors
func (c *MockDOClient) SimulateAuthenticationFailure() {
	c.MockDropletService.SimulateAuthenticationFailure()
	c.MockKeyService.SimulateAuthenticationFailure()
	c.MockStorageService.SimulateAuthenticationFailure()
}

// SimulateNotFound configures all services to return not found errors
func (c *MockDOClient) SimulateNotFound() {
	c.MockDropletService.SimulateNotFound()
	c.MockKeyService.SimulateNotFound()
	c.MockStorageService.SimulateNotFound()
}

// SimulateRateLimit configures all services to return rate limit errors
func (c *MockDOClient) SimulateRateLimit() {
	c.MockDropletService.SimulateRateLimit()
	c.MockKeyService.SimulateRateLimit()
	c.MockStorageService.SimulateRateLimit()
}

// MockDropletService implements types.DropletService for testing
type MockDropletService struct {
	std          *StandardResponses
	CreateFunc   func(_ context.Context, _ *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	GetFunc      func(_ context.Context, _ int) (*godo.Droplet, *godo.Response, error)
	DeleteFunc   func(_ context.Context, _ int) (*godo.Response, error)
	ListFunc     func(_ context.Context, opt *godo.ListOptions) ([]godo.Droplet, *godo.Response, error)
	attemptCount int // Track number of attempts for retry simulations
}

// setupStandardDropletResponses configures the standard success responses for droplet service
func setupStandardDropletResponses(s *MockDropletService) {
	s.CreateFunc = func(_ context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		droplet := *s.std.Droplets.DefaultDroplet // Create a copy
		droplet.Name = req.Name                   // Use requested name
		droplet.Region.Slug = req.Region          // Use requested region
		if req.Size != "" {
			droplet.Size.Slug = req.Size
		}
		return &droplet, nil, nil
	}

	s.GetFunc = func(_ context.Context, _ int) (*godo.Droplet, *godo.Response, error) {
		return s.std.Droplets.DefaultDroplet, nil, nil
	}

	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		return s.std.Droplets.DefaultDropletList, nil, nil
	}

	s.DeleteFunc = func(_ context.Context, _ int) (*godo.Response, error) {
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
func (s *MockDropletService) Create(ctx context.Context, req *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
	return s.CreateFunc(ctx, req)
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
	s.GetFunc = func(_ context.Context, _ int) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.NotFoundError
	}
	s.DeleteFunc = func(_ context.Context, _ int) (*godo.Response, error) {
		return nil, s.std.Droplets.NotFoundError
	}
}

// SimulateRateLimit configures the service to return rate limit errors
func (s *MockDropletService) SimulateRateLimit() {
	s.CreateFunc = func(_ context.Context, _ *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.RateLimitError
	}
	s.GetFunc = func(_ context.Context, _ int) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.RateLimitError
	}
	s.DeleteFunc = func(_ context.Context, _ int) (*godo.Response, error) {
		return nil, s.std.Droplets.RateLimitError
	}
	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.RateLimitError
	}
}

// SimulateAuthenticationFailure configures the service to return authentication errors
func (s *MockDropletService) SimulateAuthenticationFailure() {
	s.CreateFunc = func(_ context.Context, _ *godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
	s.GetFunc = func(_ context.Context, _ int) (*godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
	s.DeleteFunc = func(_ context.Context, _ int) (*godo.Response, error) {
		return nil, s.std.Droplets.AuthenticationError
	}
	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
		return nil, nil, s.std.Droplets.AuthenticationError
	}
}

// SimulateDelayedSuccess configures the service to succeed after a specific number of attempts
func (s *MockDropletService) SimulateDelayedSuccess(successAfterAttempts int) {
	s.attemptCount = 0

	// For waitForIP testing
	s.GetFunc = func(_ context.Context, _ int) (*godo.Droplet, *godo.Response, error) {
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
	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
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
	s.GetFunc = func(_ context.Context, _ int) (*godo.Droplet, *godo.Response, error) {
		s.attemptCount++
		droplet := *s.std.Droplets.DefaultDroplet
		droplet.Networks = &godo.Networks{} // Always return no IP
		return &droplet, nil, nil
	}

	// For waitForDeletion testing
	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Droplet, *godo.Response, error) {
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
	ListFunc func(_ context.Context, _ *godo.ListOptions) ([]godo.Key, *godo.Response, error)
}

// setupStandardKeyResponses configures the standard success responses for key service
func setupStandardKeyResponses(s *MockKeyService) {
	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
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
	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return nil, nil, s.std.Keys.NotFoundError
	}
}

// SimulateRateLimit configures the service to return rate limit errors
func (s *MockKeyService) SimulateRateLimit() {
	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return nil, nil, s.std.Keys.RateLimitError
	}
}

// SimulateAuthenticationFailure configures the service to return authentication errors
func (s *MockKeyService) SimulateAuthenticationFailure() {
	s.ListFunc = func(_ context.Context, _ *godo.ListOptions) ([]godo.Key, *godo.Response, error) {
		return nil, nil, s.std.Keys.AuthenticationError
	}
}

// MockStorageService implements types.StorageService for testing
type MockStorageService struct {
	std                 *StandardResponses
	CreateVolumeFunc    func(_ context.Context, _ *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error)
	DeleteVolumeFunc    func(_ context.Context, _ string) (*godo.Response, error)
	ListVolumesFunc     func(_ context.Context, _ *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error)
	GetVolumeFunc       func(_ context.Context, _ string) (*godo.Volume, *godo.Response, error)
	GetVolumeActionFunc func(_ context.Context, _ string, _ int) (*godo.Action, *godo.Response, error)
	AttachVolumeFunc    func(_ context.Context, _ string, _ int) (*godo.Response, error)
	DetachVolumeFunc    func(_ context.Context, _ string, _ int) (*godo.Response, error)
	attemptCount        int // Track number of attempts for retry simulations
}

// setupStandardStorageResponses configures the standard success responses for storage service
func setupStandardStorageResponses(s *MockStorageService) {
	s.CreateVolumeFunc = func(_ context.Context, _ *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
		return s.std.Volumes.DefaultVolume, nil, nil
	}
	s.DeleteVolumeFunc = func(_ context.Context, _ string) (*godo.Response, error) {
		return nil, nil
	}
	s.ListVolumesFunc = func(_ context.Context, _ *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
		return s.std.Volumes.DefaultVolumeList, nil, nil
	}
	s.GetVolumeFunc = func(_ context.Context, _ string) (*godo.Volume, *godo.Response, error) {
		return s.std.Volumes.DefaultVolume, nil, nil
	}
	s.GetVolumeActionFunc = func(_ context.Context, _ string, _ int) (*godo.Action, *godo.Response, error) {
		return nil, nil, nil
	}
	s.AttachVolumeFunc = func(_ context.Context, _ string, _ int) (*godo.Response, error) {
		return nil, nil
	}
	s.DetachVolumeFunc = func(_ context.Context, _ string, _ int) (*godo.Response, error) {
		return nil, nil
	}
}

// NewMockStorageService creates a new MockStorageService with standard responses
func NewMockStorageService(std *StandardResponses) *MockStorageService {
	s := &MockStorageService{std: std}
	setupStandardStorageResponses(s)
	return s
}

// ResetToStandard resets the storage service back to standard success responses
func (s *MockStorageService) ResetToStandard() {
	setupStandardStorageResponses(s)
}

// CreateVolume calls the mocked CreateVolume function
func (s *MockStorageService) CreateVolume(ctx context.Context, req *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
	return s.CreateVolumeFunc(ctx, req)
}

// DeleteVolume calls the mocked DeleteVolume function
func (s *MockStorageService) DeleteVolume(ctx context.Context, id string) (*godo.Response, error) {
	return s.DeleteVolumeFunc(ctx, id)
}

// ListVolumes calls the mocked ListVolumes function
func (s *MockStorageService) ListVolumes(ctx context.Context, opt *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
	return s.ListVolumesFunc(ctx, opt)
}

// GetVolume calls the mocked GetVolume function
func (s *MockStorageService) GetVolume(ctx context.Context, id string) (*godo.Volume, *godo.Response, error) {
	return s.GetVolumeFunc(ctx, id)
}

// GetVolumeAction calls the mocked GetVolumeAction function
func (s *MockStorageService) GetVolumeAction(ctx context.Context, id string, actionID int) (*godo.Action, *godo.Response, error) {
	return s.GetVolumeActionFunc(ctx, id, actionID)
}

// AttachVolume calls the mocked AttachVolume function
func (s *MockStorageService) AttachVolume(ctx context.Context, id string, dropletID int) (*godo.Response, error) {
	return s.AttachVolumeFunc(ctx, id, dropletID)
}

// DetachVolume calls the mocked DetachVolume function
func (s *MockStorageService) DetachVolume(ctx context.Context, id string, dropletID int) (*godo.Response, error) {
	return s.DetachVolumeFunc(ctx, id, dropletID)
}

// SimulateNotFound configures the service to return not found errors
func (s *MockStorageService) SimulateNotFound() {
	s.ListVolumesFunc = func(_ context.Context, _ *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
		return nil, nil, s.std.Volumes.NotFoundError
	}
	s.GetVolumeFunc = func(_ context.Context, _ string) (*godo.Volume, *godo.Response, error) {
		return nil, nil, s.std.Volumes.NotFoundError
	}
	s.GetVolumeActionFunc = func(_ context.Context, _ string, _ int) (*godo.Action, *godo.Response, error) {
		return nil, nil, s.std.Volumes.NotFoundError
	}
	s.AttachVolumeFunc = func(_ context.Context, _ string, _ int) (*godo.Response, error) {
		return nil, s.std.Volumes.NotFoundError
	}
	s.DetachVolumeFunc = func(_ context.Context, _ string, _ int) (*godo.Response, error) {
		return nil, s.std.Volumes.NotFoundError
	}
}

// SimulateRateLimit configures the service to return rate limit errors
func (s *MockStorageService) SimulateRateLimit() {
	s.ListVolumesFunc = func(_ context.Context, _ *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
		return nil, nil, s.std.Volumes.RateLimitError
	}
	s.CreateVolumeFunc = func(_ context.Context, _ *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
		return nil, nil, s.std.Volumes.RateLimitError
	}
	s.DeleteVolumeFunc = func(_ context.Context, _ string) (*godo.Response, error) {
		return nil, s.std.Volumes.RateLimitError
	}
	s.GetVolumeFunc = func(_ context.Context, _ string) (*godo.Volume, *godo.Response, error) {
		return nil, nil, s.std.Volumes.RateLimitError
	}
	s.GetVolumeActionFunc = func(_ context.Context, _ string, _ int) (*godo.Action, *godo.Response, error) {
		return nil, nil, s.std.Volumes.RateLimitError
	}
	s.AttachVolumeFunc = func(_ context.Context, _ string, _ int) (*godo.Response, error) {
		return nil, s.std.Volumes.RateLimitError
	}
	s.DetachVolumeFunc = func(_ context.Context, _ string, _ int) (*godo.Response, error) {
		return nil, s.std.Volumes.RateLimitError
	}
}

// SimulateAuthenticationFailure configures the service to return authentication errors
func (s *MockStorageService) SimulateAuthenticationFailure() {
	s.ListVolumesFunc = func(_ context.Context, _ *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
		return nil, nil, s.std.Volumes.AuthenticationError
	}
	s.CreateVolumeFunc = func(_ context.Context, _ *godo.VolumeCreateRequest) (*godo.Volume, *godo.Response, error) {
		return nil, nil, s.std.Volumes.AuthenticationError
	}
	s.DeleteVolumeFunc = func(_ context.Context, _ string) (*godo.Response, error) {
		return nil, s.std.Volumes.AuthenticationError
	}
	s.GetVolumeFunc = func(_ context.Context, _ string) (*godo.Volume, *godo.Response, error) {
		return nil, nil, s.std.Volumes.AuthenticationError
	}
	s.GetVolumeActionFunc = func(_ context.Context, _ string, _ int) (*godo.Action, *godo.Response, error) {
		return nil, nil, s.std.Volumes.AuthenticationError
	}
	s.AttachVolumeFunc = func(_ context.Context, _ string, _ int) (*godo.Response, error) {
		return nil, s.std.Volumes.AuthenticationError
	}
	s.DetachVolumeFunc = func(_ context.Context, _ string, _ int) (*godo.Response, error) {
		return nil, s.std.Volumes.AuthenticationError
	}
}

// SimulateDelayedSuccess configures the service to succeed after a specific number of attempts
func (s *MockStorageService) SimulateDelayedSuccess(successAfterAttempts int) {
	s.attemptCount = 0

	// For waitForIP testing
	s.GetVolumeFunc = func(_ context.Context, _ string) (*godo.Volume, *godo.Response, error) {
		s.attemptCount++
		if s.attemptCount >= successAfterAttempts {
			return s.std.Volumes.DefaultVolume, nil, nil
		}
		// Return a volume with no ID before success
		volume := *s.std.Volumes.DefaultVolume
		volume.ID = ""
		return &volume, nil, nil
	}

	// For waitForDeletion testing
	s.ListVolumesFunc = func(_ context.Context, _ *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
		s.attemptCount++
		if s.attemptCount >= successAfterAttempts {
			// Return empty list to simulate deletion
			return []godo.Volume{}, nil, nil
		}
		// Return list with the volume still present
		return s.std.Volumes.DefaultVolumeList, nil, nil
	}
}

// SimulateMaxRetries configures the service to always fail until max retries are hit
func (s *MockStorageService) SimulateMaxRetries() {
	s.attemptCount = 0

	// For waitForIP testing
	s.GetVolumeFunc = func(_ context.Context, _ string) (*godo.Volume, *godo.Response, error) {
		s.attemptCount++
		volume := *s.std.Volumes.DefaultVolume
		volume.ID = ""
		return &volume, nil, nil
	}

	// For waitForDeletion testing
	s.ListVolumesFunc = func(_ context.Context, _ *godo.ListVolumeParams) ([]godo.Volume, *godo.Response, error) {
		s.attemptCount++
		// Always return list with the volume present
		return s.std.Volumes.DefaultVolumeList, nil, nil
	}
}

// GetAttemptCount returns the current attempt count
func (s *MockStorageService) GetAttemptCount() int {
	return s.attemptCount
}

// ResetAttemptCount resets the attempt counter
func (s *MockStorageService) ResetAttemptCount() {
	s.attemptCount = 0
}
