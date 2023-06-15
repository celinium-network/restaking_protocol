package keeper

import (
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

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
	scopedKeeper restaking.ScopedKeeper
	bankKeeper   restaking.BankKeeper

	channelKeeper     restaking.ChannelKeeper
	portKeeper        restaking.PortKeeper
	connectionKeeper  restaking.ConnectionKeeper
	clientKeeper      restaking.ClientKeeper
	ibcTransferKeeper restaking.IBCTransferKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	scopedKeeper restaking.ScopedKeeper,
	bankKeeper restaking.BankKeeper,
	channelKeeper restaking.ChannelKeeper,
	portKeeper restaking.PortKeeper,
	connectionKeeper restaking.ConnectionKeeper,
	clientKeeper restaking.ClientKeeper,
	ibcTransferKeeper restaking.IBCTransferKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	// if !paramSpace.HasKeyTable() {
	// 		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	// }

	k := Keeper{
		storeKey:          storeKey,
		cdc:               cdc,
		scopedKeeper:      scopedKeeper,
		bankKeeper:        bankKeeper,
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

func (k Keeper) SetConsumerClientIDToChannel(ctx sdk.Context, clientID, channelID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ConsumerClientIDKey(clientID), []byte(channelID))
}

func (k Keeper) GetConsumerClientIDToChannel(ctx sdk.Context, clientID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsumerClientIDKey(clientID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
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

func (k Keeper) SetConsumerRestakingToken(ctx sdk.Context, clientID string, tokens []string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ConsumerRestakingTokensKey(clientID), []byte(strings.Join(tokens, types.StringListSplitter)))
}

func (k Keeper) GetConsumerRestakingToken(ctx sdk.Context, clientID string) []string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsumerRestakingTokensKey(clientID))

	return strings.Split(string(bz), types.StringListSplitter)
}

func (k Keeper) SetConsumerRewardToken(ctx sdk.Context, clientID string, tokens []string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ConsumerRewardTokensKey(clientID), []byte(strings.Join(tokens, types.StringListSplitter)))
}

func (k Keeper) GetConsumerRewardToken(ctx sdk.Context, clientID string) []string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsumerRewardTokensKey(clientID))

	return strings.Split(string(bz), types.StringListSplitter)
}

func (k Keeper) SetOperator(ctx sdk.Context, operator *types.Operator) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(operator)
	store.Set(types.OperatorKey(operator.OperatorAddress), bz)
}

func (k Keeper) GetOperator(ctx sdk.Context, addr string) (*types.Operator, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.OperatorKey(addr))
	if bz == nil {
		return nil, false
	}
	var operator types.Operator
	k.cdc.MustUnmarshal(bz, &operator)

	return &operator, true
}

func (k Keeper) GetAllOperators(ctx sdk.Context) (operators []types.Operator) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{types.OperatorPrefix})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var op types.Operator
		k.cdc.MustUnmarshal(bz, &op)
		operators = append(operators, op)
	}

	return operators
}

// TODO convert blockHeight to epoch?
func (k Keeper) GetOperatorDelegateRecord(ctx sdk.Context, blockHeight uint64, operatorAddr string) (*types.OperatorDelegationRecord, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.DelegationRecordKey(blockHeight, operatorAddr))
	if bz == nil {
		return nil, false
	}
	var record types.OperatorDelegationRecord
	k.cdc.Unmarshal(bz, &record)

	return &record, true
}

func (k Keeper) GetOperatorDelegateRecordByKey(ctx sdk.Context, key string) (*types.OperatorDelegationRecord, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get([]byte(key))
	if bz == nil {
		return nil, false
	}
	var record types.OperatorDelegationRecord
	k.cdc.Unmarshal(bz, &record)

	return &record, true
}

func (k Keeper) SetOperatorDelegateRecord(ctx sdk.Context, blockHeight uint64, record *types.OperatorDelegationRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(record)
	store.Set(types.DelegationRecordKey(blockHeight, record.OperatorAddress), bz)
}

func (k Keeper) SetOperatorDelegateRecordByKey(ctx sdk.Context, key string, record *types.OperatorDelegationRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(record)
	store.Set([]byte(key), bz)
}

