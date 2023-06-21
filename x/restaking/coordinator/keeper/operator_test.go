package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (s *KeeperTestSuite) setupConsumerChain(
	ctx sdk.Context,
	chainID string,
	clientID string,
	validatorAddresses []string,
	restakingTokens []string,
	rewardToken []string,
) {
	s.coordinatorKeeper.SetConsumerClientID(ctx, chainID, clientID)
	s.coordinatorKeeper.SetConsumerRestakingToken(ctx, clientID, restakingTokens)
	s.coordinatorKeeper.SetConsumerRewardToken(ctx, clientID, rewardToken)

	for _, valAddr := range validatorAddresses {
		s.coordinatorKeeper.SetConsumerValidator(ctx, clientID, types.ConsumerValidator{
			Address: valAddr,
		})
	}
}

func (s *KeeperTestSuite) TestRegisterOperator() {
	ctx, keeper := s.ctx, s.coordinatorKeeper
	consumerChainIDs := []string{"consumer-0", "consumer-1", "consumer-2"}
	consumerClientIDs := []string{"client-0", "client-1", "client-2"}

	addr := simtestutil.CreateIncrementalAccounts(1)

	var validatorAddress []string
	for i := 0; i < len(consumerChainIDs); i++ {
		keeper.SetConsumerClientID(ctx, consumerChainIDs[i], consumerClientIDs[i])

		valAddr := sdk.ValAddress(PKs[i].Address().Bytes()).String()
		validatorAddress = append(validatorAddress, valAddr)

		keeper.SetConsumerValidator(ctx, consumerClientIDs[i], types.ConsumerValidator{
			Address: valAddr,
		})

		keeper.SetConsumerRestakingToken(ctx, consumerClientIDs[i], []string{"stake"})
		keeper.SetConsumerRewardToken(ctx, consumerClientIDs[i], []string{"stake"})
	}

	err := keeper.RegisterOperator(ctx, types.MsgRegisterOperator{
		ConsumerChainIDs:           consumerChainIDs,
		ConsumerValidatorAddresses: validatorAddress,
		RestakingDenom:             "stake",
		Sender:                     addr[0].String(),
	})

	s.Require().NoError(err)
}

func (s *KeeperTestSuite) TestDelegate() {
	consumerChainIDs := []string{"consumer-0", "consumer-1", "consumer-2"}
	consumerClientIDs := []string{"client-0", "client-1", "client-2"}

	var consumerValidatorAddresses []string
	for i := 0; i < 3; i++ {
		consumerValidatorAddresses = append(consumerValidatorAddresses, sdk.ValAddress(PKs[i].Address().Bytes()).String())
	}

	for i, chainID := range consumerChainIDs {
		s.setupConsumerChain(
			s.ctx,
			chainID,
			consumerClientIDs[i],
			consumerValidatorAddresses,
			[]string{"stake"},
			[]string{"stake"},
		)
	}

	accounts := simtestutil.CreateIncrementalAccounts(1)
	user := accounts[0]
	err := s.coordinatorKeeper.RegisterOperator(s.ctx, types.MsgRegisterOperator{
		ConsumerChainIDs:           consumerChainIDs,
		ConsumerValidatorAddresses: consumerValidatorAddresses,
		RestakingDenom:             "stake",
		Sender:                     user.String(),
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

func (s *KeeperTestSuite) TestUndelegate() {
	operator := s.mockOperator()
	ctx, keeper := s.ctx, s.coordinatorKeeper

	accounts := simtestutil.CreateIncrementalAccounts(1)
	delegatorAccAddr := accounts[0]
	delegatorAddress := delegatorAccAddr.String()

	// TODO maybe become a mock function like: mockDelegation
	delegateAmt := math.NewIntFromUint64(100000)
	keeper.SetDelegation(ctx, delegatorAddress, operator.OperatorAddress, &types.Delegation{
		Delegator: delegatorAddress,
		Operator:  operator.OperatorAddress,
		Shares:    delegateAmt,
	})
	operator.Shares = delegateAmt
	operator.RestakedAmount = delegateAmt
	keeper.SetOperator(ctx, operator)

	operatorAccAddr := sdk.MustAccAddressFromBech32(operator.OperatorAddress)

	err := keeper.Undelegate(ctx, delegatorAccAddr, operatorAccAddr, delegateAmt)
	s.Require().NoError(err)

	// check UnbondDelegationRecord
	record, found := keeper.GetOperatorUndelegationRecord(ctx, uint64(ctx.BlockHeight()), operator.OperatorAddress)
	s.Require().True(found)
	s.Require().Equal(record.Status, types.OpUndelegationRecordPending)
	s.Require().Equal(len(record.UnbondingEntryIds), 1)
	s.Require().True(record.UndelegationAmount.Equal(delegateAmt))

	unbonding, found := keeper.GetUnbondingDelegation(ctx, delegatorAccAddr, operatorAccAddr)
	s.Require().True(found)
	s.Require().Equal(len(unbonding.Entries), 1)

	unbondingEntry := unbonding.Entries[0]
	s.Require().Equal(unbondingEntry.Id, record.UnbondingEntryIds[0])
	s.True(unbondingEntry.Amount.Amount.Equal(delegateAmt))
}
