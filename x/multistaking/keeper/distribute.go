package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/multistaking/types"
)

func (k Keeper) WithdrawRestakingReward(ctx sdk.Context, agentID uint64, delegator string) (sdk.Coin, error) {
	var reward sdk.Coin
	shares := k.GetMultiStakingShares(ctx, agentID, delegator)
	if shares.IsZero() {
		return sdk.Coin{}, types.ErrNoShares
	}

	agent, found := k.GetMultiStakingAgentByID(ctx, agentID)
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
	agentAccAddr := sdk.MustAccAddressFromBech32(agent.DelegateAddress)

	reward.Denom = k.stakingkeeper.BondDenom(ctx)
	reward.Amount = amount
	if err := k.sendCoinsFromAccountToAccount(
		ctx, agentAccAddr, delegatorAccAddr, sdk.Coins{reward},
	); err != nil {
		return sdk.Coin{}, err
	}
	return reward, nil
}
