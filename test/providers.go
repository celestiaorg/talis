package test

import (
	"github.com/celestiaorg/talis/internal/compute"
	"github.com/celestiaorg/talis/internal/types"
)

// SetupMockDOClient sets up a mock DigitalOcean client
func SetupMockDOClient(suite *TestSuite) {
	suite.MockDOClient = compute.NewMockDOClient()
}

func NewTestProvider() types.ComputeProvider {
	return compute.NewMockDOClient()
}

func NewTestDOClient() *compute.MockDOClient {
	return compute.NewMockDOClient()
}
