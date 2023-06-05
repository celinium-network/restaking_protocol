package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

type Keeper struct {
	storeKey     storetypes.StoreKey
	cdc          codec.Codec
	scopedKeeper exported.ScopedKeeper

	channelKeeper     restaking.ChannelKeeper
	portKeeper        restaking.PortKeeper
	connectionKeeper  restaking.ConnectionKeeper
	clientKeeper      restaking.ClientKeeper
	ibcTransferKeeper ibctransferkeeper.Keeper

	standaloneStakingKeeper restaking.StakingKeeper
	slashingKeeper          restaking.SlashingKeeper
	bankKeeper              restaking.BankKeeper
	authKeeper              restaking.AccountKeeper
}

func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.Codec,
	scopedKeeper exported.ScopedKeeper,
	channelKeeper restaking.ChannelKeeper,
	portKeeper restaking.PortKeeper,
	connectionKeeper restaking.ConnectionKeeper,
	clientKeeper restaking.ClientKeeper,
	ibcTransferKeeper ibctransferkeeper.Keeper,
	standaloneStakingKeeper restaking.StakingKeeper,
	slashingKeeper restaking.SlashingKeeper,
	bankKeeper restaking.BankKeeper,
	authKeeper restaking.AccountKeeper,
) Keeper {
	k := Keeper{
		storeKey:                storeKey,
		cdc:                     cdc,
		scopedKeeper:            scopedKeeper,
		channelKeeper:           channelKeeper,
		portKeeper:              portKeeper,
		connectionKeeper:        connectionKeeper,
		clientKeeper:            clientKeeper,
		ibcTransferKeeper:       ibcTransferKeeper,
		standaloneStakingKeeper: standaloneStakingKeeper,
		slashingKeeper:          slashingKeeper,
		bankKeeper:              bankKeeper,
		authKeeper:              authKeeper,
	}
	return k
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
