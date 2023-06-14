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
	multistakingtypes "github.com/celinium-network/restaking_protocol/x/multistaking/types"
	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
	restaking "github.com/celinium-network/restaking_protocol/x/restaking/types"
)

func (s *KeeperTestSuite) TestHandleRestakingDelegationPacket() {
	validatorPk, err := cryptoutil.CreateTmProtoPublicKey()
	s.Require().NoError(err)

	operatorAccounts := simtestutil.CreateIncrementalAccounts(1)

	restakingDelegation := restaking.DelegationPacket{
		OperatorAddress: operatorAccounts[0].String(),
		ValidatorPk:     validatorPk,
		Amount:          sdk.NewCoin("restakingDenom", math.NewIntFromUint64(100000)),
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

	validatorPkBz := s.codec.MustMarshal(&validatorPk)
	localOperator := s.keeper.GetOrCreateOperatorLocalAddress(
		s.ctx,
		packet.SourceChannel,
		packet.SourcePort,
		restakingDelegation.OperatorAddress,
		validatorPkBz)

	sdkPk, err := cryptocodec.FromTmProtoPublicKey(validatorPk)
	s.Require().NoError(err)
	valAddress := sdk.ValAddress(sdkPk.Address().Bytes())

	s.stakingKeeper.EXPECT().GetValidatorByConsAddr(gomock.Any(), gomock.Any()).Return(stakingtypes.Validator{
		OperatorAddress: valAddress.String(),
	}, true)

	s.bankKeeper.EXPECT().MintCoins(gomock.Any(), types.ModuleName, sdk.Coins{restakingDelegation.Amount})
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
		gomock.Any(), types.ModuleName, localOperator, sdk.Coins{restakingDelegation.Amount})

	s.multiStakingKeeper.EXPECT().MultiStakingDelegate(gomock.Any(), multistakingtypes.MsgMultiStakingDelegate{
		DelegatorAddress: localOperator.String(),
		ValidatorAddress: valAddress.String(),
		Amount:           restakingDelegation.Amount,
	})

	s.keeper.HandleRestakingDelegationPacket(s.ctx, packet, &restakingDelegation)
}
