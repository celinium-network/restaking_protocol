package keeper_test

import (
	abci "github.com/cometbft/cometbft/abci/types"
	cryptocodec "github.com/cometbft/cometbft/crypto/encoding"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/golang/mock/gomock"

	sdkmock "github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

func (s *KeeperTestSuite) setupConsumerChain(
	ctx sdk.Context,
	chainID string,
	clientID string,
	validators []tmprotocrypto.PublicKey,
	restakingTokens []string,
	rewardToken []string,
) {
	s.coordinatorKeeper.SetConsumerClientID(ctx, chainID, clientID)
	s.coordinatorKeeper.SetConsumerRestakingToken(ctx, clientID, restakingTokens)
	s.coordinatorKeeper.SetConsumerRewardToken(ctx, clientID, rewardToken)

	validatorUpdates := abci.ValidatorUpdates{}
	for _, pk := range validators {
		validatorUpdates = append(validatorUpdates, abci.ValidatorUpdate{
			PubKey: pk,
			Power:  1,
		})
	}

	s.coordinatorKeeper.SetConsumerValidator(ctx, clientID, validatorUpdates)
}

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

func (s *KeeperTestSuite) TestDelegate() {
	consumerChainIDs := []string{"consumer-0", "consumer-1", "consumer-2"}
	consumerClientIDs := []string{"client-0", "client-1", "client-2"}

	var validatorPks []tmprotocrypto.PublicKey
	for i := 0; i < 3; i++ {
		pk, err := mockTmProtoPublicKey()
		s.Require().NoError(err)
		validatorPks = append(validatorPks, pk)
	}

	for i, chainID := range consumerChainIDs {
		s.setupConsumerChain(
			s.ctx,
			chainID,
			consumerClientIDs[i],
			validatorPks,
			[]string{"stake"},
			[]string{"stake"},
		)
	}

	accounts := simtestutil.CreateIncrementalAccounts(1)
	user := accounts[0]
	err := s.coordinatorKeeper.RegisterOperator(s.ctx, types.MsgRegisterOperator{
		ConsumerChainIDs:     consumerChainIDs,
		ConsumerValidatorPks: validatorPks,
		RestakingDenom:       "stake",
		Sender:               user.String(),
	})
	s.Require().NoError(err)

	createdOperator := s.coordinatorKeeper.GetAllOperators(s.ctx)[0]
	createdOperatorAccAddr := sdk.MustAccAddressFromBech32(createdOperator.OperatorAddress)

	delegateAmt := sdk.NewIntFromUint64(100000)
	delegateCoins := sdk.Coins{sdk.NewCoin("stake", delegateAmt)}

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), user, types.ModuleName, delegateCoins)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, createdOperatorAccAddr, delegateCoins)

	err = s.coordinatorKeeper.Delegate(s.ctx, user, createdOperatorAccAddr, delegateAmt)
	s.Require().NoError(err)

	delegation, found := s.coordinatorKeeper.GetDelegation(s.ctx, user.String(), createdOperator.OperatorAddress)
	s.Require().True(found)
	s.Require().Equal(delegation.Delegator, user.String())
	s.Require().Equal(delegation.Operator, createdOperator.OperatorAddress)
	s.Require().True(delegation.Shares.Equal(delegateAmt))

	opDelegationRecord, found := s.coordinatorKeeper.GetOperatorDelegateRecord(s.ctx, uint64(s.ctx.BlockHeight()), createdOperator.OperatorAddress)
	s.Require().True(found)
	s.Require().True(opDelegationRecord.DelegationAmount.Equal(delegateAmt))
	s.Require().Equal(opDelegationRecord.Status, types.OpDelRecordPending)
	s.Require().Equal(len(opDelegationRecord.IbcCallbackIds), 0)
}
