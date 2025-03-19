package test

import (
	"github.com/celestiaorg/talis/test/mocks"
)

func SetupMockDOClient(suite *TestSuite) {
	suite.MockDOClient = mocks.NewMockDOClient()
}
