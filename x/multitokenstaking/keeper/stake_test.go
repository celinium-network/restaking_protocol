package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	mtstakingtypes "github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

var (
	defaultBondDenom  = "test1"
	defaultUnbondTime = time.Minute * 10
	mtStakingDenom    = "test2"
	pks               = simtestutil.CreateTestPubKeys(5)
	accounts          = simtestutil.CreateIncrementalAccounts(2)
)

func mustNewIntForStr(str string) math.Int {
	return math.LegacyMustNewDecFromStr(str).TruncateInt()
}

func (s *KeeperTestSuite) delegateExpectOtherKeeperAction(
	delegateCoin sdk.Coin,
	validator stakingtypes.Validator,
	delegator sdk.AccAddress,
	eqCoin sdk.Coin,
	agentAccAddr sdk.AccAddress,
) {
	valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	s.Require().NoError(err, validator.OperatorAddress, "is invalid validator address")

	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
	s.stakingKeeper.EXPECT().GetValidator(gomock.Any(), valAddr).Return(validator, true)
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), delegator, mtstakingtypes.ModuleName, sdk.Coins{delegateCoin}).Return(nil)
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), mtstakingtypes.ModuleName, agentAccAddr, sdk.Coins{delegateCoin}).Return(nil)
	s.bankKeeper.EXPECT().MintCoins(gomock.Any(), mtstakingtypes.ModuleName, sdk.Coins{eqCoin})
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), mtstakingtypes.ModuleName, agentAccAddr, sdk.Coins{eqCoin})
	s.stakingKeeper.EXPECT().Delegate(gomock.Any(), agentAccAddr, eqCoin.Amount, stakingtypes.Unbonded, gomock.Any(), true)
}

