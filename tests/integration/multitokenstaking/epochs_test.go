package mtstaking_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RiseRateCalculateEquivalentCoin(ctx sdk.Context, denom string) (sdk.Dec, error) {
	return sdk.NewDec(2), nil
}

func DeclineRateCalculateEquivalentCoin(ctx sdk.Context, denom string) (sdk.Dec, error) {
	return sdk.MustNewDecFromStr("0.5"), nil
}

func (suite *KeeperTestSuite) TestRefreshDelegationAmountWhenRateRise() {
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

	suite.app.MTStakingKeeper.RefreshAllAgentDelegation(suite.ctx)

	// increaseCoin, _ := RiseRateCalculateEquivalentCoin(suite.ctx, mockMultiRestakingDenom)

	// agents := suite.app.MTStakingKeeper.GetAllAgent(suite.ctx)
	// suite.Require().Equal(len(agents), 1)

	// agent, found := suite.app.MTStakingKeeper.GetMTStakingAgentByAddress(suite.ctx, agents[0].AgentAddress)
	// suite.Require().True(found)
	// suite.Require().True(agent.Shares.Equal(multiRestakingCoin.Amount))
	// suite.Require().True(agent.StakedAmount.Equal(multiRestakingCoin.Amount))

	// agentDelegateAccAddr := sdk.MustAccAddressFromBech32(agent.AgentAddress)
	// valAddr, _ := sdk.ValAddressFromBech32(validators[0].OperatorAddress)

	// v, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	// suite.True(found)
	// delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, agentDelegateAccAddr, valAddr)
	// suite.Require().True(found)
	// token := v.TokensFromShares(delegation.Shares)
	// suite.Require().True(increaseCoin.Equal(token.TruncateInt()))
}
