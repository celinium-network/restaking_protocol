package keeper

import (
	"fmt"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"golang.org/x/exp/slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) HandleOperatorWithdrawRewardCallback(
	ctx sdk.Context,
	packet *channeltypes.Packet,
	acknowledgement []byte,
	callback *types.IBCCallback,
) error {
	if callback.CallType != types.InterChainWithdrawRewardCall {
		// TODO check or not check? panic?
		panic("mismatch callback type with handler")
	}

	recordKey := callback.Args
	record, found := k.GetOperatorWithdrawRewardRecordByKey(ctx, recordKey)
	if !found {
		// TODO correct error
		return types.ErrMismatchStatus
	}

	callbackID := types.IBCCallbackKey(packet.SourceChannel, packet.SourcePort, packet.Sequence)
	index := slices.Index(record.IbcCallbackIds, string(callbackID))
	record.Statues[index] = types.OpTransferringReward
	if len(record.TransferIds) == 0 {
		record.TransferIds = make([]string, len(record.IbcCallbackIds))
		record.Rewards = make([]sdk.Coin, len(record.IbcCallbackIds))
	}

	slices.Delete(record.IbcCallbackIds, index, index+1)

	ackResp, err := GetResultFromAcknowledgement(acknowledgement)
	if err != nil {
		return err
	}

	var resp restaking.ConsumerWithdrawRewardResponse
	k.cdc.MustUnmarshal(ackResp, &resp)

	transferKey := types.ConsumerTransferRewardKey(resp.TransferDestChannel, resp.TransferDestPort, resp.TransferDestSeq)
	record.TransferIds[index] = string(transferKey)

	k.SetOperatorWithdrawRewardRecordByKey(ctx, recordKey, &record)
	k.SetTransferIDToWithdrawRewardRecordKey(ctx, string(transferKey), recordKey)

	return nil
}

// OnOperatorReceiveReward define a method for operator receive restaking reward.
// It should be called at ibc transfer ack.
// When receive rewards from all consumers then the operator start a new period
func (k Keeper) OnOperatorReceiveReward(ctx sdk.Context, chainID string, operatorAccAddr sdk.AccAddress, rewards []sdk.Coin) {
	var rewardRatios sdk.DecCoins
	operator, found := k.GetOperator(ctx, operatorAccAddr)
	if !found {
		panic("not found")
	}
	for _, r := range rewards {
		rewardRatios = append(rewardRatios, sdk.NewDecCoin(r.Denom, r.Amount.Quo(operator.RestakedAmount)))
	}

	lastPeriod, found := k.GetOperatorLastRewardPeriod(ctx, operatorAccAddr)
	if !found {
		lastPeriod = 0
		k.SetOperatorHistoricalRewards(ctx, lastPeriod, operatorAccAddr, types.OperatorHistoricalRewards{
			CumulativeRewardRatios: rewardRatios,
		})
	} else {
		lastHistoricalReward, found := k.GetOperatorHistoricalRewards(ctx, lastPeriod-1, operatorAccAddr)
		if !found {
			panic("todo")
		}

		lastHistoricalReward.CumulativeRewardRatios = rewardRatios.Add(lastHistoricalReward.CumulativeRewardRatios...)
		k.SetOperatorHistoricalRewards(ctx, lastPeriod, operatorAccAddr, lastHistoricalReward)
	}

	nextPeriod := lastPeriod + 1
	k.SetOperatorLastRewardPeriod(ctx, operatorAccAddr, nextPeriod)
}

// WithdrawOperatorsReward define a method to withdraw all operator restaking reward from consumer
// TODO there maybe to many operators, so call it by offChain service? or make a queues, don't iterate all operator?
func (k Keeper) WithdrawOperatorsReward(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{types.OperatorPrefix})

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var operator types.Operator
		if err := k.cdc.Unmarshal(bz, &operator); err != nil {
			continue
		}

		k.operatorWithdrawReward(ctx, &operator)
	}
}

// operatorWithdrawReward define a method to withdraw the operator reward
func (k Keeper) operatorWithdrawReward(ctx sdk.Context, operator *types.Operator) error {
	withdrawingRecord := types.OperatorWithdrawRewardRecord{
		OperatorAddress: operator.OperatorAddress,
	}
	operatorAccAddr := sdk.MustAccAddressFromBech32(operator.OperatorAddress)

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

		withdrawPacket := restaking.WithdrawRewardPacket{
			OperatorAddress:  operator.OperatorAddress,
			ValidatorAddress: va.ValidatorAddress,
		}

		bz := k.cdc.MustMarshal(&withdrawPacket)
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

		callback := types.IBCCallback{
			CallType: types.InterChainWithdrawRewardCall,
			Args:     string(types.DelegationRecordKey(uint64(ctx.BlockHeight()), operatorAccAddr)),
		}

		ibcCallbackKey := types.IBCCallbackKey(channel, restaking.CoordinatorPortID, seq)
		withdrawingRecord.IbcCallbackIds = append(withdrawingRecord.IbcCallbackIds, string(ibcCallbackKey))
		withdrawingRecord.Statues = append(withdrawingRecord.Statues, types.OpWithdrawingReward)

		k.SetCallback(ctx, channel, restaking.CoordinatorPortID, seq, callback)
	}

	k.SetOperatorWithdrawRewardRecord(ctx, uint64(ctx.BlockHeight()), operatorAccAddr, &withdrawingRecord)
	return nil
}

