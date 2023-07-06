package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"

	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

var PKs = simtestutil.CreateTestPubKeys(500)

func (s *KeeperTestSuite) TestHandleRestakingDelegationPacket() {
	operatorAccounts := simtestutil.CreateIncrementalAccounts(1)
	valAddr := sdk.ValAddress(PKs[0].Address().Bytes())
	valAddress := valAddr.String()

	restakingDelegation := restaking.DelegationPacket{
		OperatorAddress:  operatorAccounts[0].String(),
		ValidatorAddress: valAddress,
		Balance:          sdk.NewCoin("restakingDenom", math.NewIntFromUint64(100000)),
	}

	restakingDelegationBz := s.codec.MustMarshal(&restakingDelegation)
	timeoutTimestamp := uint64(s.ctx.BlockTime().Add(time.Minute * 10).UnixNano())
	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         restaking.CoordinatorPortID,
		SourceChannel:      "channel-0",
		DestinationPort:    restaking.ConsumerPortID,
		DestinationChannel: "channel-0",
		Data:               restakingDelegationBz,
		TimeoutHeight:      clienttypes.Height{},
		TimeoutTimestamp:   timeoutTimestamp,
	}

	s.keeper.SetCoordinatorChannelID(s.ctx, packet.SourceChannel)

	localOperator := s.keeper.GetOrCreateOperatorLocalAddress(
		s.ctx,
		packet.SourceChannel,
		packet.SourcePort,
		restakingDelegation.OperatorAddress,
		valAddress)

	s.bankKeeper.EXPECT().MintCoins(gomock.Any(), types.ModuleName, sdk.Coins{restakingDelegation.Balance})
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(), types.ModuleName, localOperator, sdk.Coins{restakingDelegation.Balance})

	s.multiStakingKeeper.EXPECT().MTStakingDelegate(gomock.Any(), localOperator, valAddr, restakingDelegation.Balance)

	s.keeper.HandleRestakingDelegationPacket(s.ctx, packet, &restakingDelegation)
}

func (s *KeeperTestSuite) TestHandleRestakingUndelegationPacket() {
	operatorAccounts := simtestutil.CreateIncrementalAccounts(1)
	valAddr := sdk.ValAddress(PKs[0].Address().Bytes())
	valAddress := valAddr.String()

	restakingUndelegation := restaking.UndelegationPacket{
		OperatorAddress:  operatorAccounts[0].String(),
		ValidatorAddress: valAddress,
		Balance:          sdk.NewCoin("restakingDenom", math.NewIntFromUint64(100000)),
	}

	restakingDelegationBz := s.codec.MustMarshal(&restakingUndelegation)
	timeoutTimestamp := uint64(s.ctx.BlockTime().Add(time.Minute * 10).UnixNano())
	packet := channeltypes.Packet{
		Sequence:           1,
		SourcePort:         restaking.CoordinatorPortID,
		SourceChannel:      "channel-0",
		DestinationPort:    restaking.ConsumerPortID,
		DestinationChannel: "channel-0",
		Data:               restakingDelegationBz,
		TimeoutHeight:      clienttypes.Height{},
		TimeoutTimestamp:   timeoutTimestamp,
	}

	s.keeper.SetCoordinatorChannelID(s.ctx, packet.SourceChannel)

	localOperator := s.keeper.GetOrCreateOperatorLocalAddress(
		s.ctx,
		packet.SourceChannel,
		packet.SourcePort,
		restakingUndelegation.OperatorAddress,
		valAddress,
	)

	s.multiStakingKeeper.EXPECT().Unbond(gomock.Any(), localOperator, valAddr, restakingUndelegation.Balance)
	err := s.keeper.HandleRestakingUndelegationPacket(s.ctx, packet, &restakingUndelegation)
	s.Require().NoError(err)
}
