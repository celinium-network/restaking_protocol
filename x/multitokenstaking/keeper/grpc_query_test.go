package keeper_test

import (
	gocontext "context"
	"fmt"

	"github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestGRPCQueryAgent() {
	ctx, keeper, queryClient := s.ctx, s.mtStakingKeeper, s.queryClient

	valAddr := sdk.ValAddress(pks[0].Address())
	agentAccAddr := accounts[0]

	agent := types.MTStakingAgent{
		AgentAddress:     agentAccAddr.String(),
		StakeDenom:       mtStakingDenom,
		ValidatorAddress: valAddr.String(),
	}

	keeper.SetMTStakingAgent(ctx, agentAccAddr, &agent)
	keeper.SetMTStakingDenomAndValWithAgentAddress(ctx, agentAccAddr, mtStakingDenom, valAddr)

	var req *types.QueryAgentRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"",
			func() {
				req = &types.QueryAgentRequest{}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryAgentRequest{
					Denom:         mtStakingDenom,
					ValidatorAddr: valAddr.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.Agent(gocontext.Background(), req)
			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(agent.AgentAddress, res.Agent.AgentAddress)
				s.Require().Equal(agent.ValidatorAddress, res.Agent.ValidatorAddress)
				s.Require().Equal(agent.StakeDenom, res.Agent.StakeDenom)
			} else {
				s.Require().Error(err)
				s.Require().Nil(res)
			}
		})
	}
}