func (s *KeeperTestSuite) TestMTStakingDelegate() {
	valAddr := sdk.ValAddress(pks[0].Address())
	validator := stakingtypes.Validator{
		OperatorAddress: valAddr.String(),
	}

	// TODO more test case, such as 1)multi agent
	tests := []struct {
		describe                 string
		isExistedAgent           bool
		agent                    mtstakingtypes.MTStakingAgent
		delegatorAccAddr         sdk.AccAddress
		existedDelegatorShares   math.Int
		validator                stakingtypes.Validator
		delegateAmount           math.Int
		toDefaultDenomMultiplier math.LegacyDec
		expectedDelegatorShares  math.Int
		expectedAgent            mtstakingtypes.MTStakingAgent
	}{
		{
			"delegate without agent",
			false,
			mtstakingtypes.MTStakingAgent{},
			accounts[0],
			math.ZeroInt(),
			validator,
			mustNewIntForStr("100000000"),
			math.LegacyMustNewDecFromStr("2"),
			mustNewIntForStr("100000000"),
			mtstakingtypes.MTStakingAgent{
				StakedAmount: mustNewIntForStr("100000000"),
				Shares:       mustNewIntForStr("100000000"),
			},
		},
		{
			"overflow test", // TODO use bigger rand u256 do overflow test ?
			true,
			mtstakingtypes.MTStakingAgent{
				StakeDenom:   mtStakingDenom,
				StakedAmount: mustNewIntForStr("340282366920938463463374607431768211455"), // u128max
				Shares:       mustNewIntForStr("340282366920938463463374607431768211455"),
			},
			accounts[0],
			mustNewIntForStr("1000000000000000000000000000"),
			validator,
			mustNewIntForStr("2000000000000000000000000000"),
			math.LegacyMustNewDecFromStr("2"),
			mustNewIntForStr("3000000000000000000000000000"),
			mtstakingtypes.MTStakingAgent{
				StakedAmount: mustNewIntForStr("340282366922938463463374607431768211455"), // u128max +  3000000000000000000000000000
				Shares:       mustNewIntForStr("340282366922938463463374607431768211455"),
			},
		},
	}

	for _, t := range tests {
		s.SetupTest()
		s.mtStakingKeeper.AddMTStakingDenom(s.ctx, mtStakingDenom)
		s.mtStakingKeeper.SetEquivalentNativeCoinMultiplier(s.ctx, 1, mtStakingDenom, t.toDefaultDenomMultiplier)

		delegatorAddress := t.delegatorAccAddr.String()
		agentAccAddr := s.mtStakingKeeper.GenerateAccount(s.ctx, mtStakingDenom, t.validator.OperatorAddress).GetAddress()
		if t.isExistedAgent {
			t.agent.AgentAddress = agentAccAddr.String()
			t.agent.ValidatorAddress = t.validator.OperatorAddress
			s.mtStakingKeeper.SetMTStakingAgent(s.ctx, agentAccAddr, &t.agent)
			s.mtStakingKeeper.SetMTStakingDenomAndValWithAgentAddress(s.ctx, agentAccAddr, mtStakingDenom, valAddr)
			s.mtStakingKeeper.IncreaseDelegatorAgentShares(s.ctx, t.existedDelegatorShares, agentAccAddr, t.delegatorAccAddr)
		}

		delegateCoin := sdk.NewCoin(mtStakingDenom, t.delegateAmount)
		eqCoin := sdk.NewCoin(defaultBondDenom, t.toDefaultDenomMultiplier.MulInt(t.delegateAmount).TruncateInt())

		s.delegateExpectOtherKeeperAction(delegateCoin, t.validator, t.delegatorAccAddr, eqCoin, agentAccAddr)

		err := s.mtStakingKeeper.MTStakingDelegate(s.ctx, mtstakingtypes.MsgMTStakingDelegate{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: t.validator.OperatorAddress,
			Balance:          delegateCoin,
		})

		s.Require().NoError(err, t.describe)

		agent, found := s.mtStakingKeeper.GetMTStakingAgent(s.ctx, mtStakingDenom, valAddr)
		s.Require().True(found, t.describe, "agent not exist after delegate successfully")
		s.Require().Equal(agent.StakeDenom, mtStakingDenom, t.describe, "agent has mismatch stakeDenom")
		s.Require().Equal(agent.ValidatorAddress, t.validator.OperatorAddress, t.describe, "agent has mismatch Shares")
		s.Require().True(agent.StakedAmount.Equal(t.expectedAgent.StakedAmount), t.describe, "agent has mismatch stakedAmount")
		s.Require().True(agent.Shares.Equal(t.expectedAgent.Shares), t.describe, "agent has mismatch Shares")

		shares := s.mtStakingKeeper.GetDelegatorAgentShares(s.ctx, agentAccAddr, t.delegatorAccAddr)
		if t.isExistedAgent {
			s.Require().True(shares.Sub(t.existedDelegatorShares).Equal(t.agent.Shares.Mul(t.delegateAmount).Quo(t.agent.StakedAmount)), t.describe, "delegator has mismatch Shares")
		} else {
			s.Require().True(shares.Equal(t.expectedAgent.Shares))
		}

	}
}

