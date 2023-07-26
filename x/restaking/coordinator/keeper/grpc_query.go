package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

// Operators implements types.QueryServer.
func (k Querier) Operators(goCtx context.Context, req *types.QueryOperatorsRequest) (*types.QueryOperatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	sharesStore := prefix.NewStore(store, []byte{types.OperatorPrefix})

	operators, pageRes, err := query.GenericFilteredPaginate(
		k.cdc, sharesStore, req.Pagination,
		func(key []byte, delegation *types.Operator) (*types.Operator, error) {
			return delegation, nil
		},
		func() *types.Operator {
			return &types.Operator{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var os []types.Operator
	for _, a := range operators {
		os = append(os, *a)
	}

	return &types.QueryOperatorsResponse{
		Operators:  os,
		Pagination: pageRes,
	}, nil
}

// Operator implements types.QueryServer.
func (k Querier) Operator(goCtx context.Context, req *types.QueryOperatorRequest) (*types.QueryOperatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	operatorAccAddr, err := sdk.AccAddressFromBech32(req.OperatorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "operator address is invalid")
	}

	operator, found := k.GetOperator(ctx, operatorAccAddr)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "not found")
	}

	return &types.QueryOperatorResponse{
		Operator: *operator,
	}, nil
}

// OperatorDelegations implements types.QueryServer.
func (k Querier) OperatorDelegations(goCtx context.Context, req *types.QueryOperatorDelegationsRequest) (*types.QueryOperatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.OperatorAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	var delegations []types.Delegation
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, []byte{types.DelegationPrefix})

	pageRes, err := query.Paginate(delegationStore, req.Pagination, func(key []byte, value []byte) error {
		var delegation types.Delegation
		err := k.cdc.Unmarshal(value, &delegation)
		if err != nil {
			return err
		}

		if delegation.Delegator != string(delegatorAccAddr) {
			return nil
		}

		delegations = append(delegations, delegation)

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryOperatorDelegationsResponse{
		Delegations: delegations,
		Pagination:  pageRes,
	}, nil
}

// Delegation implements types.QueryServer.
func (k Querier) Delegation(goCtx context.Context, req *types.QueryDelegationRequest) (*types.QueryDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	operatorAccAddr, err := sdk.AccAddressFromBech32(req.OperatorAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	delegation, found := k.GetDelegation(ctx, operatorAccAddr, delegatorAccAddr)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "delegation not found")
	}

	operator, found := k.GetOperator(ctx, operatorAccAddr)
	if !found {
		return nil, status.Errorf(codes.Internal, "agent not found")
	}

	delegationAmt := delegation.Shares.Mul(operator.RestakedAmount).Quo(operator.Shares)

	return &types.QueryDelegationResponse{
		Delegation: *delegation,
		Balance:    sdk.NewCoin(operator.RestakingDenom, delegationAmt),
	}, nil
}

