package test

import (
	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/types"
	"github.com/celestiaorg/talis/test/mocks"
)

// SetupMockDOClient sets up a mock DigitalOcean client
func SetupMockDOClient(suite *Suite) {
	suite.MockDOClient = mocks.NewMockDOClient()
}

func NewTestProvider() types.ComputeProvider {
	return compute.NewMockDOClient()
}

func NewTestDOClient() *compute.MockDOClient {
	return compute.NewMockDOClient()
}
