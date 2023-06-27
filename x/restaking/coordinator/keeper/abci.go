package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) {
	k.ProcessPendingOperatorDelegationRecord(ctx)

	k.ProcessPendingOperatorUndelegationRecord(ctx)

	k.ProcessCompletedUnbonding(ctx)
}

func (k Keeper) ProcessCompletedUnbonding(ctx sdk.Context) {
	matureUnbonds := k.DequeueAllMatureUBDQueue(ctx, ctx.BlockHeader().Time)
	for _, dvPair := range matureUnbonds {
		delAccAddr := sdk.MustAccAddressFromBech32(dvPair.Delegator)
		opAccAddr := sdk.MustAccAddressFromBech32(dvPair.Operator)
		_, err := k.CompleteUnbonding(ctx, delAccAddr, opAccAddr)
		if err != nil {
			continue
		}
	}
}

func (k Keeper) DequeueAllMatureUBDQueue(ctx sdk.Context, currTime time.Time) (matureUnbonds []types.DOPair) {
	store := ctx.KVStore(k.storeKey)

	unbondingTimesliceIterator := k.UBDQueueIterator(ctx, currTime)
	defer unbondingTimesliceIterator.Close()

	for ; unbondingTimesliceIterator.Valid(); unbondingTimesliceIterator.Next() {
		timeslice := types.DOPairs{}
		value := unbondingTimesliceIterator.Value()
		k.cdc.MustUnmarshal(value, &timeslice)

		matureUnbonds = append(matureUnbonds, timeslice.Pairs...)

		store.Delete(unbondingTimesliceIterator.Key())
	}

	return matureUnbonds
}

func (k Keeper) CompleteUnbonding(ctx sdk.Context, delAddr sdk.AccAddress, opAddr sdk.AccAddress) (sdk.Coins, error) {
	ubd, found := k.GetUnbondingDelegation(ctx, delAddr, opAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	balances := sdk.NewCoins()
	ctxTime := ctx.BlockHeader().Time

	// loop through all the entries and complete unbonding mature entries
	for i := 0; i < len(ubd.Entries); i++ {
		entry := ubd.Entries[i]
		if entry.IsMature(ctxTime) {
			ubd.RemoveEntry(int64(i))
			i--
			k.DeleteUnbondingIndex(ctx, entry.Id)

			// track undelegation only when remaining or truncated shares are non-zero
			if !entry.Amount.IsZero() {
				if err := k.sendCoinsFromAccountToAccount(
					ctx, opAddr, delAddr, sdk.Coins{entry.Amount},
				); err != nil {
					return nil, err
				}

				balances = balances.Add(entry.Amount)
			}
		}
	}

	// set the unbonding delegation or remove it if there are no more entries
	if len(ubd.Entries) == 0 {
		k.RemoveUnbondingDelegation(ctx, ubd)
	} else {
		k.SetUnbondingDelegation(ctx, ubd)
	}

	return balances, nil
}

func (k Keeper) ProcessPendingOperatorDelegationRecord(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{types.DelegationRecordPrefix})
	defer iterator.Close()

	var pendingRecords []types.OperatorDelegationRecord

	for ; iterator.Valid(); iterator.Next() {
		var record types.OperatorDelegationRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			ctx.Logger().Error("unmarshal OperatorDelegationRecord failed, key ", string(iterator.Key()))
			continue
		}

		if record.Status != types.OpDelRecordPending {
			continue
		}

		pendingRecords = append(pendingRecords, record)

		store.Delete(iterator.Key())
	}

	recordMap := make(map[string]math.Int)
	var recordMapKeys []string
	for _, r := range pendingRecords {
		amount, ok := recordMap[r.OperatorAddress]
		if !ok {
			recordMap[r.OperatorAddress] = r.DelegationAmount
			recordMapKeys = append(recordMapKeys, r.OperatorAddress)
		} else {
			recordMap[r.OperatorAddress] = amount.Add(r.DelegationAmount)
		}
	}

	for _, key := range recordMapKeys {
		amount := recordMap[key]
		operator, found := k.GetOperator(ctx, key)
		if !found {
			ctx.Logger().Error("operator not found, operator address: ", key)
			continue
		}
		k.SendDelegation(ctx, operator, amount)
	}
}

