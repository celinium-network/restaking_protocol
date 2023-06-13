package integration

import (
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"

	rscoordinatortypes "github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (s *IntegrationTestSuite) xTestRegisterOperator() {
	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordApp.RestakingCoordinatorKeeper.SetConsumerAdditionProposal(coordCtx, proposal)

	s.SetupRestakingPath()

	consumerChainID := s.rsConsumerChain.ChainID
	registerAccAddr := s.path.EndpointA.Chain.SenderAccount.GetAddress()

	valSets := s.getConsumerValidators(consumerChainID)
	s.registerOperator(consumerChainID, proposal.RestakingTokens[0], valSets[0].PubKey, registerAccAddr)

	coordCtx = s.rsCoordinatorChain.GetContext()
	allOperator := coordApp.RestakingCoordinatorKeeper.GetAllOperators(coordCtx)

	s.Require().Equal(len(allOperator), 1)
	strings.EqualFold(allOperator[0].Owner, registerAccAddr.String())
}

func (s *IntegrationTestSuite) getConsumerValidators(chainID string) []abci.ValidatorUpdate {
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)

	clientID, found := coordApp.RestakingCoordinatorKeeper.GetConsumerClientID(coordCtx, chainID)
	s.Require().True(found)
	valUpdates, found := coordApp.RestakingCoordinatorKeeper.GetConsumerValidator(coordCtx, string(clientID))
	s.Require().True(found)

	return valUpdates
}

func (s *IntegrationTestSuite) registerOperator(
	consumerChainID,
	restakingDenom string,
	consumerValidatorPk crypto.PublicKey,
	register sdk.AccAddress,
) {
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)

	err := coordApp.RestakingCoordinatorKeeper.RegisterOperator(coordCtx, rscoordinatortypes.MsgRegisterOperator{
		ConsumerChainIDs:     []string{consumerChainID},
		ConsumerValidatorPks: []crypto.PublicKey{consumerValidatorPk},
		RestakingDenom:       restakingDenom,
		Sender:               register.String(),
	})

	s.Require().NoError(err)
}
