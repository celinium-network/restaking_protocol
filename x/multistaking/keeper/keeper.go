package keeper

import (
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/celinium-network/restaking_protocol/x/multistaking/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec

	accountKeeper      types.AccountKeeper
	bankKeeper         types.BankKeeper
	epochKeeper        types.EpochKeeper
	stakingkeeper      types.StakingKeeper
	distributionKeeper types.DistributionKeeper

	EquivalentCoinCalculator CalculateEquivalentCoin
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
		storeKey:                 storeKey,
		cdc:                      cdc,
		accountKeeper:            accountKeeper,
		bankKeeper:               bankKeeper,
		epochKeeper:              epochKeeper,
		stakingkeeper:            stakingKeeper,
		distributionKeeper:       distributionKeeper,
		EquivalentCoinCalculator: defaultCalculateEquivalentCoin,
	}
}

// TODO Temporarily use this method to feed prices !!!
type CalculateEquivalentCoin func(ctx sdk.Context, coin sdk.Coin, targetDenom string) (sdk.Coin, error)

func defaultCalculateEquivalentCoin(ctx sdk.Context, coin sdk.Coin, targetDenom string) (sdk.Coin, error) {
	return sdk.NewCoin(targetDenom, coin.Amount), nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) GetMultiStakingDenomWhiteList(ctx sdk.Context) (*types.MultiStakingDenomWhiteList, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.MultiStakingDenomWhiteListKey)
	if bz == nil {
		return nil, false
	}

	whiteList := &types.MultiStakingDenomWhiteList{}

	if err := k.cdc.Unmarshal(bz, whiteList); err != nil {
		return nil, false
	}

	return whiteList, true
}

func (k Keeper) SetMultiStakingDenom(ctx sdk.Context, denom string) bool {
	whiteList, found := k.GetMultiStakingDenomWhiteList(ctx)
	if !found || whiteList == nil {
		whiteList = &types.MultiStakingDenomWhiteList{
			DenomList: []string{denom},
		}
	} else {
		for _, existedDenom := range whiteList.DenomList {
			if strings.Compare(existedDenom, denom) == 0 {
				return false
			}
		}

		whiteList.DenomList = append(whiteList.DenomList, denom)
	}

	bz, err := k.cdc.Marshal(whiteList)
	if err != nil {
		return false
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.MultiStakingDenomWhiteListKey, bz)

	return true
}

func (k Keeper) GetMultiStakingAgentByID(ctx sdk.Context, id uint64) (*types.MultiStakingAgent, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMultiStakingAgentKey(id))

	if bz == nil {
		return nil, false
	}

	agent := &types.MultiStakingAgent{}
	k.cdc.MustUnmarshal(bz, agent)

	return agent, true
}

func (k Keeper) GetMultiStakingAgent(ctx sdk.Context, denom string, valAddr string) (*types.MultiStakingAgent, bool) {
	agentID, found := k.GetMultiStakingAgentIDByDenomAndVal(ctx, denom, valAddr)
	if !found {
		return nil, false
	}

	return k.GetMultiStakingAgentByID(ctx, agentID)
}

func (k Keeper) SetMultiStakingAgent(ctx sdk.Context, agent *types.MultiStakingAgent) {
	bz := k.cdc.MustMarshal(agent)
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetMultiStakingAgentKey(agent.Id), bz)

	k.SetMultiStakingAgentIDByDenomAndVal(ctx, agent.Id, agent.StakeDenom, agent.ValidatorAddress)
	k.SetLatestMultiStakingAgentID(ctx, agent.Id)
}

func (k Keeper) GetMultiStakingAgentIDByDenomAndVal(ctx sdk.Context, denom string, valAddr string) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMultiStakingAgentIDKey(denom, valAddr))
	if bz == nil {
		return 0, false
	}

	return sdk.BigEndianToUint64(bz), true
}

func (k Keeper) SetMultiStakingAgentIDByDenomAndVal(ctx sdk.Context, id uint64, denom, valAddr string) {
	store := ctx.KVStore(k.storeKey)
	idBz := sdk.Uint64ToBigEndian(id)

	store.Set(types.GetMultiStakingAgentIDKey(denom, valAddr), idBz)
}

