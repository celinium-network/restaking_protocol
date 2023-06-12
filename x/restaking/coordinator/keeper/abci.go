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
			OperatorAddress: operator.OperatorAddress,
			ValidatorPk:     va.ValidatorPk,
			Amount:          amount,
		}

		restakingPacket, err := restaking.BuildRestakingProtocolPacket(k.cdc, delegationPacket)
		if err != nil {
			ctx.Logger().Error("marshal restaking.Delegation has err: ", err)
			// TODO continue ?
			continue
		}

		restakingProtocolPacketBz, err := k.cdc.Marshal(restakingPacket)
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
