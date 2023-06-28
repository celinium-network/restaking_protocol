package keeper

import (
	"fmt"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SlashDelegatingAgentsToValidator define a method to slash all agent which delegate to the slashed validator.
func (k Keeper) SlashDelegatingAgentsToValidator(ctx sdk.Context, valAddr sdk.ValAddress, slashFactor sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.AgentPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var agent types.MTStakingAgent
		err := k.cdc.Unmarshal(iterator.Value(), &agent)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("unmarshal has err %s", err))
			continue
		}

		slashAmount := sdk.NewDecFromInt(agent.StakedAmount).Mul(slashFactor).TruncateInt()
		slashCoin := sdk.NewCoin(agent.StakeDenom, slashAmount)
		agentAccAddr, err := sdk.AccAddressFromBech32(agent.AgentAddress)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("agent't delegator is invalid: %s", err))
			continue
		}

		if err := k.bankKeeper.SendCoinsFromAccountToModule(
			ctx, agentAccAddr, types.ModuleName, sdk.Coins{slashCoin},
		); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("send agent coins to module failed, agentID %s,error: %s", agent.AgentAddress, err))
			continue
		}

		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("burn agent cons failed: %s", err))
			continue
		}

		agent.StakedAmount = agent.StakedAmount.Sub(slashAmount)

		bz, err := k.cdc.Marshal(&agent)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("marshal agent failed: %v", agent))
			continue
		}
		store.Set(iterator.Key(), bz)
	}
}

// SlashDelegator define a method for slash the delegator of an agent.
func (k Keeper) SlashDelegator(ctx sdk.Context, valAddr sdk.ValAddress, delegator sdk.AccAddress, slashCoin sdk.Coin) error {
	agent, found := k.GetMTStakingAgent(ctx, slashCoin.Denom, valAddr.String())
	if !found {
		return types.ErrNotExistedAgent
	}

	removedShares := agent.Shares.Mul(slashCoin.Amount).Quo(agent.StakedAmount)
	err := k.DecreaseDelegatorAgentShares(ctx, removedShares, agent.AgentAddress, delegator.String())
	if err != nil {
		return err
	}
	agent.StakedAmount = agent.StakedAmount.Sub(slashCoin.Amount)

	agentAccAddr, err := sdk.AccAddressFromBech32(agent.AgentAddress)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("agent't delegator is invalid: %s", err))
		return err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, agentAccAddr, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("send agent coins to module failed, agentID %s,error: %s", agent.AgentAddress, err))
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("burn agent cons failed: %s", err))
		return err
	}

	// the delegation of the agent must be slashed immediately.
	if err := k.refreshAgentDelegation(ctx, agent); err != nil {
		ctx.Logger().Error(fmt.Sprintf("refreshAgentDelegation failed, agentAddress %s", err))
		return err
	}

	k.SetMTStakingAgent(ctx, agent)

	return nil
}
