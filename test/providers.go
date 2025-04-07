package test

import (
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/test/mocks"
)

// SetupMockDOClient sets up a mock DigitalOcean client
func SetupMockDOClient(suite *Suite) {
	suite.MockDOClient = mocks.NewMockDOClient()
}

// NewTestProvider creates a new test provider instance
func NewTestProvider() types.ComputeProvider {
	return mocks.NewMockDOClient()
}

// NewTestDOClient creates a new test DigitalOcean client instance
func NewTestDOClient() *mocks.MockDOClient {
	return mocks.NewMockDOClient()
}
