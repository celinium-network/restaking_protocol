package keeper

import (
	"encoding/binary"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

type Keeper struct {
	storeKey     storetypes.StoreKey
	cdc          codec.Codec
	scopedKeeper restaking.ScopedKeeper

	channelKeeper     restaking.ChannelKeeper
	portKeeper        restaking.PortKeeper
	connectionKeeper  restaking.ConnectionKeeper
	clientKeeper      restaking.ClientKeeper
	ibcTransferKeeper restaking.IBCTransferKeeper

	stakingKeeper  restaking.StakingKeeper
	slashingKeeper restaking.SlashingKeeper
	bankKeeper     restaking.BankKeeper
	authKeeper     restaking.AccountKeeper

	multiStakingKeeper restaking.MultiStakingKeeper
}

func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.Codec,
	scopedKeeper restaking.ScopedKeeper,
	channelKeeper restaking.ChannelKeeper,
	portKeeper restaking.PortKeeper,
	connectionKeeper restaking.ConnectionKeeper,
	clientKeeper restaking.ClientKeeper,
	ibcTransferKeeper restaking.IBCTransferKeeper,
	bankKeeper restaking.BankKeeper,
	standaloneStakingKeeper restaking.StakingKeeper,
	slashingKeeper restaking.SlashingKeeper,
	authKeeper restaking.AccountKeeper,
	multiStakingKeeper restaking.MultiStakingKeeper,
) Keeper {
	k := Keeper{
		storeKey:           storeKey,
		cdc:                cdc,
		scopedKeeper:       scopedKeeper,
		channelKeeper:      channelKeeper,
		portKeeper:         portKeeper,
		connectionKeeper:   connectionKeeper,
		clientKeeper:       clientKeeper,
		ibcTransferKeeper:  ibcTransferKeeper,
		stakingKeeper:      standaloneStakingKeeper,
		slashingKeeper:     slashingKeeper,
		bankKeeper:         bankKeeper,
		authKeeper:         authKeeper,
		multiStakingKeeper: multiStakingKeeper,
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

func (k Keeper) GetAllValidators(ctx sdk.Context) []stakingtypes.Validator {
	return k.stakingKeeper.GetLastValidators(ctx)
}

func (k Keeper) GetInitialValidator(ctx sdk.Context) abci.ValidatorUpdates {
	var valUpdates abci.ValidatorUpdates

	for _, v := range k.stakingKeeper.GetValidatorUpdates(ctx) {
		valUpdates = append(valUpdates, abci.ValidatorUpdate{
			PubKey: v.PubKey,
			Power:  v.Power,
		})
	}

	return valUpdates
}

func (k Keeper) GetValidatorSetUpdateID(ctx sdk.Context) (validatorSetUpdateID uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorSetUpdateIDKey())

	if bz == nil {
		validatorSetUpdateID = 0
	} else {
		validatorSetUpdateID = binary.BigEndian.Uint64(bz)
	}

	return validatorSetUpdateID
}

func (k Keeper) SetValidatorSetUpdateID(ctx sdk.Context, valUpdateID uint64) {
	store := ctx.KVStore(k.storeKey)

	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, valUpdateID)

	store.Set(types.GetValidatorSetUpdateIDKey(), bz)
}

func (k Keeper) GetPendingVSCPackets(ctx sdk.Context) []restaking.ValidatorSetChange {
	var packets restaking.ValidatorSetChanges

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPendingValidatorChangeSetKey())
	if bz == nil {
		return []restaking.ValidatorSetChange{}
	}
	if err := packets.Unmarshal(bz); err != nil {
		// An error here would indicate something is very wrong,
		// the PendingVSCPackets are assumed to be correctly serialized in AppendPendingVSCPackets.
		panic(fmt.Errorf("cannot unmarshal pending validator set changes: %w", err))
	}
	return packets.ValidatorSetChanges
}

func (k Keeper) DeletePendingVSCPackets(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPendingValidatorChangeSetKey())
}

func (k Keeper) AppendPendingVSCPackets(ctx sdk.Context, addedPackets ...restaking.ValidatorSetChange) {
	packets := append(k.GetPendingVSCPackets(ctx), addedPackets...)

	store := ctx.KVStore(k.storeKey)
	newPackets := restaking.ValidatorSetChanges{ValidatorSetChanges: packets}
	buf, err := newPackets.Marshal()
	if err != nil {
		// An error here would indicate something is very wrong,
		// packets is instantiated in this method and should be able to be marshaled.
		panic(fmt.Errorf("cannot marshal pending validator set changes: %w", err))
	}
	store.Set(types.GetPendingValidatorChangeSetKey(), buf)
}

func (k Keeper) SetCoordinatorChannelID(ctx sdk.Context, channelID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetCoordinatorChannelIDKey(), []byte(channelID))
}

func (k Keeper) GetCoordinatorChannelID(ctx sdk.Context) (string, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCoordinatorChannelIDKey())
	if bz == nil {
		return "", types.ErrRestakingChannelNotFound
	}
	return string(bz), nil
}

func (k Keeper) GetOperatorLocalAddress(ctx sdk.Context, operatorAddress string, validatorPk []byte) (addr sdk.AccAddress, found bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.OperatorAddressKey(validatorPk, operatorAddress))
	if bz == nil {
		return addr, false
	}
	addr, err := sdk.AccAddressFromBech32(string(bz))
	if err != nil {
		return addr, false
	}

	return addr, true
}

func (k Keeper) GetOrCreateOperatorLocalAddress(
	ctx sdk.Context,
	srcChannel, srcPort, operatorAddress string,
	validatorPk []byte,
) sdk.AccAddress {
	operatorLocalAddress, found := k.GetOperatorLocalAddress(ctx, operatorAddress, validatorPk)
	if !found {
		operatorLocalAccount := k.GenerateOperatorAccount(
			ctx,
			srcChannel,
			srcPort,
			operatorAddress,
			validatorPk,
		)

		operatorLocalAddress = operatorLocalAccount.GetAddress()
	}
	return operatorLocalAddress
}

func (k Keeper) SetOperatorLocalAddress(ctx sdk.Context, operatorAddress string, validatorPk []byte, localAddress sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.OperatorAddressKey(validatorPk, operatorAddress), []byte(localAddress.String()))
}
