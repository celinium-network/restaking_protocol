package keeper_test

import (
	"github.com/golang/mock/gomock"

	abci "github.com/cometbft/cometbft/abci/types"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	cryptoutil "github.com/celinium-network/restaking_protocol/testutil/crypto"
	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (s *KeeperTestSuite) TestProcessPendingOperatorDelegationRecord() {
	ctx, keeper := s.ctx, s.coordinatorKeeper
	consumerChainIDs := []string{"consumer-0", "consumer-1", "consumer-2"}
	consumerClientIDs := []string{"client-0", "client-1", "client-2"}
	consumerChannels := []string{"channel-0", "channel-1", "channel-2"}

	var tmPubkeys []tmprotocrypto.PublicKey
	for i := 0; i < len(consumerChainIDs); i++ {
		keeper.SetConsumerClientID(ctx, consumerChainIDs[i], consumerClientIDs[i])
		keeper.SetConsumerClientIDToChannel(ctx, consumerClientIDs[i], consumerChannels[i])

		tmProtoPk, err := cryptoutil.CreateTmProtoPublicKey()
		s.Require().NoError(err)
		tmPubkeys = append(tmPubkeys, tmProtoPk)

		keeper.SetConsumerValidator(ctx, consumerClientIDs[i], []abci.ValidatorUpdate{{
			PubKey: tmProtoPk,
			Power:  1,
		}})

		keeper.SetConsumerRestakingToken(ctx, consumerClientIDs[i], []string{"stake"})
		keeper.SetConsumerRewardToken(ctx, consumerClientIDs[i], []string{"stake"})
	}

	addrs := simtestutil.CreateIncrementalAccounts(2)
	userAddr := addrs[0]
	operatorAddr := addrs[1]

	operator := types.Operator{
		RestakingDenom:  "stake",
		OperatorAddress: operatorAddr.String(),
		RestakedAmount:  math.ZeroInt(),
		Shares:          math.ZeroInt(),
		OperatedValidators: []types.OperatedValidator{
			{
				ChainID:     consumerChainIDs[0],
				ValidatorPk: tmPubkeys[0],
			},
			{
				ChainID:     consumerChainIDs[1],
				ValidatorPk: tmPubkeys[1],
			},
			{
				ChainID:     consumerChainIDs[2],
				ValidatorPk: tmPubkeys[2],
			},
		},
		Owner: userAddr.String(),
	}

	keeper.SetOperator(ctx, &operator)

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