// TODO change amount type from math.Int to sdk.Coin
func (k Keeper) SendDelegation(ctx sdk.Context, operator *types.Operator, amount math.Int) {
	if amount.IsZero() {
		return
	}

	processingRecord := types.OperatorDelegationRecord{
		OperatorAddress:  operator.OperatorAddress,
		DelegationAmount: amount,
		Status:           types.OpDelRecordProcessing,
		IbcCallbackIds:   []string{},
	}

	for _, va := range operator.OperatedValidators {
		tmClientID, found := k.GetConsumerClientID(ctx, va.ChainID)
		if !found {
			ctx.Logger().Error("operator contain chain which has't tendermint light client. ChainID: ",
				va.ChainID, " Operator address", operator.OperatorAddress)
			continue
		}
		channel, found := k.GetConsumerClientIDToChannel(ctx, string(tmClientID))
		if !found {
			ctx.Logger().Error(fmt.Sprintf(
				"the consumer chain of operator has't IBC Channel, chainID: %s, operator address: %s",
				va.ChainID, operator.OperatedValidators))
			continue
		}

		// TODO correct TIMEOUT
		timeout := time.Minute * 10

		delegationPacket := restaking.DelegationPacket{
			OperatorAddress:  operator.OperatorAddress,
			ValidatorAddress: va.ValidatorAddress,
			Balance:          sdk.NewCoin(operator.RestakingDenom, amount),
		}

		bz := k.cdc.MustMarshal(&delegationPacket)
		restakingPacket := restaking.CoordinatorPacket{
			Type: 0,
			Data: string(bz),
		}

		restakingProtocolPacketBz, err := k.cdc.Marshal(&restakingPacket)
		if err != nil {
			ctx.Logger().Error("marshal restaking.Delegation has err: ", err)
			// TODO continue ?
			continue
		}
		seq, err := restaking.SendIBCPacket(
			ctx,
			k.scopedKeeper,
			k.channelKeeper,
			channel,
			restaking.CoordinatorPortID,
			restakingProtocolPacketBz,
			timeout,
		)
		if err != nil {
			ctx.Logger().Error("send ibc packet has error:", err)
		}

		ibcCallbackKey := types.IBCCallbackKey(channel, restaking.CoordinatorPortID, seq)
		processingRecord.IbcCallbackIds = append(processingRecord.IbcCallbackIds, string(ibcCallbackKey))

		callback := types.IBCCallback{
			CallType: types.InterChainDelegateCall,
			Args:     string(types.DelegationRecordKey(uint64(ctx.BlockHeight()), operator.OperatorAddress)),
		}

		k.SetCallback(ctx, channel, restaking.CoordinatorPortID, seq, callback)
	}

	k.SetOperatorDelegateRecord(ctx, uint64(ctx.BlockHeight()), &processingRecord)
}

