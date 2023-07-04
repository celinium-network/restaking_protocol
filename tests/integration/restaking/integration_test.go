package integration_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	restaking "github.com/celinium-network/restaking_protocol/tests/integration/restaking"
)

func TestRestakingSuite(t *testing.T) {
	restakingSuite := restaking.NewIntegrationTestSuite()

	suite.Run(t, restakingSuite)
}