func (s *KeeperTestSuite) TestMTStakingUndelegate() {
	expectOtherKeeperAction := func(
		validator stakingtypes.Validator,
		delegator sdk.AccAddress,
		delegatorShares math.Int,
		undelegateAmount math.Int,
		agent mtstakingtypes.MTStakingAgent,
		multiplier math.LegacyDec,
	) {
		agentAccAddr := sdk.MustAccAddressFromBech32(agent.AgentAddress)
		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		s.Require().NoError(err, validator.OperatorAddress, "is invalid validator address")

		rewardAmount := mustNewIntForStr("340282366920938463463374607431753211455")
		rewardCoins := sdk.Coins{sdk.NewCoin(defaultBondDenom, rewardAmount)}
		delegatorRewardAmount := rewardAmount.Mul(delegatorShares).Quo(agent.Shares)
		delegatorRewardCoins := sdk.Coins{sdk.NewCoin(defaultBondDenom, delegatorRewardAmount)}
		agentUndelegateAmount := multiplier.MulInt(undelegateAmount).TruncateInt()
		unbondShares := validator.DelegatorShares.MulInt(agentUndelegateAmount).QuoInt(validator.Tokens)
		unbondCoins := sdk.Coins{sdk.NewCoin(defaultBondDenom, unbondShares.TruncateInt())}

		var unbondedPoolName string
		if validator.IsBonded() {
			unbondedPoolName = stakingtypes.BondedPoolName
		} else {
			unbondedPoolName = stakingtypes.NotBondedPoolName
		}

		s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
		s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
		s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
		s.distributerKeeper.EXPECT().WithdrawDelegationRewards(gomock.Any(), agentAccAddr, valAddr).Return(rewardCoins, nil)
		s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), agentAccAddr, mtstakingtypes.ModuleName, delegatorRewardCoins).Return(nil)
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), mtstakingtypes.ModuleName, delegator, delegatorRewardCoins).Return(nil)
		s.stakingKeeper.EXPECT().ValidateUnbondAmount(gomock.Any(), agentAccAddr, valAddr, agentUndelegateAmount).Return(unbondShares, nil)
		s.stakingKeeper.EXPECT().GetValidator(gomock.Any(), valAddr).Return(validator, true)
		s.stakingKeeper.EXPECT().Unbond(gomock.Any(), agentAccAddr, valAddr, unbondShares).Return(unbondShares.TruncateInt(), nil)
		s.stakingKeeper.EXPECT().GetParams(gomock.Any()).Return(stakingtypes.Params{UnbondingTime: defaultUnbondTime})
		s.bankKeeper.EXPECT().UndelegateCoinsFromModuleToAccount(gomock.Any(), unbondedPoolName, agentAccAddr, unbondCoins).Return(nil)
		s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), agentAccAddr, mtstakingtypes.ModuleName, unbondCoins).Return(nil)
		s.bankKeeper.EXPECT().BurnCoins(gomock.Any(), mtstakingtypes.ModuleName, unbondCoins)
	}

	valAddr := sdk.ValAddress(pks[0].Address())

	// TODO more test case, such as 1)multi agent
	tests := []struct {
		describe                 string
		isExistedAgent           bool
		agent                    mtstakingtypes.MTStakingAgent
		delegatorAccAddr         sdk.AccAddress
		delegatorShares          math.Int
		validator                stakingtypes.Validator
		undelegateAmount         math.Int
		toDefaultDenomMultiplier math.LegacyDec
		expectedDelegatorShares  math.Int
		expectedAgent            mtstakingtypes.MTStakingAgent
		expectedError            error
	}{
		{
			"undelegate should work",
			true,
			mtstakingtypes.MTStakingAgent{
				StakedAmount:     mustNewIntForStr("100000000"),
				Shares:           mustNewIntForStr("100000000"),
				ValidatorAddress: valAddr.String(),
				StakeDenom:       mtStakingDenom,
			},
			accounts[0],
			mustNewIntForStr("50000000"),
			stakingtypes.Validator{
				OperatorAddress: valAddr.String(),
				Tokens:          mustNewIntForStr("100000000"),
				DelegatorShares: sdk.MustNewDecFromStr("100000000"),
				Status:          stakingtypes.Unbonded, // unbonded validator
			},
			mustNewIntForStr("15000000"),
			math.LegacyMustNewDecFromStr("2"),
			mustNewIntForStr("35000000"),
			mtstakingtypes.MTStakingAgent{
				StakedAmount: mustNewIntForStr("85000000"),
				Shares:       mustNewIntForStr("85000000"),
			},
			nil,
		},
		{
			"overflow test", // TODO use bigger rand u256 do overflow test ?
			true,
			mtstakingtypes.MTStakingAgent{
				StakedAmount:     mustNewIntForStr("340282366920938463463374607431768211455"),
				Shares:           mustNewIntForStr("340282366920938463463374607431768211455"),
				ValidatorAddress: valAddr.String(),
				StakeDenom:       mtStakingDenom,
			},
			accounts[0],
			mustNewIntForStr("50000000"),
			stakingtypes.Validator{
				OperatorAddress: valAddr.String(),
				Tokens:          mustNewIntForStr("340282366920938463463374607431768211455"),
				DelegatorShares: sdk.MustNewDecFromStr("340282366920938463463374607431768211455"),
				Status:          stakingtypes.Unbonded, // unbonded validator
			},
			mustNewIntForStr("15000000"),
			math.LegacyMustNewDecFromStr("2"),
			mustNewIntForStr("35000000"),
			mtstakingtypes.MTStakingAgent{
				StakedAmount: mustNewIntForStr("340282366920938463463374607431753211455"),
				Shares:       mustNewIntForStr("340282366920938463463374607431753211455"),
			},
			nil,
		},
	}

	for _, t := range tests {
		s.SetupTest()
		delegatorAddress := t.delegatorAccAddr.String()
		agentAccAddr := s.mtStakingKeeper.GenerateAccount(s.ctx, mtStakingDenom, t.validator.OperatorAddress).GetAddress()
		t.agent.AgentAddress = agentAccAddr.String()

		s.mtStakingKeeper.AddMTStakingDenom(s.ctx, mtStakingDenom)
		s.mtStakingKeeper.SetEquivalentNativeCoinMultiplier(s.ctx, 1, mtStakingDenom, t.toDefaultDenomMultiplier)

		if t.isExistedAgent {
			s.mtStakingKeeper.SetMTStakingAgent(s.ctx, agentAccAddr, &t.agent)
			s.mtStakingKeeper.SetMTStakingDenomAndValWithAgentAddress(s.ctx, agentAccAddr, mtStakingDenom, valAddr)
			s.mtStakingKeeper.IncreaseDelegatorAgentShares(s.ctx, t.delegatorShares, agentAccAddr, t.delegatorAccAddr)
		}

		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100)

		expectOtherKeeperAction(
			t.validator,
			t.delegatorAccAddr,
			t.delegatorShares,
			t.undelegateAmount,
			t.agent,
			t.toDefaultDenomMultiplier,
		)

		undelegateCoin := sdk.NewCoin(mtStakingDenom, t.undelegateAmount)
		err := s.mtStakingKeeper.MTStakingUndelegate(s.ctx, &mtstakingtypes.MsgMTStakingUndelegate{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: t.validator.OperatorAddress,
			Balance:          undelegateCoin,
		})
		if t.expectedError != nil {
			s.Require().Equal(err, t.expectedAgent)
			continue
		}
		s.Require().NoError(err, t.describe)

		agent, found := s.mtStakingKeeper.GetMTStakingAgent(s.ctx, mtStakingDenom, valAddr)
		s.Require().True(found, t.describe, "agent not exist after delegate successfully")
		s.Require().Equal(agent.StakeDenom, mtStakingDenom, t.describe, "agent has mismatch stakeDenom")
		s.Require().Equal(agent.ValidatorAddress, t.validator.OperatorAddress, t.describe, "agent has mismatch Shares")
		s.Require().True(agent.StakedAmount.Equal(t.expectedAgent.StakedAmount), t.describe, "agent has mismatch stakedAmount")
		s.Require().True(agent.Shares.Equal(t.expectedAgent.Shares), t.describe, "agent has mismatch Shares")
		shares := s.mtStakingKeeper.GetDelegatorAgentShares(s.ctx, agentAccAddr, t.delegatorAccAddr)
		s.Require().True(shares.Equal(t.expectedDelegatorShares), t.describe, "delegator has mismatch Shares")

		// check unbonding state
		unbonding, found := s.mtStakingKeeper.GetMTStakingUnbonding(s.ctx, agentAccAddr, t.delegatorAccAddr)
		s.Require().True(found, t.describe, "no unbonding")
		s.Require().Equal(len(unbonding.Entries), 1, t.describe, "unbonding entries len mismatch")
		s.Require().True(unbonding.Entries[0].CompletionTime.Equal(
			s.ctx.BlockTime().Add(defaultUnbondTime)),
			t.describe, "mismatch unbonding complete time")
		s.Require().True(unbonding.Entries[0].Balance.Amount.Equal(t.undelegateAmount), t.describe, "mismatch unbonding balance")
	}
}

