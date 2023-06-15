package keeper_test

import (
	"github.com/golang/mock/gomock"

	"cosmossdk.io/math"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (s *KeeperTestSuite) TestProcessPendingOperatorDelegationRecord() {
	operator := s.mockOperator()
	ctx, keeper := s.ctx, s.coordinatorKeeper

	operatorDelegateRecord := types.OperatorDelegationRecord{
		OperatorAddress:  operator.OperatorAddress,
		DelegationAmount: math.NewIntFromUint64(1000000),
		Status:           0,
		IbcCallbackIds:   []string{},
	}

	keeper.SetOperatorDelegateRecord(ctx, uint64(ctx.BlockHeight()), &operatorDelegateRecord)
	cap := capabilitytypes.Capability{}

	for i, channel := range consumerChannels {
		s.scopedKeeper.EXPECT().GetCapability(
			gomock.Any(),
			host.ChannelCapabilityPath(restaking.CoordinatorPortID, channel)).
			Return(&cap, true)

		s.channelKeeper.EXPECT().SendPacket(
			gomock.Any(),
			gomock.Any(),
			restaking.CoordinatorPortID,
			channel,
			clienttypes.Height{},
			gomock.Any(), gomock.Any()).Return(uint64(i), nil)
	}

	keeper.ProcessPendingOperatorDelegationRecord(ctx)

	// check callback
	var ibcCallbackIDs []string
	for i, channel := range consumerChannels {
		callback, found := keeper.GetCallback(ctx, channel, restaking.CoordinatorPortID, uint64(i))
		ibcCallbackIDs = append(ibcCallbackIDs, string(types.IBCCallbackKey(channel, restaking.CoordinatorPortID, uint64(i))))
		s.Require().True(found)
		s.Require().Equal(callback.CallType, types.InterChainDelegateCall)
		s.Require().Equal(callback.Args, string(types.DelegationRecordKey(uint64(ctx.BlockHeight()), operator.OperatorAddress)))
	}

	// check delegateRecord
	processedOperatorDelegation, found := keeper.GetOperatorDelegateRecord(ctx, uint64(ctx.BlockHeight()), operator.OperatorAddress)
	s.Require().True(found)
	s.Require().True(processedOperatorDelegation.DelegationAmount.Equal(operatorDelegateRecord.DelegationAmount))

	s.Require().ElementsMatch(processedOperatorDelegation.IbcCallbackIds, ibcCallbackIDs)
}

func (s *KeeperTestSuite) TestProcessPendingOperatorUndelegationRecord() {
	operator := s.mockOperator()
	ctx, keeper := s.ctx, s.coordinatorKeeper

	operatorDelegateRecord := types.OperatorUndelegationRecord{
		OperatorAddress:    operator.OperatorAddress,
		UndelegationAmount: math.NewIntFromUint64(1000000),
		Status:             types.OpUndelegationRecordPending,
		IbcCallbackIds:     []string{},
		UnbondingEntryIds:  []uint64{1},
	}

	keeper.SetOperatorUndelegationRecord(ctx, uint64(ctx.BlockHeight()), &operatorDelegateRecord)
	cap := capabilitytypes.Capability{}

	for i, channel := range consumerChannels {
		s.scopedKeeper.EXPECT().GetCapability(
			gomock.Any(),
			host.ChannelCapabilityPath(restaking.CoordinatorPortID, channel)).
			Return(&cap, true)

		s.channelKeeper.EXPECT().SendPacket(
			gomock.Any(),
			gomock.Any(),
			restaking.CoordinatorPortID,
			channel,
			clienttypes.Height{},
			gomock.Any(), gomock.Any()).Return(uint64(i), nil)
	}
	keeper.ProcessPendingOperatorUndelegationRecord(ctx)

	var ibcCallbackIDs []string
	for i, channel := range consumerChannels {
		callback, found := keeper.GetCallback(ctx, channel, restaking.CoordinatorPortID, uint64(i))
		ibcCallbackIDs = append(ibcCallbackIDs, string(types.IBCCallbackKey(channel, restaking.CoordinatorPortID, uint64(i))))
		s.Require().True(found)
		s.Require().Equal(callback.CallType, types.InterChainUndelegateCall)
		s.Require().Equal(callback.Args, string(types.UndelegationRecordKey(uint64(ctx.BlockHeight()), operator.OperatorAddress)))
	}

	processedOUndelegationRecord, found := keeper.GetOperatorUndelegationRecord(ctx, uint64(ctx.BlockHeight()), operator.OperatorAddress)
	s.Require().True(found)
	s.Require().True(processedOUndelegationRecord.UndelegationAmount.Equal(operatorDelegateRecord.UndelegationAmount))
	s.Require().ElementsMatch(processedOUndelegationRecord.IbcCallbackIds, ibcCallbackIDs)
	s.Require().ElementsMatch(processedOUndelegationRecord.UnbondingEntryIds, []uint64{1})
}
