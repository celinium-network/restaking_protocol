package keeper_test

import (
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func (s *KeeperTestSuite) TestGRPCQueryOperators() {
	operator0 := s.mockOperator()
	operator1 := *operator0
	operator2 := *operator0

	addrs := simtestutil.CreateIncrementalAccounts(10)
	operator1.OperatorAddress = addrs[2].String()
	operator2.OperatorAddress = addrs[3].String()

	s.coordinatorKeeper.SetOperator(s.ctx, addrs[2], &operator1)
	s.coordinatorKeeper.SetOperator(s.ctx, addrs[3], &operator2)

	res, err := s.queryClient.Operators(s.ctx, &types.QueryOperatorsRequest{
		Pagination: &query.PageRequest{},
	})

	s.Require().NoError(err)

	s.Require().Equal(res.Operators[0].OperatorAddress, operator0.OperatorAddress)
	s.Require().Equal(res.Operators[1].OperatorAddress, operator1.OperatorAddress)
	s.Require().Equal(res.Operators[2].OperatorAddress, operator2.OperatorAddress)
}

func (s *KeeperTestSuite) TestGRPCQueryOperator() {
	operator0 := s.mockOperator()
	operator1 := *operator0
	operator2 := *operator0

	addrs := simtestutil.CreateIncrementalAccounts(10)
	operator1.OperatorAddress = addrs[2].String()
	operator2.OperatorAddress = addrs[3].String()

	s.coordinatorKeeper.SetOperator(s.ctx, addrs[2], &operator1)
	s.coordinatorKeeper.SetOperator(s.ctx, addrs[3], &operator2)

	res, err := s.queryClient.Operator(s.ctx, &types.QueryOperatorRequest{
		OperatorAddress: operator1.OperatorAddress,
	})

	s.Require().NoError(err)

	s.Require().Equal(res.Operator.OperatorAddress, operator1.OperatorAddress)
}
