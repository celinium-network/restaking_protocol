package integration

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

func (s *IntegrationTestSuite) TestReward() {
	s.coordinator.Setup(s.transferPath)

	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	proposal.TransferChannelId = s.transferPath.EndpointA.ChannelID

	coordApp := getCoordinatorApp(s.rsCoordinatorChain)
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordKeeper := coordApp.RestakingCoordinatorKeeper
	coordKeeper.SetConsumerAdditionProposal(coordCtx, proposal)

	s.SetupRestakingPath()
	consumerChainID := s.rsConsumerChain.ChainID
	user := s.path.EndpointA.Chain.SenderAccount.GetAddress()

	valSets := s.getConsumerValidators(consumerChainID)
	s.registerOperator(consumerChainID, proposal.RestakingTokens[0], valSets[0].Address, user)

	coordCtx = s.rsCoordinatorChain.GetContext()
	operator := coordKeeper.GetAllOperators(coordCtx)[0]
	operatorAccAddr := sdk.MustAccAddressFromBech32(operator.OperatorAddress)
	amount := math.NewIntFromUint64(10000)

	mintCoins := sdk.Coins{sdk.NewCoin(proposal.RestakingTokens[0], amount)}
	coordApp.BankKeeper.MintCoins(coordCtx, ibctransfertypes.ModuleName, mintCoins)
	coordApp.BankKeeper.SendCoinsFromModuleToAccount(coordCtx, ibctransfertypes.ModuleName, user, mintCoins)

	err := coordKeeper.Delegate(coordCtx, user, operatorAccAddr, amount)
	s.Require().NoError(err)

	events := NextBlockWithEvents(s.rsCoordinatorChain)
	err = s.path.EndpointA.UpdateClient()
	s.Require().NoError(err)

	path := s.path
	path.EndpointA, path.EndpointB = path.EndpointB, path.EndpointA
	s.RelayIBCPacket(s.path, events, user.String())

	coordCtx = s.rsCoordinatorChain.GetContext()
	coordKeeper.WithdrawOperatorsReward(coordCtx)

	s.path.EndpointA.UpdateClient()
	s.path.EndpointB.UpdateClient()

	// set reward to the validator
	consumerApp := getConsumerApp(s.rsConsumerChain)
	consumerCtx := s.rsConsumerChain.GetContext()
	valAddr, err := sdk.ValAddressFromBech32(valSets[0].Address)
	s.Require().NoError(err)
	validator, found := consumerApp.StakingKeeper.GetValidator(consumerCtx, valAddr)
	s.Require().True(found)

	consumerApp.DistrKeeper.AllocateTokensToValidator(consumerCtx, validator, sdk.DecCoins{
		sdk.NewDecCoinFromDec("stake", sdk.MustNewDecFromStr("100000000")),
	})

	// s.RelayIBCPacket(s.path, coordCtx.EventManager().ABCIEvents(), user.String())
	events = coordCtx.EventManager().ABCIEvents()
	msgRecvPackets := parseMsgRecvPacketFromEvents(s.rsCoordinatorChain, events, user.String())

	consumerCtx = s.rsConsumerChain.GetContext()
	_, err = consumerApp.GetIBCKeeper().RecvPacket(consumerCtx, &msgRecvPackets[0])
	s.Require().NoError(err)

	s.rsConsumerChain.NextBlock()
	s.path.EndpointA.UpdateClient()

	ack, err := assembleAckPacketFromEvents(s.rsConsumerChain, msgRecvPackets[0].Packet, consumerCtx.EventManager().Events())
	s.Require().NoError(err)

	coordCtx = s.rsCoordinatorChain.GetContext()
	_, err = coordApp.GetIBCKeeper().Acknowledgement(coordCtx, ack)
	s.Require().NoError(err)

	s.transferPath.EndpointA.UpdateClient()
	s.transferPath.EndpointB.UpdateClient()

	recvPkg, err := assembleRecvPacketByEvents(s.rsConsumerChain, consumerCtx.EventManager().Events())
	s.Require().NoError(err)

	coordCtx = s.rsCoordinatorChain.GetContext()
	_, err = coordApp.GetIBCKeeper().RecvPacket(coordCtx, recvPkg)
	s.Require().NoError(err)

	coordCtx = s.rsCoordinatorChain.GetContext()
	historicalReward, found := coordApp.RestakingCoordinatorKeeper.GetOperatorHistoricalRewards(coordCtx, 0, operatorAccAddr)
	s.Require().True(found)
	s.Require().Equal(len(historicalReward.CumulativeRewardRatios), 1)
	s.Require().True(historicalReward.CumulativeRewardRatios[0].Amount.GT(math.LegacyZeroDec()))
}
