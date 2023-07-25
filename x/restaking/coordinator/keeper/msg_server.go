package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

// Delegate implements types.MsgServer.
func (m msgServer) Delegate(goCtx context.Context, msg *types.MsgDelegateRequest) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	operatorAccAddr, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return nil, err
	}

	err = m.keeper.Delegate(ctx, delegatorAccAddr, operatorAccAddr, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgDelegateResponse{}, nil
}

// RegisterOperator implements types.MsgServer.
func (m msgServer) RegisterOperator(goCtx context.Context, msg *types.MsgRegisterOperatorRequest) (*types.MsgRegisterOperatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.keeper.RegisterOperator(ctx, *msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterOperatorResponse{}, nil
}

// Undelegate implements types.MsgServer.
func (m msgServer) Undelegate(goCtx context.Context, msg *types.MsgUndelegateRequest) (*types.MsgUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	operatorAccAddr, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return nil, err
	}

	if err = m.keeper.Undelegate(ctx, delegatorAccAddr, operatorAccAddr, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgUndelegateResponse{}, nil
}

// WithdrawReward implements types.MsgServer.
func (m msgServer) WithdrawReward(goCtx context.Context, msg *types.MsgWithdrawRewardRequest) (*types.MsgWithdrawRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	operatorAccAddr, err := sdk.AccAddressFromBech32(msg.OperatorAddress)
	if err != nil {
		return nil, err
	}

	err = m.keeper.WithdrawDelegatorRewards(ctx, delegatorAccAddr, operatorAccAddr)
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawRewardResponse{}, nil
}
