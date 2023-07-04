package keeper

import (
	"fmt"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetAllAgentsByVal(ctx sdk.Context, valAddr sdk.ValAddress) []types.MTStakingAgent {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.AgentPrefix)
	defer iterator.Close()

	valAddrStr := valAddr.String()
	var agents []types.MTStakingAgent
	for ; iterator.Valid(); iterator.Next() {
		var agent types.MTStakingAgent

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
