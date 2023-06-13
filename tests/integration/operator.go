package integration

import (
	"strings"

	"github.com/cometbft/cometbft/proto/tendermint/crypto"

	rscoordinatortypes "github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (s *IntegrationTestSuite) TestRegisterOperator() {
	consumerChainID := s.rsConsumerChain.ChainID
	proposal := CreateConsumerAdditionalProposal(s.path, s.rsConsumerChain)
	coordApp := getCoordinatorApp(s.rsCoordinatorChain)
	coordCtx := s.rsCoordinatorChain.GetContext()
	coordApp.RestakingCoordinatorKeeper.SetConsumerAdditionProposal(coordCtx, proposal)

	s.SetupRestakingPath()

	coordCtx = s.rsCoordinatorChain.GetContext()
	clientID, found := coordApp.RestakingCoordinatorKeeper.GetConsumerClientID(coordCtx, consumerChainID)
	s.Require().True(found)
	valUpdates, found := coordApp.RestakingCoordinatorKeeper.GetConsumerValidator(coordCtx, string(clientID))
	s.Require().True(found)

	consumerUserAddr := s.path.EndpointA.Chain.SenderAccount.GetAddress()

	coordApp.RestakingCoordinatorKeeper.RegisterOperator(coordCtx, rscoordinatortypes.MsgRegisterOperator{
		ConsumerChainIDs:     []string{consumerChainID},
		ConsumerValidatorPks: []crypto.PublicKey{valUpdates[0].PubKey},
		RestakingDenom:       proposal.RestakingTokens[0],
		Sender:               consumerUserAddr.String(),
	})

	allOperator := coordApp.RestakingCoordinatorKeeper.GetAllOperators(coordCtx)

	s.Require().Equal(len(allOperator), 1)
	strings.EqualFold(allOperator[0].Owner, consumerUserAddr.String())
}
