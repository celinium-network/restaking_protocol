package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

// Agent implements types.QueryServer.
func (k Querier) Agent(goCtx context.Context, req *types.QueryAgentRequest) (*types.QueryAgentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validator address is invalid, %s", err)
	}

	agentAccAddrBz, found := k.GetMTStakingAgentAddressByDenomAndVal(ctx, req.Denom, valAddr)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "agent not found by denom:%s, validator:%s", req.Denom, req.ValidatorAddr)
	}

	agent, found := k.GetMTStakingAgentByAddress(ctx, sdk.AccAddress(agentAccAddrBz))
	if !found {
		return nil, status.Errorf(codes.Internal, "agent not found by address,denom:%s, validator:%s", req.Denom, req.ValidatorAddr)
	}

	return &types.QueryAgentResponse{
		Agent: *agent,
	}, nil
}

// AgentDelegations implements types.QueryServer.
func (k Querier) AgentDelegations(goCtx context.Context, req *types.QueryAgentDelegationsRequest) (*types.QueryAgentDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	agentAccAddr, err := sdk.AccAddressFromBech32(req.AgentAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "agent address is invalid, %s", err)
	}

	agentAddBz := address.MustLengthPrefix([]byte(agentAccAddr))
	// TODO replace it by copy?
	agentSharesKey := append(types.DelegationPrefix, agentAddBz...) //nolint:gocritic
	store := ctx.KVStore(k.storeKey)
	sharesStore := prefix.NewStore(store, agentSharesKey)

	delegations, pageRes, err := query.GenericFilteredPaginate(k.cdc, sharesStore, req.Pagination, func(key []byte, delegation *types.MTStakingDelegation) (*types.MTStakingDelegation, error) {
		return delegation, nil
	}, func() *types.MTStakingDelegation {
		return &types.MTStakingDelegation{}
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var dels []types.MTStakingDelegation
	for _, d := range delegations {
		dels = append(dels, *d)
	}

	return &types.QueryAgentDelegationsResponse{
		Delegations: dels,
		Pagination:  pageRes,
	}, nil
}

// AgentUnbondingDelegations implements types.QueryServer.
func (k Querier) AgentUnbondingDelegations(goCtx context.Context, req *types.AgentUnbondingDelegationsRequest) (*types.AgentUnbondingDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	agentAccAddr, err := sdk.AccAddressFromBech32(req.AgentAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "agent address is invalid, %s", err)
	}

	agentAddBz := address.MustLengthPrefix([]byte(agentAccAddr))
	// TODO replace it by copy?
	agentUnbondingDelegationKey := append(types.UnbondingPrefix, agentAddBz...) //nolint:gocritic
	store := ctx.KVStore(k.storeKey)
	sharesStore := prefix.NewStore(store, agentUnbondingDelegationKey)

	unbondDelegations, pageRes, err := query.GenericFilteredPaginate(k.cdc, sharesStore, req.Pagination, func(key []byte, delegation *types.MTStakingUnbondingDelegation) (*types.MTStakingUnbondingDelegation, error) {
		return delegation, nil
	}, func() *types.MTStakingUnbondingDelegation {
		return &types.MTStakingUnbondingDelegation{}
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var unbonding []types.MTStakingUnbondingDelegation
	for _, d := range unbondDelegations {
		unbonding = append(unbonding, *d)
	}

	return &types.AgentUnbondingDelegationsResponse{
		Unbondings: unbonding,
		Pagination: pageRes,
	}, nil
}

// Agents implements types.QueryServer.
func (k Querier) Agents(goCtx context.Context, req *types.QueryAgentsRequest) (*types.QueryAgentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	sharesStore := prefix.NewStore(store, types.AgentPrefix)

	agents, pageRes, err := query.GenericFilteredPaginate(k.cdc, sharesStore, req.Pagination, func(key []byte, delegation *types.MTStakingAgent) (*types.MTStakingAgent, error) {
		return delegation, nil
	}, func() *types.MTStakingAgent {
		return &types.MTStakingAgent{}
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var as []types.MTStakingAgent
	for _, a := range agents {
		as = append(as, *a)
	}

	return &types.QueryAgentsResponse{
		Agents:     as,
		Pagination: pageRes,
	}, nil
}

// Delegation implements types.QueryServer.
func (k Querier) Delegation(goCtx context.Context, req *types.DelegationRequest) (*types.DelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	agentAccAddr, err := sdk.AccAddressFromBech32(req.AgentAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	delegation, found := k.GetDelegation(ctx, agentAccAddr, delegatorAccAddr)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "delegation not found")
	}

	agent, found := k.GetMTStakingAgentByAddress(ctx, agentAccAddr)
	if !found {
		return nil, status.Errorf(codes.Internal, "agent not found")
	}

	delegationAmt := delegation.Shares.Mul(agent.StakedAmount).Quo(agent.Shares)

	return &types.DelegationResponse{
		Delegation: *delegation,
		Balance:    sdk.NewCoin(agent.StakeDenom, delegationAmt),
	}, nil
}

// DelegatorAgents implements types.QueryServer.
func (k Querier) DelegatorAgents(goCtx context.Context, req *types.DelegatorAgentsRequest) (*types.DelegatorAgentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	var agents []types.MTStakingAgent
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, types.DelegationPrefix)

	pageRes, err := query.Paginate(delegationStore, req.Pagination, func(key []byte, value []byte) error {
		var delegation types.MTStakingDelegation
		err := k.cdc.Unmarshal(value, &delegation)
		if err != nil {
			return err
		}

		if delegation.DelegatorAddress != string(delegatorAccAddr) {
			return nil
		}

		agentAccAddr := sdk.AccAddress(delegation.AgentAddress)
		agent, found := k.GetMTStakingAgentByAddress(ctx, agentAccAddr)
		if !found {
			return types.ErrNotExistedAgent
		}

		agents = append(agents, *agent)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.DelegatorAgentsResponse{
		Agents:     agents,
		Pagination: pageRes,
	}, nil
}

// DelegatorDelegations implements types.QueryServer.
func (k Querier) DelegatorDelegations(goCtx context.Context, req *types.DelegatorDelegationsRequest) (*types.DelegatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	var delegations []types.MTStakingDelegation
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	delegationStore := prefix.NewStore(store, types.DelegationPrefix)

	pageRes, err := query.Paginate(delegationStore, req.Pagination, func(key []byte, value []byte) error {
		var delegation types.MTStakingDelegation
		err := k.cdc.Unmarshal(value, &delegation)
		if err != nil {
			return err
		}

		if delegation.DelegatorAddress != string(delegatorAccAddr) {
			return nil
		}

		delegations = append(delegations, delegation)

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.DelegatorDelegationsResponse{
		Delegations: delegations,
		Pagination:  pageRes,
	}, nil
}

// DelegatorUnbondingDelegations implements types.QueryServer.
func (k Querier) DelegatorUnbondingDelegations(goCtx context.Context, req *types.DelegatorUnbondingDelegationsRequest) (*types.DelegatorUnbondingDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	var unbondingDelegations []types.MTStakingUnbondingDelegation
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	unbondingDelegationStore := prefix.NewStore(store, types.UnbondingPrefix)

	pageRes, err := query.Paginate(unbondingDelegationStore, req.Pagination, func(key []byte, value []byte) error {
		var delegation types.MTStakingUnbondingDelegation
		err := k.cdc.Unmarshal(value, &delegation)
		if err != nil {
			return err
		}

		if delegation.DelegatorAddress != string(delegatorAccAddr) {
			return nil
		}

		unbondingDelegations = append(unbondingDelegations, delegation)

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.DelegatorUnbondingDelegationsResponse{
		Unbond:     unbondingDelegations,
		Pagination: pageRes,
	}, nil
}

// DenomAgents implements types.QueryServer.
func (k Querier) DenomAgents(goCtx context.Context, req *types.QueryDenomAgentsRequest) (*types.QueryDenomAgentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var agents []types.MTStakingAgent
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	agentsStore := prefix.NewStore(store, types.AgentPrefix)

	pageRes, err := query.Paginate(agentsStore, req.Pagination, func(key []byte, value []byte) error {
		var agent types.MTStakingAgent
		err := k.cdc.Unmarshal(value, &agent)
		if err != nil {
			return err
		}

		if agent.StakeDenom != req.Denom {
			return nil
		}

		agents = append(agents, agent)

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryDenomAgentsResponse{
		Agents:     agents,
		Pagination: pageRes,
	}, nil
}

// UnbondingDelegation implements types.QueryServer.
func (k Querier) UnbondingDelegation(goCtx context.Context, req *types.UnbondingDelegationRequest) (*types.UnbondingDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	delegatorAccAddr, err := sdk.AccAddressFromBech32(req.DelegatorAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "delegator address is invalid, %s", err)
	}

	agentAccAddr, err := sdk.AccAddressFromBech32(req.AgentAddr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "agent address is invalid, %s", err)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	unbondingDelegation, found := k.GetMTStakingUnbonding(ctx, agentAccAddr, delegatorAccAddr)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "unbonding delegation not found")
	}

	return &types.UnbondingDelegationResponse{
		Unbond: *unbondingDelegation,
	}, nil
}
