package integration

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	rscoordinatortypes "github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (s *IntegrationTestSuite) TestDelegation() {
	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordKeeper := coordApp.RestakingCoordinatorKeeper
	coordKeeper.SetConsumerAdditionProposal(coordCtx, proposal)

	s.SetupRestakingPath()

	consumerChainID := s.rsConsumerChain.ChainID
	user := s.path.EndpointA.Chain.SenderAccount.GetAddress()

	valSets := s.getConsumerValidators(consumerChainID)
	s.registerOperator(consumerChainID, proposal.RestakingTokens[0], valSets[0].PubKey, user)

	coordCtx = s.rsCoordinatorChain.GetContext()
	operator := coordKeeper.GetAllOperators(coordCtx)[0]
	operatorAccAddr := sdk.MustAccAddressFromBech32(operator.OperatorAddress)
	amount := math.NewIntFromUint64(10000)

	mintCoins := sdk.Coins{sdk.NewCoin(proposal.RestakingTokens[0], amount)}
	coordApp.BankKeeper.MintCoins(coordCtx, ibctransfertypes.ModuleName, mintCoins)
	coordApp.BankKeeper.SendCoinsFromModuleToAccount(coordCtx, ibctransfertypes.ModuleName, user, mintCoins)

	err := coordKeeper.Delegate(coordCtx, user, operatorAccAddr, amount)
	s.Require().NoError(err)

	operatorDelegationRecord, found := coordKeeper.GetOperatorDelegateRecord(coordCtx, uint64(coordCtx.BlockHeight()), operator.OperatorAddress)
	s.Require().True(found)
	s.Require().Equal(operatorDelegationRecord.Status, rscoordinatortypes.InterChainDelegateCall)
	s.Require().True(operatorDelegationRecord.DelegationAmount.Equal(amount))

	events := NextBlockWithEvents(s.rsCoordinatorChain)
	s.path.EndpointA.UpdateClient()

	path := s.path
	path.EndpointA, path.EndpointB = path.EndpointB, path.EndpointA
	s.RelayIBCPacket(s.path, events, user.String())

	consumerApp := getConsumerApp(s.rsConsumerChain)
	consumerCtx := s.rsConsumerChain.GetContext()
	consumerKeeper := consumerApp.RestakingConsumerKeeper

	validatorPkBz := consumerApp.AppCodec().MustMarshal(&operator.OperatedValidators[0].ValidatorPk)
	localOperatorAccAddr, found := consumerKeeper.GetOperatorLocalAddress(consumerCtx, operator.OperatorAddress, validatorPkBz)
	s.Require().True(found)

	agents := consumerApp.MultiStakingKeeper.GetAllAgent(consumerCtx)

	agentAccAddr := sdk.MustAccAddressFromBech32(agents[0].DelegateAddress)

	delegations := consumerApp.StakingKeeper.GetDelegatorDelegations(consumerCtx, agentAccAddr, 10)
	s.Require().Equal(len(delegations), 1)
	shares := consumerApp.MultiStakingKeeper.GetMultiStakingShares(consumerCtx, agents[0].Id, localOperatorAccAddr.String())
	s.Require().True(shares.Equal(amount))
}
