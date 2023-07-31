package integration

import (
	"fmt"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	coordapp "github.com/celinium-network/restaking_protocol/app/coordinator"
	rscoordinatortypes "github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (s *IntegrationTestSuite) TestChannelInit() {
	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	app := getCoordinatorApp(s.rsCoordinatorChain)
	ctx := s.rsCoordinatorChain.GetContext()
	app.RestakingCoordinatorKeeper.SetConsumerAdditionProposal(ctx, proposal)

	s.SetupRestakingPath()

	ctx = s.rsCoordinatorChain.GetContext()
	_, found := app.RestakingCoordinatorKeeper.GetConsumerAdditionProposal(ctx, s.rsConsumerChain.ChainID)
	s.Require().False(found)
	_, found = app.RestakingCoordinatorKeeper.GetConsumerClientID(ctx, s.rsConsumerChain.ChainID)
	s.Require().True(found)
}

func (s *IntegrationTestSuite) SetupRestakingPath() {
	s.coordinator.SetupConnections(s.path)

	err := s.path.EndpointA.ChanOpenInit()
	s.Require().NoError(err)

	err = s.path.EndpointB.ChanOpenTry()
	s.Require().NoError(err)

	events := s.ChanOpenAck(s.path.EndpointA)

	// Consumer send validatorSet information to coordinator.
	// So we must get ibc packet from events which emit in `EndBlock` and relay it to coordinator
	err = s.path.EndpointB.ChanOpenConfirm()
	s.Require().NoError(err)

	consumerUserAddr := s.path.EndpointA.Chain.SenderAccount.GetAddress()
	s.RelayIBCPacket(s.path, events, consumerUserAddr.String())
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
		RestakingTokens:       []string{coordapp.DefaultBondDenom},
		RewardTokens:          []string{coordapp.DefaultBondDenom},
	}
}
