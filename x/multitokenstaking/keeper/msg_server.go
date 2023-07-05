package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

type msgServer struct {
	*Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// MTStakingDelegate implements types.MsgServer.
func (k msgServer) MTStakingDelegate(goCtx context.Context, msg *types.MsgMTStakingDelegate) (*types.MsgMTStakingDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	delegatorAccAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	newShares, err := k.Keeper.MTStakingDelegate(ctx, delegatorAccAddr, valAddr, msg.Balance)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Balance.String()),
			sdk.NewAttribute(types.AttributeKeyNewShares, newShares.String()),
		),
	})

	return &types.MsgMTStakingDelegateResponse{}, err
}

// MTStakingUndelegate implements types.MsgServer.
func (k msgServer) MTStakingUndelegate(goCtx context.Context, msg *types.MsgMTStakingUndelegate) (*types.MsgMTStakingUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	completionTime, err := k.Keeper.MTStakingUndelegate(ctx, delegatorAccAddr, valAddr, msg.Balance)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Balance.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
	})

	return &types.MsgMTStakingUndelegateResponse{
		CompletionTime: completionTime,
		Amount:         msg.Balance,
	}, err
}

// MTStakingWithdrawReward implements types.MsgServer.
func (msgServer) MTStakingWithdrawReward(context.Context, *types.MsgMTStakingWithdrawReward) (*types.MsgMTStakingWithdrawRewardResponse, error) {
	panic("unimplemented")
}