func (s *KeeperTestSuite) TestRefreshAgentDelegationV2() {
	expectNothingChangedAction := func(
		validator stakingtypes.Validator,
		agentAccAddr sdk.AccAddress,
	) {
		valAddr := validator.GetOperator()
		s.stakingKeeper.EXPECT().GetValidator(gomock.Any(), valAddr).Return(validator, true)
		s.stakingKeeper.EXPECT().GetDelegation(gomock.Any(), agentAccAddr, valAddr).Return(stakingtypes.Delegation{
			DelegatorAddress: agentAccAddr.String(),
			ValidatorAddress: valAddr.String(),
			Shares:           validator.DelegatorShares,
		}, true)
		s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
	}

	expectMintAction := func(
		validator stakingtypes.Validator,
		agentAccAddr sdk.AccAddress,
		multiplier math.LegacyDec,
		delegationAmount math.Int,
	) {
		valAddr := validator.GetOperator()

		shares, err := validator.SharesFromTokens(delegationAmount)
		s.Require().NoError(err)
		s.stakingKeeper.EXPECT().GetValidator(gomock.Any(), valAddr).Return(validator, true)
		s.stakingKeeper.EXPECT().GetDelegation(gomock.Any(), agentAccAddr, valAddr).Return(stakingtypes.Delegation{
			DelegatorAddress: agentAccAddr.String(),
			ValidatorAddress: valAddr.String(),
			Shares:           shares,
		}, true)

		s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
		eqCoin := sdk.NewCoin(defaultBondDenom, multiplier.Sub(math.LegacyOneDec()).MulInt(delegationAmount).TruncateInt())
		s.bankKeeper.EXPECT().MintCoins(gomock.Any(), mtstakingtypes.ModuleName, sdk.Coins{eqCoin})
		s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), mtstakingtypes.ModuleName, agentAccAddr, sdk.Coins{eqCoin})
		s.stakingKeeper.EXPECT().Delegate(gomock.Any(), agentAccAddr, eqCoin.Amount, stakingtypes.Unbonded, gomock.Any(), true)
	}

	expectBurnAction := func(
		validator stakingtypes.Validator,
		agent mtstakingtypes.MTStakingAgent,
		multiplier math.LegacyDec,
		agentDelegationAmount math.Int,
	) {
		valAddr := validator.GetOperator()
		agentAccAddr := sdk.MustAccAddressFromBech32(agent.AgentAddress)

		unbondAmount := agentDelegationAmount.Sub(multiplier.MulInt(agentDelegationAmount).TruncateInt())
		unbondShares, err := validator.SharesFromTokens(unbondAmount)
		s.Require().NoError(err)
		unbondCoins := sdk.Coins{sdk.NewCoin(defaultBondDenom, unbondShares.TruncateInt())}
		var unbondedPoolName string
		if validator.IsBonded() {
			unbondedPoolName = stakingtypes.BondedPoolName
		} else {
			unbondedPoolName = stakingtypes.NotBondedPoolName
		}

		s.stakingKeeper.EXPECT().ValidateUnbondAmount(gomock.Any(), agentAccAddr, valAddr, unbondAmount).Return(unbondShares, nil)
		s.stakingKeeper.EXPECT().GetValidator(gomock.Any(), valAddr).Return(validator, true)
		s.stakingKeeper.EXPECT().GetDelegation(gomock.Any(), agentAccAddr, valAddr).Return(stakingtypes.Delegation{
			DelegatorAddress: agentAccAddr.String(),
			ValidatorAddress: valAddr.String(),
			Shares:           validator.DelegatorShares,
		}, true)
		s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
		s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return(defaultBondDenom)
		s.stakingKeeper.EXPECT().GetValidator(gomock.Any(), valAddr).Return(validator, true)
		s.stakingKeeper.EXPECT().Unbond(gomock.Any(), agentAccAddr, valAddr, unbondShares).Return(unbondShares.TruncateInt(), nil)
		s.bankKeeper.EXPECT().UndelegateCoinsFromModuleToAccount(gomock.Any(), unbondedPoolName, agentAccAddr, unbondCoins).Return(nil)
		s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), agentAccAddr, mtstakingtypes.ModuleName, unbondCoins).Return(nil)
		s.bankKeeper.EXPECT().BurnCoins(gomock.Any(), mtstakingtypes.ModuleName, unbondCoins)
	}

	valAddr := sdk.ValAddress(pks[0].Address())

	tests := []struct {
		describe   string
		agent      mtstakingtypes.MTStakingAgent
		validator  stakingtypes.Validator
		multiplier string
		amount     string
	}{
		{
			describe: "multiplier no change",
			agent: mtstakingtypes.MTStakingAgent{
				StakeDenom:       mtStakingDenom,
				ValidatorAddress: valAddr.String(),
				StakedAmount:     mustNewIntForStr("100000000"),
				Shares:           mustNewIntForStr("100000000"),
			},
			validator: stakingtypes.Validator{
				OperatorAddress: valAddr.String(),
				Jailed:          false,
				Status:          stakingtypes.Unbonded,
				Tokens:          mustNewIntForStr("100000000"),
				DelegatorShares: sdk.MustNewDecFromStr("100000000"),
			},
			multiplier: "1",
			amount:     "100000000",
		},
		{
			describe: "multiplier bigger then 1",
			agent: mtstakingtypes.MTStakingAgent{
				StakeDenom:       mtStakingDenom,
				ValidatorAddress: valAddr.String(),
				StakedAmount:     mustNewIntForStr("100000000"),
				Shares:           mustNewIntForStr("100000000"),
			},
			validator: stakingtypes.Validator{
				OperatorAddress: valAddr.String(),
				Jailed:          false,
				Status:          stakingtypes.Unbonded,
				Tokens:          mustNewIntForStr("100000000"),
				DelegatorShares: sdk.MustNewDecFromStr("100000000"),
			},
			multiplier: "1.5",
			amount:     "100000000",
		},
		{
			describe: "multiplier lower then 1",
			agent: mtstakingtypes.MTStakingAgent{
				StakeDenom:       mtStakingDenom,
				ValidatorAddress: valAddr.String(),
				StakedAmount:     mustNewIntForStr("100000000"),
				Shares:           mustNewIntForStr("100000000"),
			},
			validator: stakingtypes.Validator{
				OperatorAddress: valAddr.String(),
				Jailed:          false,
				Status:          stakingtypes.Unbonded,
				Tokens:          mustNewIntForStr("100000000"),
				DelegatorShares: sdk.MustNewDecFromStr("100000000"),
			},
			multiplier: "0.8",
			amount:     "100000000",
		},
	}

	for _, t := range tests {
		s.SetupTest()

		t.agent.AgentAddress = s.mtStakingKeeper.GenerateAccount(s.ctx, mtStakingDenom, t.validator.OperatorAddress).GetAddress().String()
		multiplier := sdk.MustNewDecFromStr(t.multiplier)
		s.mtStakingKeeper.SetEquivalentNativeCoinMultiplier(s.ctx, 1, mtStakingDenom, multiplier)
		agentAccAddr := sdk.MustAccAddressFromBech32(t.agent.AgentAddress)

		amount := mustNewIntForStr(t.amount)
		if multiplier.Equal(sdk.OneDec()) { //nolint:gocritic
			expectNothingChangedAction(t.validator, agentAccAddr)
		} else if multiplier.GT(sdk.OneDec()) {
			expectMintAction(t.validator, agentAccAddr, multiplier, amount)
		} else {
			expectBurnAction(t.validator, t.agent, multiplier, amount)
		}

		err := s.mtStakingKeeper.RefreshAgentDelegation(s.ctx, &t.agent)
		s.Require().NoError(err)
	}
}
