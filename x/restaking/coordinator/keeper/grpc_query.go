package keeper

import (
	"context"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

// Delegation implements types.QueryServer.
func (Querier) Delegation(context.Context, *types.QueryDelegationRequest) (*types.QueryDelegationResponse, error) {
	panic("unimplemented")
}

// DelegatorDelegations implements types.QueryServer.
func (Querier) DelegatorDelegations(context.Context, *types.QueryDelegatorDelegationsRequest) (*types.QueryDelegatorDelegationsResponse, error) {
	panic("unimplemented")
}

// DelegatorOperators implements types.QueryServer.
func (Querier) DelegatorOperators(context.Context, *types.QueryDelegatorOperatorsRequest) (*types.QueryDelegatorOperatorsResponse, error) {
	panic("unimplemented")
}

// DelegatorUnbondingDelegations implements types.QueryServer.
func (Querier) DelegatorUnbondingDelegations(context.Context, *types.QueryDelegatorUnbondingDelegationsRequest) (*types.QueryDelegatorUnbondingDelegationsResponse, error) {
	panic("unimplemented")
}

// Operator implements types.QueryServer.
func (Querier) Operator(context.Context, *types.QueryOperatorRequest) (*types.QueryOperatorResponse, error) {
	panic("unimplemented")
}

// OperatorDelegations implements types.QueryServer.
func (Querier) OperatorDelegations(context.Context, *types.QueryOperatorDelegationsRequest) (*types.QueryOperatorDelegationsResponse, error) {
	panic("unimplemented")
}

// OperatorUnbondingDelegations implements types.QueryServer.
func (Querier) OperatorUnbondingDelegations(context.Context, *types.QueryOperatorUnbondingDelegationsRequest) (*types.QueryOperatorUnbondingDelegationsResponse, error) {
	panic("unimplemented")
}

// Operators implements types.QueryServer.
func (Querier) Operators(context.Context, *types.QueryOperatorsRequest) (*types.QueryOperatorsResponse, error) {
	panic("unimplemented")
}

// UnbondingDelegation implements types.QueryServer.
func (Querier) UnbondingDelegation(context.Context, *types.QueryUnbondingDelegationRequest) (*types.QueryUnbondingDelegationResponse, error) {
	panic("unimplemented")
}
