package test

import (
	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/test/mocks"
)

// SetupMockDOClient sets up a mock DigitalOcean client
func SetupMockDOClient(suite *Suite) {
	suite.MockDOClient = mocks.NewMockDOClient()
}

// NewTestProvider creates a new test provider instance
func NewTestProvider() compute.Provider {
	return mocks.NewMockDOClient()
}

// NewTestDOClient creates a new test DigitalOcean client instance
func NewTestDOClient() *mocks.MockDOClient {
	return mocks.NewMockDOClient()
}