func (k Keeper) ProcessPendingOperatorUndelegationRecord(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{types.UndelegationRecordPrefix})
	defer iterator.Close()

	var pendingRecords []types.OperatorUndelegationRecord

	for ; iterator.Valid(); iterator.Next() {
		var record types.OperatorUndelegationRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			ctx.Logger().Error("unmarshal OperatorDelegationRecord failed, key ", string(iterator.Key()))
			continue
		}

		if record.Status != types.OpUndelegationRecordPending {
			continue
		}

		pendingRecords = append(pendingRecords, record)

		store.Delete(iterator.Key())
	}

	recordMap := make(map[string]math.Int)
	entryIDsMap := make(map[string][]uint64)
	var recordMapKeys []string
	for _, r := range pendingRecords {
		amount, ok := recordMap[r.OperatorAddress]
		if !ok {
			recordMap[r.OperatorAddress] = r.UndelegationAmount
			recordMapKeys = append(recordMapKeys, r.OperatorAddress)
		} else {
			recordMap[r.OperatorAddress] = amount.Add(r.UndelegationAmount)
		}

		entryIDs, ok := entryIDsMap[r.OperatorAddress]
		if !ok {
			entryIDsMap[r.OperatorAddress] = r.UnbondingEntryIds
		} else {
			entryIDsMap[r.OperatorAddress] = append(entryIDs, r.UnbondingEntryIds...)
		}
	}

	for _, key := range recordMapKeys {
		amount := recordMap[key]
		operator, found := k.GetOperator(ctx, key)
		if !found {
			ctx.Logger().Error("operator not found, operator address: ", key)
			continue
		}
		entryIDs := entryIDsMap[key]
		k.SendUndelegation(ctx, operator, amount, entryIDs)
	}
}

// TODO change amount type from math.Int to sdk.Coin
func (k Keeper) SendUndelegation(ctx sdk.Context, operator *types.Operator, amount math.Int, entryIDs []uint64) {
	if amount.IsZero() {
		return
	}

	processingRecord := types.OperatorUndelegationRecord{
		OperatorAddress:    operator.OperatorAddress,
		UndelegationAmount: amount,
		Status:             types.OpDelRecordProcessing,
		IbcCallbackIds:     []string{},
		UnbondingEntryIds:  entryIDs,
	}

	for _, va := range operator.OperatedValidators {
		tmClientID, found := k.GetConsumerClientID(ctx, va.ChainID)
		if !found {
			ctx.Logger().Error("operator contain chain which has't tendermint light client. ChainID: ",
				va.ChainID, " Operator address", operator.OperatorAddress)
			continue
		}
		channel, found := k.GetConsumerClientIDToChannel(ctx, string(tmClientID))
		if !found {
			ctx.Logger().Error(fmt.Sprintf(
				"the consumer chain of operator has't IBC Channel, chainID: %s, operator address: %s",
				va.ChainID, operator.OperatedValidators))
			continue
		}

		// TODO correct TIMEOUT
		timeout := time.Minute * 10

		delegationPacket := restaking.UndelegationPacket{
			OperatorAddress:  operator.OperatorAddress,
			ValidatorAddress: va.ValidatorAddress,
			Balance:          sdk.NewCoin(operator.RestakingDenom, amount),
		}

		bz := k.cdc.MustMarshal(&delegationPacket)
		restakingPacket := restaking.CoordinatorPacket{
			Type: 1,
			Data: string(bz),
		}

		restakingProtocolPacketBz, err := k.cdc.Marshal(&restakingPacket)
		if err != nil {
			ctx.Logger().Error("marshal restaking.Delegation has err: ", err)
			// TODO continue ?
			continue
		}
		seq, err := restaking.SendIBCPacket(
			ctx,
			k.scopedKeeper,
			k.channelKeeper,
			channel,
			restaking.CoordinatorPortID,
			restakingProtocolPacketBz,
			timeout,
		)
		if err != nil {
			ctx.Logger().Error("send ibc packet has error:", err)
		}

		ibcCallbackKey := types.IBCCallbackKey(channel, restaking.CoordinatorPortID, seq)
		processingRecord.IbcCallbackIds = append(processingRecord.IbcCallbackIds, string(ibcCallbackKey))

		callback := types.IBCCallback{
			CallType: types.InterChainUndelegateCall,
			Args:     string(types.UndelegationRecordKey(uint64(ctx.BlockHeight()), operator.OperatorAddress)),
		}

		k.SetCallback(ctx, channel, restaking.CoordinatorPortID, seq, callback)
	}

	k.SetOperatorUndelegationRecord(ctx, uint64(ctx.BlockHeight()), &processingRecord)
}
