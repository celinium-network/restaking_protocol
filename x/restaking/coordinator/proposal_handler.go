package coordinator

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/keeper"
	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func NewProviderProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.ConsumerAdditionProposal:
			return k.HandleConsumerAdditionProposal(ctx, c)
		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ccv proposal content type: %T", c)
		}
	}
}
