package mtstaking_test

import (
	"time"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	app "github.com/celinium-network/restaking_protocol/app/consumer"
)

var (
	PKs                     = simtestutil.CreateTestPubKeys(500)
	mockMultiRestakingDenom = "non_native_token"
)

func createValAddrs(count int) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrs := app.CreateIncrementalAccounts(count)
	valAddrs := app.ConvertAddrsToValAddrs(addrs)

	return addrs, valAddrs
}

// TODO Distinguish between integration tests and unit tests. it's look like integration tests?
func (suite *KeeperTestSuite) TestDelegate() {
	delegatorAddrs, _ := createValAddrs(2)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))
	valAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	suite.Require().NoError(err)

	suite.app.MTStakingKeeper.AddMTStakingDenom(suite.ctx, mockMultiRestakingDenom)
	suite.app.MTStakingKeeper.SetEquivalentNativeCoinMultiplier(suite.ctx, 1, mockMultiRestakingDenom, sdk.MustNewDecFromStr("1"))
	suite.mintCoin(multiRestakingCoin, delegatorAddrs[0])
	_, err = suite.app.MTStakingKeeper.MTStakingDelegate(suite.ctx, delegatorAddrs[0], valAddr, multiRestakingCoin)
	suite.NoError(err)

	agents := suite.app.MTStakingKeeper.GetAllAgent(suite.ctx)
	suite.Require().Equal(len(agents), 1)

	agentAccAddr := sdk.MustAccAddressFromBech32(agents[0].AgentAddress)
	delegatorShares := suite.app.MTStakingKeeper.GetDelegatorAgentShares(suite.ctx, agentAccAddr, delegatorAddrs[0])
	suite.Require().True(delegatorShares.Equal(multiRestakingCoin.Amount))

	agent, found := suite.app.MTStakingKeeper.GetMTStakingAgentByAddress(suite.ctx, agentAccAddr)
	suite.Require().True(found)
	suite.Require().True(agent.StakedAmount.Equal(multiRestakingCoin.Amount))
	suite.Require().True(agent.Shares.Equal(delegatorShares))

	suite.mintCoin(multiRestakingCoin, delegatorAddrs[1])
	_, err = suite.app.MTStakingKeeper.MTStakingDelegate(suite.ctx, delegatorAddrs[1], valAddr, multiRestakingCoin)
	suite.NoError(err)

	delegator2Shares := suite.app.MTStakingKeeper.GetDelegatorAgentShares(suite.ctx, agentAccAddr, delegatorAddrs[1])
	suite.Require().True(delegator2Shares.Equal(multiRestakingCoin.Amount))
	agent, found = suite.app.MTStakingKeeper.GetMTStakingAgentByAddress(suite.ctx, agentAccAddr)
	suite.Require().True(found)
	suite.Require().True(agent.StakedAmount.Equal(multiRestakingCoin.Amount.MulRaw(2)))
	suite.Require().True(agent.Shares.Equal(delegatorShares.MulRaw(2)))
}

func (suite *KeeperTestSuite) TestUndelegate() {
	delegatorAddrs, _ := createValAddrs(1)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	valAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	suite.Require().NoError(err)

	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))
	suite.mintCoin(multiRestakingCoin, delegatorAddrs[0])
	suite.app.MTStakingKeeper.AddMTStakingDenom(suite.ctx, mockMultiRestakingDenom)
	suite.app.MTStakingKeeper.SetEquivalentNativeCoinMultiplier(suite.ctx, 1, mockMultiRestakingDenom, sdk.MustNewDecFromStr("1"))

	_, err = suite.app.MTStakingKeeper.MTStakingDelegate(suite.ctx, delegatorAddrs[0], valAddr, multiRestakingCoin)
	suite.Require().NoError(err)

	_, err = suite.app.MTStakingKeeper.MTStakingUndelegate(suite.ctx, delegatorAddrs[0], valAddr, multiRestakingCoin)
	suite.Require().NoError(err)

	agents := suite.app.MTStakingKeeper.GetAllAgent(suite.ctx)
	suite.Require().Equal(len(agents), 1)
	agentAccAddr := sdk.MustAccAddressFromBech32(agents[0].AgentAddress)

	delegator2Shares := suite.app.MTStakingKeeper.GetDelegatorAgentShares(suite.ctx, agentAccAddr, delegatorAddrs[0])
	suite.Require().True(delegator2Shares.Equal(math.ZeroInt()))
	agent, found := suite.app.MTStakingKeeper.GetMTStakingAgentByAddress(suite.ctx, agentAccAddr)
	suite.Require().True(found)
	suite.Require().True(agent.StakedAmount.Equal(math.ZeroInt()))
	suite.Require().True(agent.Shares.Equal(math.ZeroInt()))

	// check unbonding records
	unbonding, found := suite.app.MTStakingKeeper.GetMTStakingUnbonding(suite.ctx, agentAccAddr, delegatorAddrs[0])
	suite.Require().True(found)
	suite.Require().Equal(len(unbonding.Entries), 1)

	entry := unbonding.Entries[0]
	suite.Require().True(entry.Balance.Equal(multiRestakingCoin))
	suite.Require().True(entry.InitialBalance.Equal(multiRestakingCoin))

	unbondingTime := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime
	suite.Require().True(entry.CompletionTime.Equal(suite.ctx.BlockTime().Add(unbondingTime)))

	unbondingQueue := suite.app.MTStakingKeeper.GetUBDQueueTimeSlice(suite.ctx, entry.CompletionTime)
	suite.Require().Equal(len(unbondingQueue), 1)
	unbondingDAPair := unbondingQueue[0]
	suite.Require().Equal(unbondingDAPair.AgentAddress, string(agentAccAddr))
	suite.Require().Equal(unbondingDAPair.DelegatorAddress, string(delegatorAddrs[0]))
}

func (suite *KeeperTestSuite) TestUndelegateReward() {
	delegatorAddrs, _ := createValAddrs(1)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	valAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	suite.Require().NoError(err)

	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))
	suite.mintCoin(multiRestakingCoin, delegatorAddrs[0])
	suite.app.MTStakingKeeper.AddMTStakingDenom(suite.ctx, mockMultiRestakingDenom)
	suite.app.MTStakingKeeper.SetEquivalentNativeCoinMultiplier(suite.ctx, 1, mockMultiRestakingDenom, sdk.MustNewDecFromStr("1"))

	_, err = suite.app.MTStakingKeeper.MTStakingDelegate(suite.ctx, delegatorAddrs[0], valAddr, multiRestakingCoin)
	suite.Require().NoError(err)

	rewardAmount := sdk.NewIntFromUint64(500000)
	rewardDenom := suite.app.StakingKeeper.GetParams(suite.ctx).BondDenom
	rewardCoins := sdk.Coins{sdk.NewCoin(rewardDenom, rewardAmount)}

	err = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, rewardCoins)
	suite.Require().NoError(err)

	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, distrtypes.ModuleName, rewardCoins)
	suite.Require().NoError(err)

	validator := suite.app.StakingKeeper.Validator(suite.ctx, valAddr)

	suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, sdk.DecCoins{
		sdk.NewDecCoinFromDec(rewardDenom, sdk.NewDecFromInt(rewardAmount)),
	})

	suite.ctx = suite.ctx.
		WithBlockHeight(suite.ctx.BlockHeight() + 100).
		WithBlockTime(suite.ctx.BlockTime().Add(time.Hour))

	_, err = suite.app.MTStakingKeeper.MTStakingUndelegate(suite.ctx, delegatorAddrs[0], valAddr, multiRestakingCoin)
	suite.Require().NoError(err)

	// TODO check reward amount
}
