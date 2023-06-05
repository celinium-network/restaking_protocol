package keeper

import (
	"fmt"

	"github.com/celinium-network/restaking_protocol/x/multistaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetExpectedDelegationAmount(ctx sdk.Context, coin sdk.Coin) (sdk.Coin, error) {
	defaultBondDenom := k.stakingkeeper.BondDenom(ctx)

	return k.EquivalentCoinCalculator(ctx, coin, defaultBondDenom)
}

func (k Keeper) GetAllAgentsByVal(ctx sdk.Context, valAddr sdk.ValAddress) []types.MultiStakingAgent {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.MultiStakingAgentPrefix)
	defer iterator.Close()

	valAddrStr := valAddr.String()
	var agents []types.MultiStakingAgent
	for ; iterator.Valid(); iterator.Next() {
		var agent types.MultiStakingAgent

		err := k.cdc.Unmarshal(iterator.Value(), &agent)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("unmarshal has err %s", err))
			continue
		}

		if agent.ValidatorAddress != valAddrStr {
			continue
		}

		agents = append(agents, agent)
	}

	return agents
}
