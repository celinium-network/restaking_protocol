package keeper

import (
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) EndBlocker(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	k.ProcessCompletedUnbonding(ctx)
	return nil, nil
}
