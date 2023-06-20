package integration

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rscoordinatortypes "github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (s *IntegrationTestSuite) TestRegisterOperator() {
	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordApp.RestakingCoordinatorKeeper.SetConsumerAdditionProposal(coordCtx, proposal)

	s.SetupRestakingPath()

	consumerChainID := s.rsConsumerChain.ChainID
	registerAccAddr := s.path.EndpointA.Chain.SenderAccount.GetAddress()

	valSets := s.getConsumerValidators(consumerChainID)
	s.registerOperator(consumerChainID, proposal.RestakingTokens[0], valSets[0].Address, registerAccAddr)

	coordCtx = s.rsCoordinatorChain.GetContext()
	allOperator := coordApp.RestakingCoordinatorKeeper.GetAllOperators(coordCtx)

	s.Require().Equal(len(allOperator), 1)
	strings.EqualFold(allOperator[0].Owner, registerAccAddr.String())
}

func (s *IntegrationTestSuite) getConsumerValidators(chainID string) []rscoordinatortypes.ConsumerValidator {
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)

	clientID, found := coordApp.RestakingCoordinatorKeeper.GetConsumerClientID(coordCtx, chainID)
	s.Require().True(found)
	consumerValidators := coordApp.RestakingCoordinatorKeeper.GetConsumerValidators(coordCtx, string(clientID), 100)
	s.Require().True(found)

	return consumerValidators
}

func (s *IntegrationTestSuite) registerOperator(
	consumerChainID,
	restakingDenom string,
	consumerValidatorAddress string,
	register sdk.AccAddress,
) {
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)

	err := coordApp.RestakingCoordinatorKeeper.RegisterOperator(coordCtx, rscoordinatortypes.MsgRegisterOperator{
		ConsumerChainIDs:           []string{consumerChainID},
		ConsumerValidatorAddresses: []string{consumerValidatorAddress},
		RestakingDenom:             restakingDenom,
		Sender:                     register.String(),
	})

	s.Require().NoError(err)
}
