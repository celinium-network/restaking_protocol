package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	app "github.com/celinium-network/restaking_protocol/app/consumer"
	"github.com/celinium-network/restaking_protocol/x/epochs/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.App
	queryClient types.QueryClient
	consAddress sdk.ConsAddress
}

var _ *KeeperTestSuite

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest()
}