func (k Keeper) GetLatestMultiStakingAgentID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.MultiStakingLatestAgentIDKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetLatestMultiStakingAgentID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	idBz := sdk.Uint64ToBigEndian(id)

	store.Set(types.MultiStakingLatestAgentIDKey, idBz)
}

func (k Keeper) GetMultiStakingUnbonding(ctx sdk.Context, agentID uint64, delegatorAddr string) (*types.MultiStakingUnbonding, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetMultiStakingUnbondingKey(agentID, delegatorAddr))
	if bz == nil {
		return nil, false
	}

	unbonding := &types.MultiStakingUnbonding{}
	k.cdc.MustUnmarshal(bz, unbonding)
	return unbonding, true
}

func (k Keeper) GetOrCreateMultiStakingUnbonding(ctx sdk.Context, agentID uint64, delegatorAddr string) *types.MultiStakingUnbonding {
	unbonding, found := k.GetMultiStakingUnbonding(ctx, agentID, delegatorAddr)
	if found {
		return unbonding
	}

	unbonding = &types.MultiStakingUnbonding{
		AgentId:          agentID,
		DelegatorAddress: delegatorAddr,
		Entries:          []types.MultiStakingUnbondingEntry{},
	}
	return unbonding
}

func (k Keeper) SetMultiStakingUnbonding(ctx sdk.Context, agentID uint64, delegatorAddr string, unbonding *types.MultiStakingUnbonding) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(unbonding)

	store.Set(types.GetMultiStakingUnbondingKey(agentID, delegatorAddr), bz)
}

func (k Keeper) RemoveMultiStakingUnbonding(ctx sdk.Context, agentID uint64, delegatorAddr string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetMultiStakingUnbondingKey(agentID, delegatorAddr))
}

func (k Keeper) GetMultiStakingShares(ctx sdk.Context, agentID uint64, delegator string) math.Int {
	amount := math.ZeroInt()
	store := ctx.KVStore(k.storeKey)
	key := types.GetMultiStakingSharesKey(agentID, delegator)
	bz := store.Get(key)

	if bz == nil {
		return amount
	}

	if err := amount.Unmarshal(bz); err != nil {
		return math.ZeroInt()
	}

	return amount
}

func (k Keeper) IncreaseMultiStakingShares(ctx sdk.Context, shares math.Int, agentID uint64, delegator string) error {
	var err error
	amount := math.ZeroInt()

	store := ctx.KVStore(k.storeKey)
	key := types.GetMultiStakingSharesKey(agentID, delegator)
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

func (k Keeper) DecreaseMultiStakingShares(ctx sdk.Context, shares math.Int, agentID uint64, delegator string) error {
	var err error
	var amount math.Int

	store := ctx.KVStore(k.storeKey)
	key := types.GetMultiStakingSharesKey(agentID, delegator)
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
	return store.Iterator(types.MultiStakingUnbondingQueueKey,
		sdk.InclusiveEndBytes(types.GetMultiStakingUnbondingDelegationTimeKey(endTime)))
}

func (k Keeper) InsertUBDQueue(ctx sdk.Context, ubd *types.MultiStakingUnbonding, completionTime time.Time) {
	daPair := types.DAPair{DelegatorAddress: ubd.DelegatorAddress, AgentId: ubd.AgentId}

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

	bz := store.Get(types.GetMultiStakingUnbondingDelegationTimeKey(timestamp))
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
	store.Set(types.GetMultiStakingUnbondingDelegationTimeKey(timestamp), bz)
}

func (k Keeper) GetAllAgent(ctx sdk.Context) []types.MultiStakingAgent {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.MultiStakingAgentPrefix)

	iterator := sdk.KVStorePrefixIterator(prefixStore, nil)
	defer iterator.Close()

	agents := []types.MultiStakingAgent{}
	for ; iterator.Valid(); iterator.Next() {
		agent := types.MultiStakingAgent{}

		err := proto.Unmarshal(iterator.Value(), &agent)
		if err != nil {
			panic(err)
		}
		agents = append(agents, agent)
	}
	return agents
}
