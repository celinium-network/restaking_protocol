package keeper

import (
	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

type Keeper struct {
	storeKey     storetypes.StoreKey
	cdc          codec.Codec
	paramSpace   paramtypes.Subspace
	scopedKeeper ibcexported.ScopedKeeper

	channelKeeper     restaking.ChannelKeeper
	portKeeper        restaking.PortKeeper
	connectionKeeper  restaking.ConnectionKeeper
	clientKeeper      restaking.ClientKeeper
	ibcTransferKeeper ibctransferkeeper.Keeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	scopedKeeper ibcexported.ScopedKeeper,
	channelKeeper restaking.ChannelKeeper,
	portKeeper restaking.PortKeeper,
	connectionKeeper restaking.ConnectionKeeper,
	clientKeeper restaking.ClientKeeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
) Keeper {
	// set KeyTable if it has not already been set
	// if !paramSpace.HasKeyTable() {
	// 		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	// }

	k := Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		scopedKeeper:      scopedKeeper,
		channelKeeper:     channelKeeper,
		portKeeper:        portKeeper,
		connectionKeeper:  connectionKeeper,
		clientKeeper:      clientKeeper,
		ibcTransferKeeper: ibcTransferKeeper,
	}

	return k
}

func (k Keeper) GetConsumerClientID(ctx sdk.Context, chainID string) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsumerClientIDKey(chainID))

	return bz, bz != nil
}

func (k Keeper) SetConsumerClientID(ctx sdk.Context, chainID, clientID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ConsumerClientIDKey(chainID), []byte(clientID))
}

func (k Keeper) SetConsumerAdditionProposal(ctx sdk.Context, prop *types.ConsumerAdditionProposal) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(prop)
	store.Set(types.ConsumerAdditionProposalKey(prop.ChainId), bz)
}

func (k Keeper) DeleteConsumerAdditionProposal(ctx sdk.Context, chainID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.ConsumerAdditionProposalKey(chainID))
}

func (k Keeper) GetConsumerAdditionProposal(ctx sdk.Context, chainID string) (*types.ConsumerAdditionProposal, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.ConsumerAdditionProposalKey(chainID))
	if bz == nil {
		return nil, false
	}

	var prop types.ConsumerAdditionProposal
	if err := k.cdc.Unmarshal(bz, &prop); err != nil {
		return nil, false
	}

	return &prop, true
}

// GetPort returns the portID for the CCV module. Used in ExportGenesis
func (k Keeper) GetPort(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	return string(store.Get(types.PortKey()))
}

// SetPort sets the portID for the CCV module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Context, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey(), []byte(portID))
}

func (k Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) IsBound(ctx sdk.Context, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// BindPort defines a wrapper function for the ort Keeper's function in
// order to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx sdk.Context, portID string) error {
	cap := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// VerifyConnectingConsumer
func (k Keeper) VerifyConnectingConsumer(ctx sdk.Context, tmClient *ibctm.ClientState) error {
	prop, found := k.GetConsumerAdditionProposal(ctx, tmClient.ChainId)
	if !found {
		return errorsmod.Wrapf(restaking.ErrUnauthorizedConsumerChain,
			"Consumer chain is not authorized ChainID:%s", tmClient.ChainId)
	}

	if err := verifyConsumerAdditionProposal(prop, tmClient); err != nil {
		return err
	}

	return nil
}

// Retrieves the underlying client state corresponding to a connection ID.
func (k Keeper) GetUnderlyingClient(ctx sdk.Context, connectionID string) (
	clientID string, tmClient *ibctm.ClientState, err error,
) {
	conn, ok := k.connectionKeeper.GetConnection(ctx, connectionID)
	if !ok {
		return "", nil, errorsmod.Wrapf(conntypes.ErrConnectionNotFound,
			"connection not found for connection ID: %s", connectionID)
	}
	clientID = conn.ClientId
	clientState, ok := k.clientKeeper.GetClientState(ctx, clientID)
	if !ok {
		return "", nil, errorsmod.Wrapf(clienttypes.ErrClientNotFound,
			"client not found for client ID: %s", conn.ClientId)
	}
	tmClient, ok = clientState.(*ibctm.ClientState)
	if !ok {
		return "", nil, errorsmod.Wrapf(clienttypes.ErrInvalidClientType,
			"invalid client type. expected %s, got %s", ibcexported.Tendermint, clientState.ClientType())
	}
	return clientID, tmClient, nil
}

func (k Keeper) GetUnderlyingConnection(ctx sdk.Context, srcPortID, srcChannelID string) (string, error) {
	channel, found := k.channelKeeper.GetChannel(ctx, srcPortID, srcChannelID)
	if !found {
		return "", errorsmod.Wrapf(channeltypes.ErrChannelNotFound,
			"connection not found for srcPort: %s, srcChannel %s", srcPortID, srcChannelID)
	}
	if len(channel.ConnectionHops) == 0 {
		return "", errorsmod.Wrapf(channeltypes.ErrChannelNotFound,
			"connection not found for srcPort: %s, srcChannel %s", srcPortID, srcChannelID)
	}
	return channel.ConnectionHops[0], nil
}

func (k Keeper) GetConsumerClientIDByChannel(ctx sdk.Context, srcPortID, srcChannelID string) (string, error) {
	connectionID, err := k.GetUnderlyingConnection(ctx, srcPortID, srcChannelID)
	if err != nil {
		return "", err
	}
	clientID, _, err := k.GetUnderlyingClient(ctx, connectionID)
	if err != nil {
		return "", err
	}
	return clientID, nil
}

func (k Keeper) GetConsumerValidator(ctx sdk.Context, clientID string) ([]abci.ValidatorUpdate, bool) {
	var vus types.ConsumerValidatorUpdates

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsumerValidatorSetKey(clientID))
	if bz == nil {
		return nil, false
	}

	k.cdc.MustUnmarshal(bz, &vus)

	return vus.ValidatorUpdates, true
}

func (k Keeper) SetConsumerValidator(ctx sdk.Context, clientID string, vus abci.ValidatorUpdates) {
	vsc := types.ConsumerValidatorUpdates{
		ValidatorUpdates: vus,
	}

	bz := k.cdc.MustMarshal(&vsc)
	store := ctx.KVStore(k.storeKey)

	store.Set(types.ConsumerValidatorSetKey(clientID), bz)
}
