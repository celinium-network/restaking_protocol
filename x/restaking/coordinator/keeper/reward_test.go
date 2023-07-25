package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
)

func (s *KeeperTestSuite) TestOnOperatorReceiveAllRewards() {
	operator := s.mockOperator()
	ctx, keeper := s.ctx, s.coordinatorKeeper

	delegateAmt := math.NewIntFromUint64(100000)
	operatorAccAddr := sdk.MustAccAddressFromBech32(operator.OperatorAddress)
	operator.Shares = delegateAmt
	operator.RestakedAmount = delegateAmt
	keeper.SetOperator(ctx, operatorAccAddr, operator)

	rewards := []sdk.Coin{
		sdk.NewCoin("denom1", sdk.MustNewDecFromStr("10000000000").TruncateInt()),
		sdk.NewCoin("denom2", sdk.MustNewDecFromStr("1000100000011").TruncateInt()),
	}

	// receive first period reward
	keeper.OnOperatorReceiveAllRewards(ctx, operatorAccAddr, rewards)
	lastPeriod, found := keeper.GetOperatorLastRewardPeriod(ctx, operatorAccAddr)
	s.Require().True(found, "period not advance after receive all rewards")
	s.Require().Equal(lastPeriod, uint64(1), "mismatch period number")

	historical, found := keeper.GetOperatorHistoricalRewards(ctx, 0, operatorAccAddr)
	s.Require().True(found, "historical not found after receive all reward")
	s.Require().Equal(len(historical.CumulativeRewardRatios), 2)
	s.Require().True(historical.CumulativeRewardRatios[0].Amount.Equal(
		sdk.NewDecFromInt(rewards[0].Amount).QuoInt(operator.RestakedAmount)))
	s.Require().True(historical.CumulativeRewardRatios[1].Amount.Equal(
		sdk.NewDecFromInt(rewards[1].Amount).QuoInt(operator.RestakedAmount)))

	// receive second period reward
	rewards2 := []sdk.Coin{
		sdk.NewCoin("denom1", sdk.MustNewDecFromStr("3030303030").TruncateInt()),
		sdk.NewCoin("denom2", sdk.MustNewDecFromStr("1010101010111").TruncateInt()),
	}

	keeper.OnOperatorReceiveAllRewards(ctx, operatorAccAddr, rewards2)

	lastPeriod, found = keeper.GetOperatorLastRewardPeriod(ctx, operatorAccAddr)
	s.Require().True(found, "period not advance after receive all rewards")
	s.Require().Equal(lastPeriod, uint64(2), "mismatch period number")

	historical, found = keeper.GetOperatorHistoricalRewards(ctx, 1, operatorAccAddr)
	s.Require().True(found, "historical not found after receive all reward")
	s.Require().Equal(len(historical.CumulativeRewardRatios), 2)
	s.Require().True(historical.CumulativeRewardRatios[0].Amount.Equal(
		sdk.NewDecFromInt(rewards2[0].Amount.Add(rewards[0].Amount)).QuoInt(operator.RestakedAmount)))
	s.Require().True(historical.CumulativeRewardRatios[1].Amount.Equal(
		sdk.NewDecFromInt(rewards2[1].Amount.Add(rewards[1].Amount)).QuoInt(operator.RestakedAmount)))
}

func (s *KeeperTestSuite) TestRewardDistributionWithoutSlash() {
	operator := s.mockOperator()
	ctx, keeper := s.ctx, s.coordinatorKeeper

	accounts := simtestutil.CreateIncrementalAccounts(2)
	delegatorAccAddr := accounts[0]
	operatorAccAddr := sdk.MustAccAddressFromBech32(operator.OperatorAddress)
	delegateAmt := math.NewIntFromUint64(10000000000)

	delegate := func(delegator, operator sdk.AccAddress, amount math.Int) {
		delegateCoins := sdk.Coins{sdk.NewCoin("stake", amount)}
		s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), delegator, types.ModuleName, delegateCoins)
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, operator, delegateCoins)
		keeper.Delegate(ctx, delegator, operator, amount)
	}

	delegate(delegatorAccAddr, operatorAccAddr, delegateAmt)

	operator.RestakedAmount = delegateAmt
	operator.Shares = delegateAmt
	keeper.SetOperator(ctx, operatorAccAddr, operator)

	rewards := []sdk.Coin{
		sdk.NewCoin("denom1", sdk.MustNewDecFromStr("10000000000").TruncateInt()),
		sdk.NewCoin("denom2", sdk.MustNewDecFromStr("1000100000011").TruncateInt()),
	}

	keeper.OnOperatorReceiveAllRewards(ctx, operatorAccAddr, rewards)

	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), operatorAccAddr, types.ModuleName, rewards)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), types.ModuleName, delegatorAccAddr, rewards)

	keeper.WithdrawDelegatorRewards(ctx, delegatorAccAddr, operatorAccAddr)
}
