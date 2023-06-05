package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/celinium-network/restaking_protocol/tests/integration"
)

func TestRestakingSuite(t *testing.T) {
	restakingSuite := integration.NewIntegrationTestSuite()

	suite.Run(t, restakingSuite)
}