func (k Keeper) DeleteOperatorDelegateRecordByKey(ctx sdk.Context, key string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete([]byte(key))
}

func (k Keeper) GetOperatorUndelegationRecord(ctx sdk.Context, blockHeight uint64, operatorAddr string) (*types.OperatorUndelegationRecord, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.UndelegationRecordKey(blockHeight, operatorAddr))
	if bz == nil {
		return nil, false
	}
	var record types.OperatorUndelegationRecord
	k.cdc.Unmarshal(bz, &record)

	return &record, true
}

func (k Keeper) GetOperatorUndelegationRecordByKey(ctx sdk.Context, key string) (*types.OperatorUndelegationRecord, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get([]byte(key))
	if bz == nil {
		return nil, false
	}
	var record types.OperatorUndelegationRecord
	k.cdc.Unmarshal(bz, &record)

	return &record, true
}

func (k Keeper) SetOperatorUndelegationRecord(ctx sdk.Context, blockHeight uint64, record *types.OperatorUndelegationRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(record)
	store.Set(types.UndelegationRecordKey(blockHeight, record.OperatorAddress), bz)
}

func (k Keeper) SetOperatorUndelegationRecordByKey(ctx sdk.Context, key string, record *types.OperatorUndelegationRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(record)
	store.Set([]byte(key), bz)
}

func (k Keeper) SetDelegation(ctx sdk.Context, ownerAddr, operatorAddr string, delegation *types.Delegation) {
	store := ctx.KVStore(k.storeKey)
	if delegation.Shares.IsZero() {
		store.Delete(types.OperatorSharesKey(ownerAddr, operatorAddr))
		return
	}

	bz := k.cdc.MustMarshal(delegation)

	store.Set(types.OperatorSharesKey(ownerAddr, operatorAddr), bz)
}

func (k Keeper) GetDelegation(ctx sdk.Context, ownerAddr, operatorAddr string) (*types.Delegation, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OperatorSharesKey(ownerAddr, operatorAddr))
	if bz == nil {
		return nil, false
	}

	var delegation types.Delegation
	k.cdc.MustUnmarshal(bz, &delegation)
	return &delegation, true
}

func (k Keeper) sendCoinsFromAccountToAccount(
	ctx sdk.Context,
	senderAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	amt sdk.Coins,
) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, amt); err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiverAddr, amt)
}

func (k Keeper) SetCallback(ctx sdk.Context, channelID, portID string, seq uint64, callback types.IBCCallback) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&callback)
	store.Set(types.IBCCallbackKey(channelID, portID, seq), bz)
}

func (k Keeper) GetCallback(ctx sdk.Context, channelID, portID string, seq uint64) (*types.IBCCallback, bool) {
	var callback types.IBCCallback
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.IBCCallbackKey(channelID, portID, seq))
	if bz == nil {
		return nil, false
	}
	err := k.cdc.Unmarshal(bz, &callback)
	if err != nil {
		return nil, false
	}
	return &callback, true
}

func (k Keeper) InsertUBDQueue(ctx sdk.Context, ubd types.UnbondingDelegation, completionTime time.Time) {
	dvPair := types.DOPair{Delegator: ubd.DelegatorAddress, Operator: ubd.OperatorAddress}

	timeSlice := k.GetUBDQueueTimeSlice(ctx, completionTime)
	if len(timeSlice) == 0 {
		k.SetUBDQueueTimeSlice(ctx, completionTime, []types.DOPair{dvPair})
	} else {
		timeSlice = append(timeSlice, dvPair)
		k.SetUBDQueueTimeSlice(ctx, completionTime, timeSlice)
	}
}

func (k Keeper) GetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (dvPairs []types.DOPair) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetUnbondingDelegationTimeKey(timestamp))
	if bz == nil {
		return []types.DOPair{}
	}

	pairs := types.DOPairs{}
	k.cdc.MustUnmarshal(bz, &pairs)

	return pairs.Pairs
}

// SetUBDQueueTimeSlice sets a specific unbonding queue timeslice.
func (k Keeper) SetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []types.DOPair) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.DOPairs{Pairs: keys})
	store.Set(types.GetUnbondingDelegationTimeKey(timestamp), bz)
}
