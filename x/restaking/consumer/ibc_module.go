package consumer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

var _ porttypes.IBCModule = AppModule{}

// OnChanOpenInit implements types.IBCModule
func (am AppModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	err := am.keeper.ClaimCapability(ctx, channelCap, host.ChannelCapabilityPath(portID, channelID))
	if err != nil {
		return "", err
	}

	return restaking.Version, nil
}

// OnChanOpenAck implements types.IBCModule
func (am AppModule) OnChanOpenAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	am.keeper.SetCoordinatorChannelID(ctx, channelID)
	return nil
}

// OnChanOpenConfirm implements types.IBCModule
func (am AppModule) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	return nil
}

// OnAcknowledgementPacket implements types.IBCModule
func (AppModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return nil
}

// OnChanCloseConfirm implements types.IBCModule
func (AppModule) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return nil
}

// OnChanCloseInit implements types.IBCModule
func (AppModule) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	return nil
}

// OnChanOpenTry implements types.IBCModule
func (AppModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	return restaking.Version, nil
}

// OnRecvPacket implements types.IBCModule
func (am AppModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	var restakingProtocolPacket restaking.RestakingPacket
	err := am.cdc.Unmarshal(packet.Data, &restakingProtocolPacket)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return am.keeper.OnRecvPacket(ctx, packet, &restakingProtocolPacket)
}

// OnTimeoutPacket implements types.IBCModule
func (AppModule) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return nil
}
