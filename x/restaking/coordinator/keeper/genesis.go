package keeper

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restakingtypes "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) []abci.ValidatorUpdate {
	k.SetPort(ctx, restakingtypes.CoordinatorPortID)

	if !k.IsBound(ctx, restakingtypes.CoordinatorPortID) {
		// transfer module binds to the transfer port on InitChain
		// and claims the returned capability
		err := k.BindPort(ctx, restakingtypes.CoordinatorPortID)
		if err != nil {
			// If the binding fails, the chain MUST NOT start
			panic(fmt.Sprintf("could not claim port capability: %v", err))
		}
	}

	return nil
}
