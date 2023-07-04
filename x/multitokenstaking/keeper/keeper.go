package keeper

import (
	stdmath "math"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec

	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	epochKeeper        types.EpochKeeper
	stakingKeeper      types.StakingKeeper
	distributionKeeper types.DistributionKeeper

	EquivalentNativeCoinMultiplier EquivalentNativeCoinMultiplier
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	epochKeeper types.EpochKeeper,
	stakingKeeper types.StakingKeeper,
	distributionKeeper types.DistributionKeeper,
) Keeper {
	return Keeper{
		storeKey:                       storeKey,
		cdc:                            cdc,
		accountKeeper:                  accountKeeper,
		bankKeeper:                     bankKeeper,
		epochKeeper:                    epochKeeper,
		stakingKeeper:                  stakingKeeper,
		distributionKeeper:             distributionKeeper,
		EquivalentNativeCoinMultiplier: defaultEquivalentCoinMultiplier,
	}
}

// TODO Temporarily use this method to feed prices !!!
type EquivalentNativeCoinMultiplier func(ctx sdk.Context, denom string) (sdk.Dec, error)

func defaultEquivalentCoinMultiplier(ctx sdk.Context, denom string) (sdk.Dec, error) {
	return sdk.OneDec(), nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) SetEquivalentNativeCoinMultiplier(ctx sdk.Context, epoch int64, denom string, multiplier sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	record := types.EquivalentMultiplierRecord{
		EpochNumber: epoch,
		Denom:       denom,
		Multiplier:  multiplier,
	}
	bz := k.cdc.MustMarshal(&record)

	store.Set(types.GetMTTokenMultiplierKey(denom), bz)
}

func (k Keeper) GetEquivalentNativeCoinMultiplier(ctx sdk.Context, denom string) (multiplier sdk.Dec, found bool) {
	var record types.EquivalentMultiplierRecord
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMTTokenMultiplierKey(denom))
	if bz == nil {
		return multiplier, false
	}
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return multiplier, false
	}

	multiplier = record.Multiplier

	return multiplier, true
}

func (k Keeper) CalculateEquivalentNativeCoin(ctx sdk.Context, coin sdk.Coin) (targetCoin sdk.Coin, err error) {
	multiplier, found := k.GetEquivalentNativeCoinMultiplier(ctx, coin.Denom)
	if !found {
		return targetCoin, types.ErrNoCoinMultiplierFound
	}

	targetCoin.Denom = k.stakingKeeper.BondDenom(ctx)
	targetCoin.Amount = multiplier.MulInt(coin.Amount).TruncateInt()

	return targetCoin, nil
}

func (k Keeper) GetMTStakingAgentByAddress(ctx sdk.Context, agentAccAddr sdk.AccAddress) (*types.MTStakingAgent, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMTStakingAgentKey(agentAccAddr))

	if bz == nil {
		return nil, false
	}

	agent := &types.MTStakingAgent{}
	k.cdc.MustUnmarshal(bz, agent)

	return agent, true
}

func (k Keeper) GetMTStakingAgent(ctx sdk.Context, denom string, valAddr sdk.ValAddress) (*types.MTStakingAgent, bool) {
	agentID, found := k.GetMTStakingAgentAddressByDenomAndVal(ctx, denom, valAddr)
	if !found {
		return nil, false
	}

	return k.GetMTStakingAgentByAddress(ctx, agentID)
}

func (k Keeper) SetMTStakingAgent(ctx sdk.Context, agentAccAddr sdk.AccAddress, agent *types.MTStakingAgent) {
	bz := k.cdc.MustMarshal(agent)
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetMTStakingAgentKey(agentAccAddr), bz)
}

func (k Keeper) GetMTStakingAgentAddressByDenomAndVal(ctx sdk.Context, denom string, valAddr sdk.ValAddress) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMTStakingAgentAddressKey(denom, valAddr))
	if bz == nil {
		return nil, false
	}

	return bz, true
}

