package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

func (k Keeper) distributeDelegatorReward(
	ctx sdk.Context,
	delegatorAccAddr sdk.AccAddress,
	agentAccAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	agent *types.MTStakingAgent,
) (rewardCoin sdk.Coin, err error) {
	nativeCoinDenom := k.stakingKeeper.BondDenom(ctx)
	rewards, err := k.distributionKeeper.WithdrawDelegationRewards(ctx, agentAccAddr, valAddr)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("withdraw delegation rewards failed %s", err))
		return rewardCoin, err
	}

	delegatorShares := k.GetDelegatorAgentShares(ctx, agentAccAddr, delegatorAccAddr)
	if delegatorShares.IsZero() {
		return rewardCoin, err
	}

	// Currently only native coin rewards are considered
	agent.RewardAmount = agent.RewardAmount.Add(rewards.AmountOf(nativeCoinDenom))
	curBlockHeight := ctx.BlockHeight()
	rewardDuration := curBlockHeight - agent.CreatedBlockHeight
	// TODO rewardDuration == 0, maybe let delegator get all reward, don't care staked time?
	if !agent.RewardAmount.IsZero() && rewardDuration == 0 {
		return rewardCoin, errors.New("agent has't reward")
	}

	withdrawHeight, ok := k.GetDelegatorWithdrawRewardHeight(ctx, delegatorAccAddr, agentAccAddr)
	if !ok {
		withdrawHeight = agent.CreatedBlockHeight
	}

	withdrawDuration := curBlockHeight - withdrawHeight
	if withdrawDuration <= 0 {
		return rewardCoin, errors.New("staking duration is too short")
	}

	delegatorRewardAmt := agent.RewardAmount.Mul(delegatorShares).MulRaw(withdrawDuration).Quo(agent.Shares).QuoRaw(rewardDuration)
	if delegatorRewardAmt.IsZero() {
		return rewardCoin, errors.New("delegator has't reward")
	}

	// delegator get staking rewards immediately.
	rewardCoin = sdk.NewCoin(nativeCoinDenom, delegatorRewardAmt)
	rewardCoins := sdk.Coins{rewardCoin}
	if err := k.sendCoinsFromAccountToAccount(ctx, agentAccAddr, delegatorAccAddr, rewardCoins); err != nil {
		return rewardCoin, err
	}

	agent.RewardAmount.Sub(delegatorRewardAmt)
	// only after get some reward, the withdraw height can be reset.
	k.SetDelegatorWithdrawRewardHeight(ctx, delegatorAccAddr, agentAccAddr, curBlockHeight)

	return rewardCoin, nil
}

func (k Keeper) WithdrawReward(ctx sdk.Context, valAddr sdk.ValAddress, stakeDenom string, delegatorAccAddr sdk.AccAddress) (sdk.Coin, error) {
	agent, found := k.GetMTStakingAgent(ctx, stakeDenom, valAddr)
	if !found {
		return sdk.Coin{}, types.ErrNotExistedAgent
	}
	agentAccAddr := sdk.MustAccAddressFromBech32(agent.AgentAddress)

	return k.distributeDelegatorReward(ctx, delegatorAccAddr, agentAccAddr, valAddr, agent)
}