func (k Keeper) AfterOperatorCreated(ctx sdk.Context) {}

func (k Keeper) BeforeDelegationSharesModified(ctx sdk.Context) {}

func (k Keeper) AfterDelegationSharesModified(ctx sdk.Context) {}

func (k Keeper) SetOperatorWithdrawRewardRecord(ctx sdk.Context, blockHeight uint64, operatorAccAddr sdk.AccAddress, withdraw *types.OperatorWithdrawRewardRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(withdraw)
	store.Set(types.OperatorWithdrawRecordKey(blockHeight, operatorAccAddr), bz)
}

func (k Keeper) GetOperatorWithdrawRewardRecordByKey(ctx sdk.Context, recordKey string) (types.OperatorWithdrawRewardRecord, bool) {
	var record types.OperatorWithdrawRewardRecord

	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(recordKey))
	if len(bz) == 0 {
		return record, false
	}
	err := k.cdc.Unmarshal(bz, &record)
	if err != nil {
		return record, false
	}

	return record, true
}

func (k Keeper) OnRecvIBCTransferPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	transferID := types.ConsumerTransferRewardKey(packet.DestinationChannel, packet.DestinationPort, packet.Sequence)

	recordKey, found := k.GetWithdrawRewardRecordKeyFromTransferID(ctx, string(transferID))
	if !found {
		return nil
	}

	record, found := k.GetOperatorWithdrawRewardRecordByKey(ctx, string(recordKey))
	if !found {
		// TODO correct error
		return types.ErrMismatchStatus
	}
	token, err := getCoinFromTransferPacket(&packet)
	if err != nil {
		return err
	}

	index := slices.Index(record.TransferIds, string(transferID))
	record.Statues[index] = types.OpTransferredReward
	slices.Delete(record.TransferIds, index, index+1)
	record.Rewards[index] = token

	if len(record.IbcCallbackIds) != 0 || len(record.TransferIds) != 0 {
		k.SetOperatorWithdrawRewardRecordByKey(ctx, string(recordKey), &record)
		// TODO delete withdraw reward record now?
		return nil
	}

	k.OnOperatorReceiveReward(ctx, "", sdk.AccAddress(record.OperatorAddress), record.Rewards)
	return nil
}

func (k Keeper) SetOperatorWithdrawRewardRecordByKey(ctx sdk.Context, recordKey string, record *types.OperatorWithdrawRewardRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(record)
	store.Set([]byte(recordKey), bz)
}

func (k Keeper) SetTransferIDToWithdrawRewardRecordKey(ctx sdk.Context, transferID, withdrawRecordKey string) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(transferID), []byte(withdrawRecordKey))
}

func (k Keeper) GetWithdrawRewardRecordKeyFromTransferID(ctx sdk.Context, transferID string) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(transferID))

	if len(bz) == 0 {
		return nil, false
	}
	return bz, true
}

func getCoinFromTransferPacket(packet *channeltypes.Packet) (sdk.Coin, error) {
	var coin sdk.Coin
	var data transfertypes.FungibleTokenPacketData
	transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data)

	transferAmount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return coin, sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into math.Int", data.Amount)
	}

	voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	unprefixedDenom := data.Denom[len(voucherPrefix):]

	// coin denomination used in sending from the escrow address
	denom := unprefixedDenom

	// The denomination used to send the coins is either the native denom or the hash of the path
	// if the denomination is not native.
	denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
	if denomTrace.Path != "" {
		denom = denomTrace.IBCDenom()
	}
	coin = sdk.NewCoin(denom, transferAmount)

	return coin, nil
}

func (k Keeper) GetOperatorHistoricalRewards(ctx sdk.Context, period uint64, operatorAccAddr sdk.AccAddress) (history types.OperatorHistoricalRewards, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OperatorHistoricalRewardKey(period, operatorAccAddr))
	if bz == nil {
		return history, false
	}

	err := k.cdc.Unmarshal(bz, &history)
	if err != nil {
		return history, false
	}
	return history, true
}

func (k Keeper) SetOperatorHistoricalRewards(ctx sdk.Context, period uint64, operatorAccAddr sdk.AccAddress, history types.OperatorHistoricalRewards) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&history)
	store.Set(types.OperatorHistoricalRewardKey(period, operatorAccAddr), bz)
}

func (k Keeper) GetOperatorLastRewardPeriod(ctx sdk.Context, operatorAccAddr sdk.AccAddress) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OperatorLastRewardPeriodKey(operatorAccAddr))
	if bz == nil {
		return 0, false
	}
	return sdk.BigEndianToUint64(bz), true
}

func (k Keeper) SetOperatorLastRewardPeriod(ctx sdk.Context, operatorAccAddr sdk.AccAddress, period uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(period)
	store.Set(types.OperatorLastRewardPeriodKey(operatorAccAddr), bz)
}