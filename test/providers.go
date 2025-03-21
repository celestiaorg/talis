package test

import (
	"github.com/celestiaorg/talis/test/mocks"
)

// SetupMockDOClient sets up a mock DigitalOcean client
func SetupMockDOClient(suite *TestSuite) {
	suite.MockDOClient = mocks.NewMockDOClient()
}
