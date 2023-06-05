package integration

import (
	"fmt"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	rscoordinatortypes "github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (s *IntegrationTestSuite) TestChannelInit() {
	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	app := getCoordinatorApp(s.rsCoordinatorChain)
	ctx := s.rsCoordinatorChain.GetContext()
	app.RestakingCoordinatorKeeper.SetConsumerAdditionProposal(ctx, proposal)

	s.coordinator.Setup(s.path)

	ctx = s.rsCoordinatorChain.GetContext()
	_, found := app.RestakingCoordinatorKeeper.GetConsumerAdditionProposal(ctx, s.rsConsumerChain.ChainID)
	s.Require().False(found)
	_, found = app.RestakingCoordinatorKeeper.GetConsumerClientID(ctx, s.rsConsumerChain.ChainID)
	s.Require().True(found)
}

func CreateConsumerAdditionalProposal(path *ibctesting.Path, consumerChain *ibctesting.TestChain) *rscoordinatortypes.ConsumerAdditionProposal {
	tmClientCfg, ok := path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig)
	if !ok {
		panic("create consumer additional proposal failed ")
	}

	return &rscoordinatortypes.ConsumerAdditionProposal{
		Title:                 fmt.Sprintf("Add consumer chain %s proposal", consumerChain.ChainID),
		Description:           "test consumer addition proposal",
		ChainId:               consumerChain.ChainID,
		UnbondingPeriod:       tmClientCfg.UnbondingPeriod,
		TimeoutPeriod:         tmClientCfg.TrustingPeriod,
		TransferTimeoutPeriod: tmClientCfg.MaxClockDrift,
		AvailableCoinDenoms:   []string{"stake"},
		RewardCoinDenom:       []string{"stake"},
	}
}
