package keeper_test

import (
	abci "github.com/cometbft/cometbft/abci/types"
	cryptocodec "github.com/cometbft/cometbft/crypto/encoding"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	sdkmock "github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func mockTmProtoPublicKey() (tmprotocrypto.PublicKey, error) {
	pv := sdkmock.NewPV()
	cpv, err := pv.GetPubKey()
	if err != nil {
		return tmprotocrypto.PublicKey{}, err
	}

	return cryptocodec.PubKeyToProto(cpv)
}

// func (s *KeeperTestSuite) setupConsumerChain(
// 	ctx sdk.Context,
// 	chainID string,
// 	clientID string,
// 	validators []tmprotocrypto.PublicKey,
// 	restakingTokens []string,
// 	rewardToken []string,
// ) {
// 	s.coordinatorKeeper.SetConsumerClientID(ctx, chainID, clientID)
// 	s.coordinatorKeeper.SetConsumerRestakingToken(ctx, clientID, restakingTokens)
// 	s.coordinatorKeeper.SetConsumerRewardToken(ctx, clientID, rewardToken)

// 	validatorUpdates := abci.ValidatorUpdates{}
// 	for _, pk := range validators {
// 		validatorUpdates = append(validatorUpdates, abci.ValidatorUpdate{
// 			PubKey: pk,
// 			Power:  1,
// 		})
// 	}

// 	s.coordinatorKeeper.SetConsumerValidator(ctx, clientID, validatorUpdates)
// }

func (s *KeeperTestSuite) TestRegisterOperator() {
	ctx, keeper := s.ctx, s.coordinatorKeeper
	consumerChainIDs := []string{"consumer-0", "consumer-1", "consumer-2"}
	consumerClientIDs := []string{"client-0", "client-1", "client-2"}

	addr := simtestutil.CreateIncrementalAccounts(1)

	var tmPubkeys []tmprotocrypto.PublicKey
	for i := 0; i < len(consumerChainIDs); i++ {
		keeper.SetConsumerClientID(ctx, consumerChainIDs[i], consumerClientIDs[i])

		tmProtoPk, err := mockTmProtoPublicKey()
		s.Require().NoError(err)
		tmPubkeys = append(tmPubkeys, tmProtoPk)

		keeper.SetConsumerValidator(ctx, consumerClientIDs[i], []abci.ValidatorUpdate{{
			PubKey: tmProtoPk,
			Power:  1,
		}})

		keeper.SetConsumerRestakingToken(ctx, consumerClientIDs[i], []string{"stake"})
		keeper.SetConsumerRewardToken(ctx, consumerClientIDs[i], []string{"stake"})
	}

	err := keeper.RegisterOperator(ctx, types.MsgRegisterOperator{
		ConsumerChainIDs:     consumerChainIDs,
		ConsumerValidatorPks: tmPubkeys,
		RestakingDenom:       "stake",
		Sender:               addr[0].String(),
	})

	s.Require().NoError(err)
}
