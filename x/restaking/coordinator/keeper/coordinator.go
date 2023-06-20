package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) OnRecvConsumerVSC(
	ctx sdk.Context,
	consumerClientID string,
	changeList []restaking.ValidatorSetChange,
) ibcexported.Acknowledgement {
	for _, change := range changeList {
		if change.Type == restaking.ValidatorSetChange_Add || change.Type == restaking.ValidatorSetChange_Update {
			for _, v := range change.ValidatorUpdates {
				k.SetConsumerValidator(ctx, consumerClientID, types.ValidatorUpdateToConsumerValidator(v))
			}
		} else if change.Type == restaking.ValidatorSetChange_Remove {
			for _, v := range change.ValidatorUpdates {
				k.DeleteConsumerValidator(ctx, consumerClientID, v.PubKey)
			}
		}
	}

	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
	return ack
}

func (k Keeper) OnRecvConsumerPacketData(
	ctx sdk.Context,
	packet channeltypes.Packet,
	consumerPacket restaking.ConsumerPacketData,
) ibcexported.Acknowledgement {
	consumerClientID, err := k.GetConsumerClientIDByChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		ctx.Logger().Error("Coordinator can't get consumer clientID at receive VSC of consumer")
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return k.OnRecvConsumerVSC(ctx, consumerClientID, consumerPacket.ValidatorSetChanges)
}
