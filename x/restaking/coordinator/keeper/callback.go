package keeper

import (
	"fmt"

	"golang.org/x/exp/slices"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/celinium-network/restaking_protocol/utils"
	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (k Keeper) HandleIBCAcknowledgement(ctx sdk.Context, packet *channeltypes.Packet, acknowledgement []byte) error {
	callback, found := k.GetCallback(ctx, packet.SourceChannel, packet.SourcePort, packet.Sequence)
	if !found {
		return types.ErrIBCCallbackNotExisted
	}

	switch callback.CallType {
	case types.InterChainDelegateCall:
		operatorDelegationRecordKey := callback.Args
		record, found := k.GetOperatorDelegateRecordByKey(ctx, operatorDelegationRecordKey)
		if !found {
			return types.ErrAdditionalProposalNotFound
		}
		callbackID := types.IBCCallbackKey(packet.SourceChannel, packet.SourcePort, packet.Sequence)
		if record.Status != types.OpDelRecordProcessing {
			return types.ErrMismatchStatus
		}

		// remove callback id
		index := slices.Index(record.IbcCallbackIds, string(callbackID))
		record.IbcCallbackIds = slices.Delete(record.IbcCallbackIds, index, index+1)
		if len(record.IbcCallbackIds) == 0 {
			operatorAccAddr := sdk.MustAccAddressFromBech32(record.OperatorAddress)
			operator, found := k.GetOperator(ctx, operatorAccAddr)
			if !found {
				return types.ErrUnknownOperator
			}
			operator.RestakedAmount = operator.RestakedAmount.Add(record.DelegationAmount)
			k.DeleteOperatorDelegateRecordByKey(ctx, operatorDelegationRecordKey)
			k.SetOperator(ctx, operatorAccAddr, operator)
		} else {
			k.SetOperatorDelegateRecordByKey(ctx, operatorDelegationRecordKey, record)
		}
	case types.InterChainUndelegateCall:
		operatorUndelegationRecordKey := callback.Args
		record, found := k.GetOperatorUndelegationRecordByKey(ctx, operatorUndelegationRecordKey)
		if !found {
			// TODO correct error
			return types.ErrAdditionalProposalNotFound
		}

		callbackID := types.IBCCallbackKey(packet.SourceChannel, packet.SourcePort, packet.Sequence)
		if record.Status != types.OpDelRecordProcessing {
			return types.ErrMismatchStatus
		}
		index := slices.Index(record.IbcCallbackIds, string(callbackID))
		record.IbcCallbackIds = slices.Delete(record.IbcCallbackIds, index, index+1)

		ackResp, err := GetResultFromAcknowledgement(acknowledgement)
		if err != nil {
			return err
		}
		var resp restaking.ConsumerUndelegateResponse
		k.cdc.MustUnmarshal(ackResp, &resp)
		if record.LatestCompleteTime < resp.CompletionTime {
			record.LatestCompleteTime = resp.CompletionTime
		}

		if len(record.IbcCallbackIds) == 0 {
			operatorAccAddr := sdk.MustAccAddressFromBech32(record.OperatorAddress)
			operator, found := k.GetOperator(ctx, operatorAccAddr)
			if !found {
				return types.ErrUnknownOperator
			}

			operator.RestakedAmount = operator.RestakedAmount.Sub(record.UndelegationAmount)

			for _, ubdEntryID := range record.UnbondingEntryIds {
				unbondingDelegation, found := k.GetUnbondingDelegationByUnbondingID(ctx, ubdEntryID)
				if !found {
					// TODO how to do
					ctx.Logger().Error("...")
				}
				for i := 0; i < len(unbondingDelegation.Entries); i++ {
					if unbondingDelegation.Entries[i].Id != ubdEntryID {
						// TODO how to do? Just log it?
						continue
					}
					unbondingDelegation.Entries[i].CompleteTime = record.LatestCompleteTime
				}

				k.SetUnbondingDelegation(ctx, unbondingDelegation)
				k.InsertUBDQueue(ctx, unbondingDelegation, utils.ConvertNanoSecondToTime(record.LatestCompleteTime))
			}

			k.DeleteOperatorDelegateRecordByKey(ctx, operatorUndelegationRecordKey)
		} else {
			k.SetOperatorUndelegationRecordByKey(ctx, operatorUndelegationRecordKey, record)
		}
	case types.InterChainWithdrawRewardCall:
		k.HandleOperatorWithdrawRewardCallback(ctx, packet, acknowledgement, callback)
	default:
		return types.ErrUnknownIBCCallbackType
	}

	return nil
}

func GetResultFromAcknowledgement(acknowledgement []byte) ([]byte, error) {
	var ack channeltypes.Acknowledgement
	if err := channeltypes.SubModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return nil, err
	}

	switch response := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		if len(response.Result) == 0 {
			return nil, sdkerrors.Wrapf(channeltypes.ErrInvalidAcknowledgement, "empty acknowledgement")
		}
		return ack.GetResult(), nil
	case *channeltypes.Acknowledgement_Error:
		return nil, fmt.Errorf("acknowledgement has error: %s", ack.GetError())
	default:
		return nil, fmt.Errorf("unknown acknowledgement status")
	}
}
