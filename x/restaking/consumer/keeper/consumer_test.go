package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"

	cryptoutil "github.com/celinium-network/restaking_protocol/testutil/crypto"
	multistakingtypes "github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

var PKs = simtestutil.CreateTestPubKeys(500)

func (s *KeeperTestSuite) TestHandleRestakingDelegationPacket() {
	validatorPk, err := cryptoutil.CreateTmProtoPublicKey()
	s.Require().NoError(err)

	operatorAccounts := simtestutil.CreateIncrementalAccounts(1)
	valAddr := sdk.ValAddress(PKs[0].Address().Bytes()).String()

	restakingDelegation := restaking.DelegationPacket{
		OperatorAddress:  operatorAccounts[0].String(),
		ValidatorAddress: valAddr,
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
		valAddr)

	sdkPk, err := cryptocodec.FromTmProtoPublicKey(validatorPk)
	s.Require().NoError(err)
	valAddress := sdk.ValAddress(sdkPk.Address().Bytes())

	s.stakingKeeper.EXPECT().GetValidator(gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{
		OperatorAddress: valAddress.String(),
	}, true)

	s.bankKeeper.EXPECT().MintCoins(gomock.Any(), types.ModuleName, sdk.Coins{restakingDelegation.Balance})
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(), types.ModuleName, localOperator, sdk.Coins{restakingDelegation.Balance})

	s.multiStakingKeeper.EXPECT().MTStakingDelegate(gomock.Any(), multistakingtypes.MsgMTStakingDelegate{
		DelegatorAddress: localOperator.String(),
		ValidatorAddress: valAddress.String(),
		Amount:           restakingDelegation.Balance,
	})

	s.keeper.HandleRestakingDelegationPacket(s.ctx, packet, &restakingDelegation)
}

func (s *KeeperTestSuite) TestHandleRestakingUndelegationPacket() {
	operatorAccounts := simtestutil.CreateIncrementalAccounts(1)
	valAddr := sdk.ValAddress(PKs[0].Address().Bytes()).String()

	restakingUndelegation := restaking.UndelegationPacket{
		OperatorAddress:  operatorAccounts[0].String(),
		ValidatorAddress: valAddr,
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
		valAddr,
	)

	valAddress := sdk.ValAddress(PKs[0].Address().Bytes())

	s.stakingKeeper.EXPECT().GetValidator(gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{
		OperatorAddress: valAddress.String(),
	}, true)

	s.multiStakingKeeper.EXPECT().Unbond(gomock.Any(), localOperator, valAddress, restakingUndelegation.Balance)

	err := s.keeper.HandleRestakingUndelegationPacket(s.ctx, packet, &restakingUndelegation)
	s.Require().NoError(err)
}