// DelegatorDelegations implements types.QueryServer.
func (k Querier) DelegatorDelegations(goCtx context.Context, req *types.QueryDelegatorDelegationsRequest) (*types.QueryDelegatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "agent address is invalid, %s", err)
	}

	// TODO replace it by copy?
	delegationKey := append([]byte{types.DelegationPrefix}, delegatorAccAddr...)
	store := ctx.KVStore(k.storeKey)
	delegationsStore := prefix.NewStore(store, delegationKey)

	delegations, pageRes, err := query.GenericFilteredPaginate(k.cdc, delegationsStore, req.Pagination,
		func(key []byte, delegation *types.Delegation) (*types.Delegation, error) {
			return delegation, nil
		}, func() *types.Delegation {
			return &types.Delegation{}
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var dels []types.Delegation
	for _, d := range delegations {
		dels = append(dels, *d)
	}

	return &types.QueryDelegatorDelegationsResponse{
		Delegations: dels,
		Pagination:  pageRes,
	}, nil
}

// DelegatorOperators implements types.QueryServer.
func (k Querier) DelegatorOperators(goCtx context.Context, req *types.QueryDelegatorOperatorsRequest) (*types.QueryDelegatorOperatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	var operators []types.Operator
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, []byte{types.DelegationPrefix})

	pageRes, err := query.Paginate(delegationStore, req.Pagination, func(key []byte, value []byte) error {
		var delegation types.Delegation
		err := k.cdc.Unmarshal(value, &delegation)
		if err != nil {
			return err
		}

		if delegation.Delegator != string(delegatorAccAddr) {
			return nil
		}

		operatorAccAddr := sdk.AccAddress(delegation.Operator)
		operator, found := k.GetOperator(ctx, operatorAccAddr)
		if !found {
			return types.ErrUnknownOperator
		}

		operators = append(operators, *operator)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDelegatorOperatorsResponse{
		Operators:  operators,
		Pagination: pageRes,
	}, nil
}

// DelegatorUnbondingDelegations implements types.QueryServer.
func (k Querier) DelegatorUnbondingDelegations(goCtx context.Context, req *types.QueryDelegatorUnbondingDelegationsRequest) (*types.QueryDelegatorUnbondingDelegationsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "agent address is invalid, %s", err)
	}

	// TODO replace it by copy?
	agentUnbondingDelegationKey := append([]byte{types.UnbondingDelegationKey}, delegatorAccAddr...)
	store := ctx.KVStore(k.storeKey)
	sharesStore := prefix.NewStore(store, agentUnbondingDelegationKey)

	unbondDelegations, pageRes, err := query.GenericFilteredPaginate(k.cdc, sharesStore, req.Pagination, func(key []byte, delegation *types.UnbondingDelegation) (*types.UnbondingDelegation, error) {
		return delegation, nil
	}, func() *types.UnbondingDelegation {
		return &types.UnbondingDelegation{}
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var unbondings []types.UnbondingDelegation
	for _, d := range unbondDelegations {
		unbondings = append(unbondings, *d)
	}

	return &types.QueryDelegatorUnbondingDelegationsResponse{
		UnbondingDelegations: unbondings,
		Pagination:           pageRes,
	}, nil
}

// OperatorUnbondingDelegations implements types.QueryServer.
func (k Querier) OperatorUnbondingDelegations(goCtx context.Context, req *types.QueryOperatorUnbondingDelegationsRequest) (*types.QueryOperatorUnbondingDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	operatorAccAddr, err := sdk.AccAddressFromBech32(req.OperatorAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	var unbondingDelegations []types.UnbondingDelegation
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	unbondingDelegationStore := prefix.NewStore(store, []byte{types.UnbondingDelegationKey})

	pageRes, err := query.Paginate(unbondingDelegationStore, req.Pagination, func(key []byte, value []byte) error {
		var delegation types.UnbondingDelegation
		err := k.cdc.Unmarshal(value, &delegation)
		if err != nil {
			return err
		}

		if delegation.DelegatorAddress != operatorAccAddr.String() {
			return nil
		}

		unbondingDelegations = append(unbondingDelegations, delegation)

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryOperatorUnbondingDelegationsResponse{
		UnbondingDelegations: unbondingDelegations,
		Pagination:           pageRes,
	}, nil
}

// UnbondingDelegation implements types.QueryServer.
func (k Querier) UnbondingDelegation(goCtx context.Context, req *types.QueryUnbondingDelegationRequest) (*types.QueryUnbondingDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	operatorAccAddr, err := sdk.AccAddressFromBech32(req.OperatorAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "agent address is invalid, %s", err)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	unbondingDelegation, found := k.GetUnbondingDelegation(ctx, delegatorAccAddr, operatorAccAddr)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "unbonding delegation not found")
	}

	return &types.QueryUnbondingDelegationResponse{
		Unbond: unbondingDelegation,
	}, nil
}
