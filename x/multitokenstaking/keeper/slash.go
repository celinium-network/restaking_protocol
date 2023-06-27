package keeper

import (
	"fmt"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO more detailed research about slash
func (k Keeper) SlashAgentByValidatorSlash(ctx sdk.Context, valAddr sdk.ValAddress, slashFactor sdk.Dec) {
	agents := k.GetAllAgentsByVal(ctx, valAddr)

	slashDenom := k.stakingkeeper.GetParams(ctx).BondDenom
	for i := 0; i < len(agents); i++ {
		slashAmt := sdk.NewDecFromInt(agents[i].StakedAmount).Mul(slashFactor)
		slashCoin := sdk.NewCoin(slashDenom, slashAmt.TruncateInt())

		agentDelegatorAddr, err := sdk.AccAddressFromBech32(agents[i].AgentAddress)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("agent't delegator is invalid: %s", err))
			continue
		}

		if err := k.bankKeeper.SendCoinsFromAccountToModule(
			ctx, agentDelegatorAddr, types.ModuleName, sdk.Coins{slashCoin},
		); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("send agent coins to module failed, agentID %s,error: %s", agents[i].AgentAddress, err))
			continue
		}

		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("burn agent cons failed: %s", err))
			continue
		}

		agents[i].StakedAmount = agents[i].StakedAmount.Sub(slashCoin.Amount)
		k.SetMTStakingAgent(ctx, &agents[i])
	}
}

func (k Keeper) SlashDelegator(ctx sdk.Context, valAddr sdk.ValAddress, delegator sdk.AccAddress, slashCoin sdk.Coin) error {
	agent, found := k.GetMTStakingAgent(ctx, slashCoin.Denom, valAddr.String())
	if !found {
		return types.ErrNotExistedAgent
	}

	removedShares := agent.Shares.Mul(slashCoin.Amount).Quo(agent.StakedAmount)
	err := k.DecreaseMTStakingShares(ctx, removedShares, agent.AgentAddress, delegator.String())
	if err != nil {
		return err
	}
	agent.StakedAmount = agent.StakedAmount.Sub(slashCoin.Amount)

	agentDelegatorAddr, err := sdk.AccAddressFromBech32(agent.AgentAddress)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("agent't delegator is invalid: %s", err))
		return err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, agentDelegatorAddr, types.ModuleName, sdk.Coins{slashCoin},
	); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("send agent coins to module failed, agentID %s,error: %s", agent.AgentAddress, err))
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.Coins{slashCoin}); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("burn agent cons failed: %s", err))
		return err
	}

	k.SetMTStakingAgent(ctx, agent)

	return nil
}
