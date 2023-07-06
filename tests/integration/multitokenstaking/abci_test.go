package mtstaking_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) bootstrapABCITest() (delegator, validator string, unbondingCoin sdk.Coin) {
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

	return delegatorAddrs[0].String(), validators[0].OperatorAddress, multiRestakingCoin
}

func (suite *KeeperTestSuite) TestProcessCompleteUnbonding() {
	delegator, _, delCoin := suite.bootstrapABCITest()

	unbondingTime := suite.app.StakingKeeper.GetParams(suite.ctx).UnbondingTime
	completeTime := suite.ctx.BlockTime().Add(unbondingTime)
	suite.ctx = suite.ctx.WithBlockTime(completeTime)

	delegatorAccAddr := sdk.MustAccAddressFromBech32(delegator)
	balanceBeforeUBComplete := suite.app.BankKeeper.GetBalance(suite.ctx, delegatorAccAddr, mockMultiRestakingDenom)
	_, err := suite.app.MTStakingKeeper.EndBlocker(suite.ctx)
	suite.Require().NoError(err)

	balanceAfterUBComplete := suite.app.BankKeeper.GetBalance(suite.ctx, delegatorAccAddr, mockMultiRestakingDenom)

	suite.True(balanceAfterUBComplete.Sub(balanceBeforeUBComplete).Amount.Equal(delCoin.Amount))
}
