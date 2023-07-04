package keeper_test

import (
	"time"

	"github.com/golang/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	mtstakingtypes "github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

func (s *KeeperTestSuite) TestSlashAgentFromValidator() {
	valAddr := sdk.ValAddress(pks[0].Address())
	validator := stakingtypes.Validator{
		OperatorAddress: valAddr.String(),
	}
	delegatorAccAddr := accounts[0]
	agentAccAddr := s.mtStakingKeeper.GenerateAccount(s.ctx, mtStakingDenom, validator.OperatorAddress).GetAddress()
	toDefaultDenomMultiplier := sdk.MustNewDecFromStr("1")
	delegateAmount := mustNewIntForStr("1000000000")

	delegateCoin := sdk.NewCoin(mtStakingDenom, delegateAmount)
	eqCoin := sdk.NewCoin(defaultBondDenom, toDefaultDenomMultiplier.MulInt(delegateAmount).TruncateInt())

	s.mtStakingKeeper.AddMTStakingDenom(s.ctx, mtStakingDenom)
	s.mtStakingKeeper.SetEquivalentNativeCoinMultiplier(s.ctx, 1, mtStakingDenom, toDefaultDenomMultiplier)

	s.delegateExpectOtherKeeperAction(delegateCoin, validator, delegatorAccAddr, eqCoin, agentAccAddr)
	err := s.mtStakingKeeper.MTStakingDelegate(s.ctx, mtstakingtypes.MsgMTStakingDelegate{
		DelegatorAddress: delegatorAccAddr.String(),
		ValidatorAddress: validator.OperatorAddress,
		Balance:          delegateCoin,
	})
	s.Require().NoError(err)

	slashFactor := sdk.MustNewDecFromStr("0.05")
	slashAmount := slashFactor.MulInt(delegateAmount).TruncateInt()
	slashCoin := sdk.NewCoin(mtStakingDenom, slashAmount)

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), agentAccAddr, mtstakingtypes.ModuleName, sdk.NewCoins(slashCoin)).Return(nil)
	s.bankKeeper.EXPECT().BurnCoins(s.ctx, mtstakingtypes.ModuleName, sdk.NewCoins(slashCoin)).Return(nil)
	s.mtStakingKeeper.SlashAgentFromValidator(s.ctx, valAddr, slashFactor)

	agent, found := s.mtStakingKeeper.GetMTStakingAgentByAddress(s.ctx, agentAccAddr)
	s.Require().True(found)
	s.Require().True(agent.StakedAmount.Equal(delegateAmount.Sub(slashAmount)))
}

func (s *KeeperTestSuite) TestSlashAgentFromValidatorWithUnbonding() {
	valAddr := sdk.ValAddress(pks[0].Address())
	validator := stakingtypes.Validator{
		OperatorAddress: valAddr.String(),
	}
	delegatorAccAddr := accounts[0]
	agentAccAddr := s.mtStakingKeeper.GenerateAccount(s.ctx, mtStakingDenom, validator.OperatorAddress).GetAddress()
	toDefaultDenomMultiplier := sdk.MustNewDecFromStr("1")
	delegateAmount := mustNewIntForStr("1000000000")

	delegateCoin := sdk.NewCoin(mtStakingDenom, delegateAmount)
	eqCoin := sdk.NewCoin(defaultBondDenom, toDefaultDenomMultiplier.MulInt(delegateAmount).TruncateInt())

	s.mtStakingKeeper.AddMTStakingDenom(s.ctx, mtStakingDenom)
	s.mtStakingKeeper.SetEquivalentNativeCoinMultiplier(s.ctx, 1, mtStakingDenom, toDefaultDenomMultiplier)

	s.delegateExpectOtherKeeperAction(delegateCoin, validator, delegatorAccAddr, eqCoin, agentAccAddr)
	err := s.mtStakingKeeper.MTStakingDelegate(s.ctx, mtstakingtypes.MsgMTStakingDelegate{
		DelegatorAddress: delegatorAccAddr.String(),
		ValidatorAddress: validator.OperatorAddress,
		Balance:          delegateCoin,
	})
	s.Require().NoError(err)

	// send some unbondingDelegationEntry
	unbondingAmount := mustNewIntForStr("1000000000")
	unbonding := s.mtStakingKeeper.GetOrCreateMTStakingUnbonding(s.ctx, agentAccAddr, delegatorAccAddr)
	unbondingTime := time.Hour * 72

	undelegateCompleteTime := s.ctx.BlockTime().Add(unbondingTime)
	unbonding.Entries = append(unbonding.Entries, mtstakingtypes.MTStakingUnbondingDelegationEntry{
		CreatedHeight:  s.ctx.BlockHeight(),
		CompletionTime: undelegateCompleteTime,
		InitialBalance: sdk.NewCoin(mtStakingDenom, unbondingAmount),
		Balance:        sdk.NewCoin(mtStakingDenom, unbondingAmount),
	})

	s.mtStakingKeeper.SetMTStakingUnbondingDelegation(s.ctx, unbonding)
	s.mtStakingKeeper.InsertUBDQueue(s.ctx, unbonding, undelegateCompleteTime)

	slashFactor := sdk.MustNewDecFromStr("0.05")
	slashAmount := slashFactor.MulInt(delegateAmount).TruncateInt()
	slashCoin := sdk.NewCoin(mtStakingDenom, slashAmount)

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), agentAccAddr, mtstakingtypes.ModuleName, sdk.NewCoins(slashCoin)).Return(nil)
	s.bankKeeper.EXPECT().BurnCoins(s.ctx, mtstakingtypes.ModuleName, sdk.NewCoins(slashCoin)).Return(nil)
	s.mtStakingKeeper.SlashAgentFromValidator(s.ctx, valAddr, slashFactor)

	agent, found := s.mtStakingKeeper.GetMTStakingAgentByAddress(s.ctx, agentAccAddr)
	s.Require().True(found)
	s.Require().True(agent.StakedAmount.Equal(delegateAmount.Sub(slashAmount)))
	unbondingDelegation := s.mtStakingKeeper.GetOrCreateMTStakingUnbonding(s.ctx, agentAccAddr, delegatorAccAddr)
	s.Require().NotNil(unbondingDelegation)
	s.Require().Equal(len(unbondingDelegation.Entries), 1)
	s.Require().True(unbondingDelegation.Entries[0].Balance.Amount.
		Equal(unbondingAmount.Sub(slashFactor.MulInt(unbondingAmount).TruncateInt())))
}
