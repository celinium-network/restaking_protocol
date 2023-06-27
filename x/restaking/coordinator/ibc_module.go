package coordinator

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/keeper"
	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

var _ porttypes.IBCModule = AppModule{}

// OnChanOpenInit implements types.IBCModule
func (AppModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return version, errorsmod.Wrap(restaking.ErrInvalidChannelFlow, "channel handshake must be initiated by consumer chain")
}

// OnChanOpenTry implements types.IBCModule
func (am AppModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	// Validate parameters
	if err := validateRestakingChannelParams(
		ctx, &am.keeper, order, portID,
	); err != nil {
		return "", err
	}

	if counterparty.PortId != restaking.ConsumerPortID {
		return "", errorsmod.Wrapf(porttypes.ErrInvalidPort,
			"invalid counterparty port: %s, expected %s", counterparty.PortId, restaking.ConsumerPortID)
	}

	if counterpartyVersion != restaking.Version {
		return "", errorsmod.Wrapf(
			restaking.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s",
			counterpartyVersion, restaking.Version)
	}

	if len(connectionHops) != 1 {
		return "", errorsmod.Wrap(channeltypes.ErrTooManyConnectionHops, "must have direct connection to provider chain")
	}

	// Claim channel capability
	if err := am.keeper.ClaimCapability(
		ctx, channelCap, host.ChannelCapabilityPath(portID, channelID),
	); err != nil {
		return "", err
	}

	connectionID := connectionHops[0]
	_, tmClient, err := am.keeper.GetUnderlyingClient(ctx, connectionID)
	if err != nil {
		return "", err
	}

	if err := am.keeper.VerifyConnectingConsumer(ctx, tmClient); err != nil {
		return "", err
	}

	return "", nil
}

// OnChanOpenAck implements types.IBCModule
func (AppModule) OnChanOpenAck(
	ctx sdk.Context,
	portID string,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return errorsmod.Wrap(restaking.ErrInvalidChannelFlow, "channel handshake must be initiated by consumer chain")
}

// OnChanOpenConfirm implements types.IBCModule
func (am AppModule) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	conn, err := am.keeper.GetUnderlyingConnection(ctx, portID, channelID)
	if err != nil {
		return err
	}

	clientID, tmClient, err := am.keeper.GetUnderlyingClient(ctx, conn)
	if err != nil {
		return err
	}

	proposal, found := am.keeper.GetConsumerAdditionProposal(ctx, tmClient.ChainId)
	if !found {
		return types.ErrAdditionalProposalNotFound
	}

	am.keeper.SetConsumerRestakingToken(ctx, clientID, proposal.RestakingTokens)
	am.keeper.SetConsumerRewardToken(ctx, clientID, proposal.RewardTokens) // TODO maybe not useful
	am.keeper.SetConsumerClientID(ctx, tmClient.ChainId, clientID)
	am.keeper.SetConsumerClientIDToChannel(ctx, clientID, channelID)

	am.keeper.DeleteConsumerAdditionProposal(ctx, tmClient.ChainId)

	return nil
}

// OnAcknowledgementPacket implements types.IBCModule
func (am AppModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return am.keeper.HandleIBCAcknowledgement(ctx, &packet, acknowledgement)
}

// OnChanCloseConfirm implements types.IBCModule
func (AppModule) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnChanCloseInit implements types.IBCModule
func (AppModule) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	panic("unimplemented")
}

// OnRecvPacket implements types.IBCModule
func (am AppModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	var (
		ack                exported.Acknowledgement
		consumerPacketData restaking.ConsumerPacket
	)

	if err := consumerPacketData.Unmarshal(packet.Data); err != nil {
		errAck := channeltypes.NewErrorAcknowledgement(fmt.Errorf("cannot unmarshal CCV packet data"))
		ack = &errAck
	} else {
		ack = am.keeper.OnRecvConsumerPacketData(ctx, packet, consumerPacketData)
	}

	return ack
}

// OnTimeoutPacket implements types.IBCModule
func (AppModule) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	panic("unimplemented")
}

// validateRestakingChannelParams validates a ccv channel
func validateRestakingChannelParams(
	ctx sdk.Context,
	keeper *keeper.Keeper,
	order channeltypes.Order,
	portID string,
) error {
	if order != channeltypes.ORDERED {
		return errorsmod.Wrapf(channeltypes.ErrInvalidChannelOrdering,
			"expected %s channel, got %s ", channeltypes.ORDERED, order)
	}

	boundPort := keeper.GetPort(ctx)
	if boundPort != portID {
		return errorsmod.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}
	return nil
}
