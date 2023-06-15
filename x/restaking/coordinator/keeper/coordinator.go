package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) OnRecvConsumerValSetUpdates(
	ctx sdk.Context,
	packet channeltypes.Packet,
	changes restaking.ValidatorSetChange,
) ibcexported.Acknowledgement {
	consumerClientID, err := k.GetConsumerClientIDByChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		ctx.Logger().Error("Coordinator can't get consumer clientID at receive VSC of consumer")
		return channeltypes.NewErrorAcknowledgement(err)
	}

	curValidatorUpdates, found := k.GetConsumerValidator(ctx, consumerClientID)
	if !found {
		curValidatorUpdates = changes.ValidatorUpdates
	} else {
		curValidatorUpdates = restaking.AccumulateChanges(curValidatorUpdates, changes.ValidatorUpdates)
	}

	// TODO correct process validator set changes.
	k.SetConsumerValidator(ctx, consumerClientID, curValidatorUpdates)

	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	return ack
}
