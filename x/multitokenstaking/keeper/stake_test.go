package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	mtstakingtypes "github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

var (
	defaultBondDenom = "test1"
	mtStakingDenom   = "test2"
	pks              = simtestutil.CreateTestPubKeys(5)
	accounts         = simtestutil.CreateIncrementalAccounts(2)
)

func mustNewIntForStr(str string) math.Int {
	return math.LegacyMustNewDecFromStr("100000000").TruncateInt()
}

func (s *KeeperTestSuite) TestMTStakingDelegate() {
	expectKeeper := func(
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

	valAddr := sdk.ValAddress(pks[0].Address())
	validator := stakingtypes.Validator{
		OperatorAddress: valAddr.String(),
	}

	// TODO more unit tests
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
		if t.isExistedAgent {
			s.mtStakingKeeper.SetMTStakingAgent(s.ctx, &t.agent)
			s.mtStakingKeeper.IncreaseDelegatorAgentShares(s.ctx, t.existedDelegatorShares, t.agent.AgentAddress, delegatorAddress)
		}

		delegateCoin := sdk.NewCoin(mtStakingDenom, t.delegateAmount)
		eqCoin := sdk.NewCoin(defaultBondDenom, t.toDefaultDenomMultiplier.MulInt(t.delegateAmount).TruncateInt())
		agentAccAddr := s.mtStakingKeeper.GenerateAccount(s.ctx, mtStakingDenom, t.validator.OperatorAddress).GetAddress()

		expectKeeper(delegateCoin, t.validator, t.delegatorAccAddr, eqCoin, agentAccAddr)

		err := s.mtStakingKeeper.MTStakingDelegate(s.ctx, mtstakingtypes.MsgMTStakingDelegate{
			DelegatorAddress: delegatorAddress,
			ValidatorAddress: t.validator.OperatorAddress,
			Balance:          delegateCoin,
		})

		s.Require().NoError(err, t.describe)

		agent, found := s.mtStakingKeeper.GetMTStakingAgent(s.ctx, mtStakingDenom, t.validator.OperatorAddress)
		s.Require().True(found, t.describe, "agent not exist after delegate successfully")
		s.Require().Equal(agent.StakeDenom, mtStakingDenom, t.describe, "agent has mismatch stakeDenom")
		s.Require().Equal(agent.ValidatorAddress, t.validator.OperatorAddress, t.describe, "agent has mismatch Shares")
		s.Require().True(agent.StakedAmount.Equal(t.expectedAgent.StakedAmount), t.describe, "agent has mismatch stakedAmount")
		s.Require().True(agent.Shares.Equal(t.expectedAgent.Shares), t.describe, "agent has mismatch Shares")

		shares := s.mtStakingKeeper.GetDelegatorAgentShares(s.ctx, agent.AgentAddress, delegatorAddress)
		s.Require().True(shares.Equal(agent.Shares), t.describe, "delegator has mismatch Shares")
	}
}
