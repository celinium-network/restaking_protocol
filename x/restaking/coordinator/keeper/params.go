package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

// GetTemplateClient returns the template client for provider proposals
func (k Keeper) GetTemplateClient(ctx sdk.Context) *ibctm.ClientState {
	var cs ibctm.ClientState
	k.paramSpace.Get(ctx, types.KeyTemplateClient, &cs)
	return &cs
}
