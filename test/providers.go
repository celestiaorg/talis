package test

import (
	"github.com/celestiaorg/talis/test/mocks"
)

// SetupMockDOClient sets up a mock DigitalOcean client
func SetupMockDOClient(suite *Suite) {
	suite.MockDOClient = mocks.NewMockDOClient()
}
