package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

func (k Keeper) WithdrawRestakingReward(ctx sdk.Context, agentAddress string, delegator string) (sdk.Coin, error) {
	var reward sdk.Coin
	shares := k.GetDelegatorAgentShares(ctx, agentAddress, delegator)
	if shares.IsZero() {
		return sdk.Coin{}, types.ErrNoShares
	}

	agent, found := k.GetMTStakingAgentByAddress(ctx, agentAddress)
	if !found {
		return sdk.Coin{}, types.ErrNotExistedAgent
	}

	if agent.RewardAmount.IsZero() {
		return sdk.Coin{}, nil
	}

	amount := agent.RewardAmount.Mul(shares).Quo(agent.Shares)
	if amount.IsZero() {
		return sdk.Coin{}, nil
	}
	delegatorAccAddr := sdk.MustAccAddressFromBech32(delegator)
	agentAccAddr := sdk.MustAccAddressFromBech32(agent.AgentAddress)

	reward.Denom = k.stakingKeeper.BondDenom(ctx)
	reward.Amount = amount
	if err := k.sendCoinsFromAccountToAccount(
		ctx, agentAccAddr, delegatorAccAddr, sdk.Coins{reward},
	); err != nil {
		return sdk.Coin{}, err
	}
	return reward, nil
}
