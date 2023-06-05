package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) HandleConsumerAdditionProposal(ctx sdk.Context, proposal *types.ConsumerAdditionProposal) error {
	chainID := proposal.ChainId

	if _, found := k.GetConsumerClientID(ctx, proposal.ChainId); found {
		return errorsmod.Wrap(restaking.ErrDuplicateConsumerChain,
			fmt.Sprintf("cannot create client for existent consumer chain: %s", chainID))
	}

	k.SetConsumerAdditionProposal(ctx, proposal)

	return nil
}

func verifyConsumerAdditionProposal(proposal *types.ConsumerAdditionProposal, client *ibctm.ClientState) error {
	return nil
}

func (k Keeper) GetConsumerAdditionPropsToExecute(ctx sdk.Context) (propsToExecute []types.ConsumerAdditionProposal) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{types.ConsumerAdditionProposalPrefix})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var prop types.ConsumerAdditionProposal
		k.cdc.MustUnmarshal(iterator.Value(), &prop)
		propsToExecute = append(propsToExecute, prop)
	}

	return propsToExecute
}
