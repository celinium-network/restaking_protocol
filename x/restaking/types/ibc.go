package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
)

func SendIBCPacket(
	ctx sdk.Context,
	scopedKeeper ScopedKeeper,
	channelKeeper ChannelKeeper,
	channelID string,
	portID string,
	packetData []byte,
	timeoutPeriod time.Duration,
) (uint64, error) {
	channelCap, ok := scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !ok {
		return 0, errorsmod.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	timeoutTimestamp := uint64(ctx.BlockTime().Add(timeoutPeriod).UnixNano())
	return channelKeeper.SendPacket(ctx, channelCap, portID, channelID, clienttypes.Height{}, timeoutTimestamp, packetData)
}
