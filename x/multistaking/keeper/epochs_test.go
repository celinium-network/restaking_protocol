package keeper_test

import (
	"github.com/celinium-network/restaking_protocol/x/multistaking/types"
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
	suite.app.MultiStakingKeeper.SetMultiStakingDenom(suite.ctx, mockMultiRestakingDenom)

	err := suite.app.MultiStakingKeeper.MultiStakingDelegate(suite.ctx, types.MsgMultiStakingDelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	})
	suite.Require().NoError(err)

	suite.app.MultiStakingKeeper.EquivalentCoinCalculator = RiseRateCalculateEquivalentCoin
	suite.app.MultiStakingKeeper.RefreshAgentDelegationAmount(suite.ctx)

	increaseCoin, _ := RiseRateCalculateEquivalentCoin(suite.ctx, multiRestakingCoin, defaultBondDenom)

	agentID := suite.app.MultiStakingKeeper.GetLatestMultiStakingAgentID(suite.ctx)
	agent, found := suite.app.MultiStakingKeeper.GetMultiStakingAgentByID(suite.ctx, agentID)
	suite.Require().True(found)
	suite.Require().True(agent.Shares.Equal(multiRestakingCoin.Amount))
	suite.Require().True(agent.StakedAmount.Equal(multiRestakingCoin.Amount))

	agentDelegateAccAddr := sdk.MustAccAddressFromBech32(agent.DelegateAddress)
	valAddr, _ := sdk.ValAddressFromBech32(validators[0].OperatorAddress)

	v, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	suite.True(found)
	delegation, found := suite.app.StakingKeeper.GetDelegation(suite.ctx, agentDelegateAccAddr, valAddr)
	suite.Require().True(found)
	token := v.TokensFromShares(delegation.Shares)
	suite.Require().True(increaseCoin.Amount.Equal(token.TruncateInt()))
}
