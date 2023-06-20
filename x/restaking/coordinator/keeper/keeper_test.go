package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	testutilkeeper "github.com/celinium-network/restaking_protocol/testutil/keeper"
	coordkeeper "github.com/celinium-network/restaking_protocol/x/restaking/coordinator/keeper"
	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

var (
	consumerChainIDs  = []string{"consumer-0", "consumer-1", "consumer-2"}
	consumerClientIDs = []string{"client-0", "client-1", "client-2"}
	consumerChannels  = []string{"channel-0", "channel-1", "channel-2"}
)

type KeeperTestSuite struct {
	suite.Suite

	ctx               sdk.Context
	coordinatorKeeper *coordkeeper.Keeper
	scopedKeeper      *testutilkeeper.MockScopedKeeper
	bankKeeper        *testutilkeeper.MockBankKeeper
	channelKeeper     *testutilkeeper.MockChannelKeeper
	portKeeper        *testutilkeeper.MockPortKeeper
	connectionKeeper  *testutilkeeper.MockConnectionKeeper
	clientKeeper      *testutilkeeper.MockClientKeeper
	transferKeeper    *testutilkeeper.MockIBCTransferKeeper
}

func (s *KeeperTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(s.T())
	s.scopedKeeper = testutilkeeper.NewMockScopedKeeper(ctrl)
	s.bankKeeper = testutilkeeper.NewMockBankKeeper(ctrl)
	s.channelKeeper = testutilkeeper.NewMockChannelKeeper(ctrl)
	s.portKeeper = testutilkeeper.NewMockPortKeeper(ctrl)
	s.connectionKeeper = testutilkeeper.NewMockConnectionKeeper(ctrl)
	s.clientKeeper = testutilkeeper.NewMockClientKeeper(ctrl)
	s.transferKeeper = testutilkeeper.NewMockIBCTransferKeeper(ctrl)

	keeper := coordkeeper.NewKeeper(
		encCfg.Codec,
		key,
		s.scopedKeeper,
		s.bankKeeper,
		s.channelKeeper,
		s.portKeeper,
		s.connectionKeeper,
		s.clientKeeper,
		s.transferKeeper,
	)

	s.ctx = ctx
	s.coordinatorKeeper = &keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) mockOperator() *types.Operator {
	ctx, keeper := s.ctx, s.coordinatorKeeper

	var consumerValidatorAddresses []string
	for i := 0; i < len(consumerChainIDs); i++ {
		keeper.SetConsumerClientID(ctx, consumerChainIDs[i], consumerClientIDs[i])
		keeper.SetConsumerClientIDToChannel(ctx, consumerClientIDs[i], consumerChannels[i])

		valAddr := sdk.ValAddress(PKs[i].Address().Bytes()).String()
		consumerValidatorAddresses = append(consumerValidatorAddresses, valAddr)

		keeper.SetConsumerValidator(ctx, consumerClientIDs[i], types.ConsumerValidator{
			Address: valAddr,
		})

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
				ChainID:          consumerChainIDs[0],
				ValidatorAddress: consumerValidatorAddresses[0],
			},
			{
				ChainID:          consumerChainIDs[1],
				ValidatorAddress: consumerValidatorAddresses[1],
			},
			{
				ChainID:          consumerChainIDs[2],
				ValidatorAddress: consumerValidatorAddresses[2],
			},
		},
		Owner: userAddr.String(),
	}

	keeper.SetOperator(ctx, &operator)

	return &operator
}
