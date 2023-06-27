package integration

import (
	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var PKs = simtestutil.CreateTestPubKeys(500)

func (s *IntegrationTestSuite) consumerAddValidator(valPubKey cryptotypes.PubKey) sdk.ValAddress {
	ctx := s.rsConsumerChain.GetContext()
	consumerApp := getConsumerApp(s.rsConsumerChain)

	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := consumerApp.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	newValidator, err := stakingtypes.NewValidator(valAddr, valPubKey, stakingtypes.Description{})
	s.Require().NoError(err)
	newValidator.AddTokensFromDel(valTokens)

	consumerApp.StakingKeeper.SetValidator(ctx, newValidator)
	consumerApp.StakingKeeper.SetValidatorByPowerIndex(ctx, newValidator)
	consumerApp.StakingKeeper.SetValidatorByConsAddr(ctx, newValidator)
	return valAddr
}

func (s *IntegrationTestSuite) TestConsumerAddValidator() {
	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordKeeper := coordApp.RestakingCoordinatorKeeper
	relayerAddr := s.path.EndpointA.Chain.SenderAccount.GetAddress()

	coordKeeper.SetConsumerAdditionProposal(coordCtx, proposal)

	s.SetupRestakingPath()

	consumerCtx := s.rsConsumerChain.GetContext()
	consumerApp := getConsumerApp(s.rsConsumerChain)
	valPubKey := PKs[0]
	s.consumerAddValidator(valPubKey)
	addingValAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	consumerApp.StakingKeeper.Hooks().AfterValidatorCreated(consumerCtx, addingValAddr)

	pendingVCSList := consumerApp.RestakingConsumerKeeper.GetPendingVSCList(consumerCtx)
	s.Require().Equal(len(pendingVCSList), 1)
	s.Require().Equal(len(pendingVCSList[0].ValidatorAddresses), 1)
	s.Require().Equal(pendingVCSList[0].ValidatorAddresses[0], addingValAddr.String())

	events := NextBlockWithEvents(s.rsConsumerChain)

	s.path.EndpointA.UpdateClient()
	s.path.EndpointB.UpdateClient()

	consumerTmClientID, found := coordApp.RestakingCoordinatorKeeper.GetConsumerClientID(coordCtx, s.rsConsumerChain.ChainID)
	s.Require().True(found)
	valSetBefore := coordKeeper.GetConsumerValidators(coordCtx, string(consumerTmClientID), 100)
	valSetBeforeLen := len(valSetBefore)
	s.RelayIBCPacket(s.path, events, relayerAddr.String())

	coordCtx = s.rsCoordinatorChain.GetContext()
	valSetAfter := coordKeeper.GetConsumerValidators(coordCtx, string(consumerTmClientID), 100)
	valSetAfterLen := len(valSetAfter)
	s.Require().Equal(valSetAfterLen-valSetBeforeLen, 1)
	added := false
	for _, v := range valSetAfter {
		if v.Address == addingValAddr.String() {
			added = true
			break
		}
	}
	s.Require().True(added)
}

func (s *IntegrationTestSuite) TestConsumerRemoveValidator() {
	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordKeeper := coordApp.RestakingCoordinatorKeeper
	relayerAddr := s.path.EndpointA.Chain.SenderAccount.GetAddress()

	coordKeeper.SetConsumerAdditionProposal(coordCtx, proposal)

	s.SetupRestakingPath()

	consumerCtx := s.rsConsumerChain.GetContext()
	consumerApp := getConsumerApp(s.rsConsumerChain)
	valPubKey := PKs[0]
	valAddr := s.consumerAddValidator(valPubKey)
	addingValAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	consumerApp.StakingKeeper.Hooks().AfterValidatorCreated(consumerCtx, addingValAddr)
	events := NextBlockWithEvents(s.rsConsumerChain)

	s.path.EndpointA.UpdateClient()
	s.path.EndpointB.UpdateClient()

	s.RelayIBCPacket(s.path, events, relayerAddr.String())

	consumerCtx = s.rsConsumerChain.GetContext()
	validator, found := consumerApp.StakingKeeper.GetValidator(consumerCtx, valAddr)
	s.Require().True(found)
	validator.DelegatorShares = math.LegacyZeroDec()
	consumerApp.StakingKeeper.SetValidator(consumerCtx, validator)
	consumerApp.StakingKeeper.RemoveValidator(consumerCtx, valAddr)

	events = NextBlockWithEvents(s.rsConsumerChain)

	s.path.EndpointA.UpdateClient()
	s.path.EndpointB.UpdateClient()

	s.RelayIBCPacket(s.path, events, relayerAddr.String())

	coordCtx = s.rsCoordinatorChain.GetContext()
	consumerTmClientID, found := coordApp.RestakingCoordinatorKeeper.GetConsumerClientID(coordCtx, s.rsConsumerChain.ChainID)
	s.Require().True(found)
	valSetAfter := coordKeeper.GetConsumerValidators(coordCtx, string(consumerTmClientID), 100)
	removed := true
	for _, v := range valSetAfter {
		if v.Address == valAddr.String() {
			removed = false
		}
	}
	s.Require().True(removed)
}
