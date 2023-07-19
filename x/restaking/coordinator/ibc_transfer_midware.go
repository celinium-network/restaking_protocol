package coordinator

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/keeper"
)

var _ porttypes.Middleware = IBCTransferMiddleware{}

// Middleware just warp IBCTransferModule
type IBCTransferMiddleware struct {
	app    porttypes.IBCModule
	keeper keeper.Keeper
}

// OnAcknowledgementPacket implements types.Middleware
func (im IBCTransferMiddleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnChanCloseConfirm implements types.Middleware
func (im IBCTransferMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements types.Middleware
func (im IBCTransferMiddleware) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanOpenAck implements types.Middleware
func (im IBCTransferMiddleware) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements types.Middleware
func (im IBCTransferMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanOpenInit implements types.Middleware
func (im IBCTransferMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry implements types.Middleware
func (im IBCTransferMiddleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnRecvPacket implements types.Middleware
func (im IBCTransferMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	if err := im.keeper.OnRecvIBCTransferPacket(ctx, packet); err != nil {
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("cannot unmarshal CCV packet data"))
	}
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnTimeoutPacket implements types.Middleware
func (im IBCTransferMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// GetAppVersion implements types.Middleware
func (im IBCTransferMiddleware) GetAppVersion(ctx sdk.Context, portID string, channelID string) (string, bool) {
	panic("unimplemented")
}

// SendPacket implements types.Middleware
func (im IBCTransferMiddleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (sequence uint64, err error) {
	panic("unimplemented")
}

// WriteAcknowledgement implements types.Middleware
func (IBCTransferMiddleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	panic("unimplemented")
}

func NewIBCMiddleware(k keeper.Keeper, app porttypes.IBCModule) IBCTransferMiddleware {
	return IBCTransferMiddleware{
		app:    app,
		keeper: k,
	}
}
