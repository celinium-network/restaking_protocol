package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

// SlashAgentFromValidator define a method to slash all agent which delegate to the slashed validator.
func (k Keeper) SlashAgentFromValidator(ctx sdk.Context, valAddr sdk.ValAddress, slashFactor sdk.Dec) {
	store := ctx.KVStore(k.storeKey)

	prefix := types.MTStakingAgentAddressPrefix
	prefix = append(prefix, valAddr...)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		agentAccAddr := sdk.AccAddress(iterator.Value())
		agent, found := k.GetMTStakingAgentByAddress(ctx, agentAccAddr)
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("not found agent by address %s", agentAccAddr.String()))
			continue
		}

		slashAmount := sdk.NewDecFromInt(agent.StakedAmount).Mul(slashFactor).TruncateInt()
		unbondingDelegations := k.GetUnbondingDelegationFromAgent(ctx, agentAccAddr)
		slashCoin := sdk.NewCoin(agent.StakeDenom, slashAmount)

		for _, unbondingDelegation := range unbondingDelegations {
			k.SlashUnbondingDelegation(ctx, unbondingDelegation, ctx.BlockHeight(), slashFactor)
		}

		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, agentAccAddr, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("send agent coins to module failed, agentID %s,error: %s", agent.AgentAddress, err))
			continue
		}

		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("burn agent cons failed: %s", err))
			continue
		}

		agent.StakedAmount = agent.StakedAmount.Sub(slashAmount)
		k.SetMTStakingAgent(ctx, agentAccAddr, agent)
	}
}

// slash an unbonding delegation
// Refer to the design of cosmos sdk
func (k Keeper) SlashUnbondingDelegation(ctx sdk.Context, unbondingDelegation types.MTStakingUnbondingDelegation,
	infractionHeight int64, slashFactor sdk.Dec,
) (totalSlashAmount math.Int) {
	now := ctx.BlockHeader().Time
	totalSlashAmount = math.ZeroInt()
	burnedAmount := math.ZeroInt()

	for i, entry := range unbondingDelegation.Entries {
		if entry.CreatedHeight < infractionHeight {
			continue
		}

		if entry.IsMature(now) {
			continue
		}

		slashAmountDec := slashFactor.MulInt(entry.InitialBalance.Amount)
		slashAmount := slashAmountDec.TruncateInt()
		totalSlashAmount = totalSlashAmount.Add(slashAmount)

		unbondingSlashAmount := sdk.MinInt(slashAmount, entry.Balance.Amount)
		if unbondingSlashAmount.IsZero() {
			continue
		}

		burnedAmount = burnedAmount.Add(unbondingSlashAmount)
		entry.Balance = entry.Balance.SubAmount(unbondingSlashAmount)
		unbondingDelegation.Entries[i] = entry
		k.SetMTStakingUnbondingDelegation(ctx, &unbondingDelegation)
	}

	return totalSlashAmount
}

// InstantSlash define a method for slash the delegator of an agent.
// TODO rename such as InstantSlashAgent?
func (k Keeper) InstantSlash(ctx sdk.Context, valAddr sdk.ValAddress, delegator sdk.AccAddress, slashCoin sdk.Coin) error {
	agent, found := k.GetMTStakingAgent(ctx, slashCoin.Denom, valAddr)
	if !found {
		return types.ErrNotExistedAgent
	}

	agentAccAddr, err := sdk.AccAddressFromBech32(agent.AgentAddress)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("agent't delegator is invalid: %s", err))
		return err
	}

	removedShares := agent.Shares.Mul(slashCoin.Amount).Quo(agent.StakedAmount)
	if err = k.DecreaseDelegatorAgentShares(ctx, removedShares, agentAccAddr, delegator); err != nil {
		return err
	}

	agent.StakedAmount = agent.StakedAmount.Sub(slashCoin.Amount)

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, agentAccAddr, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("send agent coins to module failed, agentID %s,error: %s", agent.AgentAddress, err))
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("burn agent cons failed: %s", err))
		return err
	}

	// the delegation of the agent must be slashed immediately.
	if err := k.RefreshAgentDelegation(ctx, agent); err != nil {
		ctx.Logger().Error(fmt.Sprintf("refreshAgentDelegation failed, agentAddress %s", err))
		return err
	}

	k.SetMTStakingAgent(ctx, agentAccAddr, agent)

	return nil
}
