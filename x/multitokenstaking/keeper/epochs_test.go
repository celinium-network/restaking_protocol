package keeper_test

import (
	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RiseRateCalculateEquivalentCoin(ctx sdk.Context, coin sdk.Coin, targetDenom string) (sdk.Coin, error) {
	return sdk.NewCoin(targetDenom, coin.Amount.MulRaw(2)), nil
}

func DeclineRateCalculateEquivalentCoin(ctx sdk.Context, coin sdk.Coin, targetDenom string) (sdk.Coin, error) {
	return sdk.NewCoin(targetDenom, coin.Amount.QuoRaw(2)), nil
}

func (suite *KeeperTestSuite) TestRefreshDelegationAmountWhenRateRise() {
	delegatorAddrs, _ := createValAddrs(1)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)
	defaultBondDenom := suite.app.StakingKeeper.BondDenom(suite.ctx)

	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))
	suite.mintCoin(multiRestakingCoin, delegatorAddrs[0])
	suite.app.MTStakingKeeper.SetMTStakingDenom(suite.ctx, mockMultiRestakingDenom)

	err := suite.app.MTStakingKeeper.MTStakingDelegate(suite.ctx, types.MsgMTStakingDelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Balance:          multiRestakingCoin,
	})
	suite.Require().NoError(err)

	suite.app.MTStakingKeeper.EquivalentCoinCalculator = RiseRateCalculateEquivalentCoin
	suite.app.MTStakingKeeper.RefreshAgentDelegationAmount(suite.ctx)

	increaseCoin, _ := RiseRateCalculateEquivalentCoin(suite.ctx, multiRestakingCoin, defaultBondDenom)

	agents := suite.app.MTStakingKeeper.GetAllAgent(suite.ctx)
	suite.Require().Equal(len(agents), 1)

	agent, found := suite.app.MTStakingKeeper.GetMTStakingAgentByAddress(suite.ctx, agents[0].AgentAddress)
	suite.Require().True(found)
	suite.Require().True(agent.Shares.Equal(multiRestakingCoin.Amount))
	suite.Require().True(agent.StakedAmount.Equal(multiRestakingCoin.Amount))

	agentDelegateAccAddr := sdk.MustAccAddressFromBech32(agent.AgentAddress)
	valAddr, _ := sdk.ValAddressFromBech32(validators[0].OperatorAddress)

	v, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	suite.True(found)
	delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, agentDelegateAccAddr, valAddr)
	suite.Require().True(found)
	token := v.TokensFromShares(delegation.Shares)
	suite.Require().True(increaseCoin.Amount.Equal(token.TruncateInt()))
}