func (k Keeper) SetMTStakingDenomAndValWithAgentAddress(ctx sdk.Context, agentAddress sdk.AccAddress, denom string, valAddr sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetMTStakingAgentAddressKey(denom, valAddr), agentAddress)
}

func (k Keeper) GetMTStakingUnbonding(ctx sdk.Context, agentAddress sdk.AccAddress, delegatorAddr sdk.AccAddress) (*types.MTStakingUnbondingDelegation, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMTStakingUnbondingKey(agentAddress, delegatorAddr))
	if bz == nil {
		return nil, false
	}

	unbonding := &types.MTStakingUnbondingDelegation{}
	k.cdc.MustUnmarshal(bz, unbonding)
	return unbonding, true
}

func (k Keeper) GetUnbondingDelegationFromAgent(ctx sdk.Context, agentAddress sdk.AccAddress) (ubds []types.MTStakingUnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.GetMTStakingUnbondingByAgentIndexKey(agentAddress))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		unbonding := types.MTStakingUnbondingDelegation{}
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, &unbonding)
		ubds = append(ubds, unbonding)
	}
	return ubds
}

func (k Keeper) GetOrCreateMTStakingUnbonding(ctx sdk.Context, agentAddress, delegatorAddr sdk.AccAddress) *types.MTStakingUnbondingDelegation {
	unbonding, found := k.GetMTStakingUnbonding(ctx, agentAddress, delegatorAddr)
	if found {
		return unbonding
	}

	unbonding = &types.MTStakingUnbondingDelegation{
		AgentAddress:     string(agentAddress),
		DelegatorAddress: string(delegatorAddr),
		Entries:          []types.MTStakingUnbondingDelegationEntry{},
	}
	return unbonding
}

func (k Keeper) SetMTStakingUnbondingDelegation(ctx sdk.Context, unbonding *types.MTStakingUnbondingDelegation) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(unbonding)

	agentAccAddr := sdk.AccAddress(unbonding.AgentAddress)
	delegatorAccAddr := sdk.AccAddress(unbonding.DelegatorAddress)

	store.Set(types.GetMTStakingUnbondingKey(agentAccAddr, delegatorAccAddr), bz)
}

func (k Keeper) RemoveMTStakingUnbonding(ctx sdk.Context, agentAddress, delegatorAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetMTStakingUnbondingKey(agentAddress, delegatorAddr))
}

func (k Keeper) GetDelegatorAgentShares(ctx sdk.Context, agentAddress, delegator sdk.AccAddress) math.Int {
	amount := math.ZeroInt()
	store := ctx.KVStore(k.storeKey)
	key := types.GetMTStakingSharesKey(agentAddress, delegator)
	bz := store.Get(key)

	if bz == nil {
		return amount
	}

	if err := amount.Unmarshal(bz); err != nil {
		return math.ZeroInt()
	}

	return amount
}

func (k Keeper) IncreaseDelegatorAgentShares(ctx sdk.Context, shares math.Int, agentAddress, delegator sdk.AccAddress) error {
	var err error
	amount := math.ZeroInt()

	store := ctx.KVStore(k.storeKey)
	key := types.GetMTStakingSharesKey(agentAddress, delegator)
	bz := store.Get(key)
	if bz != nil {
		if err = amount.Unmarshal(bz); err != nil {
			return err
		}
	}

	amount = amount.Add(shares)
	if bz, err = amount.Marshal(); err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

func (k Keeper) DecreaseDelegatorAgentShares(ctx sdk.Context, shares math.Int, agentAddress, delegator sdk.AccAddress) error {
	var err error
	var amount math.Int

	store := ctx.KVStore(k.storeKey)
	key := types.GetMTStakingSharesKey(agentAddress, delegator)
	bz := store.Get(key)

	if bz == nil {
		return types.ErrInsufficientShares
	}

	if err = amount.Unmarshal(bz); err != nil {
		return err
	}

	if amount.LT(shares) {
		return types.ErrInsufficientShares
	}

	amount = amount.Sub(shares)
	if amount.IsZero() {
		store.Delete(key)
	}

	if bz, err = amount.Marshal(); err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

// sendCoinsFromAccountToAccount preform send coins form sender to receiver.
func (k Keeper) sendCoinsFromAccountToAccount(ctx sdk.Context, senderAddr sdk.AccAddress, receiverAddr sdk.AccAddress, amt sdk.Coins) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, senderAddr, types.ModuleName, amt); err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiverAddr, amt)
}

