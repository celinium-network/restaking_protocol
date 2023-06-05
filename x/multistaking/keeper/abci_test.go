package keeper_test

import (
	"github.com/celinium-network/restaking_protocol/x/multistaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) bootstrapABCITest() (delegator, validator string, unbondingCoin sdk.Coin) {
	delegatorAddrs, _ := createValAddrs(1)
	validators := suite.app.StakingKeeper.GetAllValidators(suite.ctx)

	multiRestakingCoin := sdk.NewCoin(mockMultiRestakingDenom, sdk.NewInt(10000000))
	suite.mintCoin(multiRestakingCoin, delegatorAddrs[0])
	suite.app.MultiStakingKeeper.SetMultiStakingDenom(suite.ctx, mockMultiRestakingDenom)

	err := suite.app.MultiStakingKeeper.MultiStakingDelegate(suite.ctx, types.MsgMultiStakingDelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	})
	suite.Require().NoError(err)

	err = suite.app.MultiStakingKeeper.MultiStakingUndelegate(suite.ctx, &types.MsgMultiStakingUndelegate{
		DelegatorAddress: delegatorAddrs[0].String(),
		ValidatorAddress: validators[0].OperatorAddress,
		Amount:           multiRestakingCoin,
	})
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
	_, err := suite.app.MultiStakingKeeper.EndBlocker(suite.ctx)
	suite.Require().NoError(err)
	
	balanceAfterUBComplete := suite.app.BankKeeper.GetBalance(suite.ctx, delegatorAccAddr, mockMultiRestakingDenom)

	suite.True(balanceAfterUBComplete.Sub(balanceBeforeUBComplete).Amount.Equal(delCoin.Amount))
}
