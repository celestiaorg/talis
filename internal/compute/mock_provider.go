package compute

import (
	"context"

	"github.com/celestiaorg/talis/internal/db/models"
)

// MockComputeProvider is a mock implementation of the ComputeProvider interface
// for testing purposes.
type MockComputeProvider struct {
}

var _ ComputeProvider = &MockComputeProvider{}

// NewMockComputeProvider creates a new instance of MockComputeProvider for testing purposes.
func NewMockComputeProvider() (*MockComputeProvider, error) {
	return &MockComputeProvider{}, nil
}

// ValidateCredentials mocks the validation of provider credentials
func (m *MockComputeProvider) ValidateCredentials() error {
	return nil
}

// GetEnvironmentVars mocks returning environment variables needed for the provider
func (m *MockComputeProvider) GetEnvironmentVars() map[string]string {
	return map[string]string{}
}

// ConfigureProvider mocks configuring the provider with the given stack
func (m *MockComputeProvider) ConfigureProvider(stack interface{}) error {
	return nil
}

// CreateInstance mocks creating a new instance
func (m *MockComputeProvider) CreateInstance(ctx context.Context, name string, config InstanceConfig) ([]InstanceInfo, error) {
	return []InstanceInfo{{
		ID:       "mock-instance-id",
		Name:     name,
		PublicIP: "192.168.1.100",
		Provider: models.ProviderID("mock"),
		Region:   config.Region,
		Size:     config.Size,
	}}, nil
}

// DeleteInstance mocks deleting an instance
func (m *MockComputeProvider) DeleteInstance(ctx context.Context, name string, region string) error {
	return nil
}