// UBDQueueIterator returns all the unbonding queue timeslices from time 0 until endTime.
func (k Keeper) UBDQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.UnbondingQueueKey,
		sdk.InclusiveEndBytes(types.GetMTStakingUnbondingDelegationTimeKey(endTime)))
}

func (k Keeper) InsertUBDQueue(ctx sdk.Context, ubd *types.MTStakingUnbondingDelegation, completionTime time.Time) {
	daPair := types.DAPair{DelegatorAddress: ubd.DelegatorAddress, AgentAddress: ubd.AgentAddress}

	timeSlice := k.GetUBDQueueTimeSlice(ctx, completionTime)
	if len(timeSlice) == 0 {
		k.SetUBDQueueTimeSlice(ctx, completionTime, []types.DAPair{daPair})
	} else {
		timeSlice = append(timeSlice, daPair)
		k.SetUBDQueueTimeSlice(ctx, completionTime, timeSlice)
	}
}

func (k Keeper) GetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time) (dvPairs []types.DAPair) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMTStakingUnbondingDelegationTimeKey(timestamp))
	if bz == nil {
		return []types.DAPair{}
	}

	pairs := types.DAPairs{}
	k.cdc.MustUnmarshal(bz, &pairs)

	return pairs.Pairs
}

func (k Keeper) SetUBDQueueTimeSlice(ctx sdk.Context, timestamp time.Time, keys []types.DAPair) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&types.DAPairs{Pairs: keys})
	store.Set(types.GetMTStakingUnbondingDelegationTimeKey(timestamp), bz)
}

func (k Keeper) GetAllAgent(ctx sdk.Context) []types.MTStakingAgent {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.AgentPrefix)

	iterator := sdk.KVStorePrefixIterator(prefixStore, nil)
	defer iterator.Close()

	agents := []types.MTStakingAgent{}
	for ; iterator.Valid(); iterator.Next() {
		agent := types.MTStakingAgent{}

		err := proto.Unmarshal(iterator.Value(), &agent)
		if err != nil {
			panic(err)
		}
		agents = append(agents, agent)
	}
	return agents
}

func (k Keeper) GetDelegatorWithdrawRewardHeight(ctx sdk.Context, delegatorAccAddr, agentAccAddr sdk.AccAddress) (int64, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMTWithdrawRewardHeightKey(delegatorAccAddr, agentAccAddr)

	bz := store.Get(key)
	if bz == nil {
		return 0, false
	}
	height := sdk.BigEndianToUint64(bz)
	return safeUint64ToInt64(height)
}

func (k Keeper) SetDelegatorWithdrawRewardHeight(ctx sdk.Context, delegatorAccAddr, agentAccAddr sdk.AccAddress, height int64) bool {
	h, ok := safeInt64ToUint64(height)
	if !ok {
		return false
	}

	bz := sdk.Uint64ToBigEndian(h)
	store := ctx.KVStore(k.storeKey)
	key := types.GetMTWithdrawRewardHeightKey(delegatorAccAddr, agentAccAddr)

	store.Set(key, bz)
	return true
}

// TODO BlockHeight should be always unsigned, and maxInt64 is big enough.
func safeUint64ToInt64(value uint64) (int64, bool) {
	if value > stdmath.MaxInt64 {
		return 0, false
	}
	return int64(value), true
}

func safeInt64ToUint64(value int64) (uint64, bool) {
	if value < 0 {
		return 0, false
	}
	return uint64(value), true
}
