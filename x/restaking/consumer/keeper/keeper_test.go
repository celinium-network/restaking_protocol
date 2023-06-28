package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	testutilkeeper "github.com/celinium-network/restaking_protocol/testutil/keeper/restaking"
	consumerkeeper "github.com/celinium-network/restaking_protocol/x/restaking/consumer/keeper"
	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	codec  codec.Codec
	keeper *consumerkeeper.Keeper

	scopedKeeper       *testutilkeeper.MockScopedKeeper
	channelKeeper      *testutilkeeper.MockChannelKeeper
	portKeeper         *testutilkeeper.MockPortKeeper
	connectionKeeper   *testutilkeeper.MockConnectionKeeper
	clientKeeper       *testutilkeeper.MockClientKeeper
	transferKeeper     *testutilkeeper.MockIBCTransferKeeper
	bankKeeper         *testutilkeeper.MockBankKeeper
	stakingKeeper      *testutilkeeper.MockStakingKeeper
	slashingKeeper     *testutilkeeper.MockSlashingKeeper
	authKeeper         *testutilkeeper.MockAccountKeeper
	multiStakingKeeper *testutilkeeper.MockMTStakingKeeper
}

func (s *KeeperTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(s.T())
	s.scopedKeeper = testutilkeeper.NewMockScopedKeeper(ctrl)
	s.channelKeeper = testutilkeeper.NewMockChannelKeeper(ctrl)
	s.portKeeper = testutilkeeper.NewMockPortKeeper(ctrl)
	s.connectionKeeper = testutilkeeper.NewMockConnectionKeeper(ctrl)
	s.clientKeeper = testutilkeeper.NewMockClientKeeper(ctrl)
	s.transferKeeper = testutilkeeper.NewMockIBCTransferKeeper(ctrl)
	s.bankKeeper = testutilkeeper.NewMockBankKeeper(ctrl)
	s.stakingKeeper = testutilkeeper.NewMockStakingKeeper(ctrl)
	s.slashingKeeper = testutilkeeper.NewMockSlashingKeeper(ctrl)
	s.authKeeper = testutilkeeper.NewMockAccountKeeper(ctrl)
	s.multiStakingKeeper = testutilkeeper.NewMockMTStakingKeeper(ctrl)

	keeper := consumerkeeper.NewKeeper(
		key,
		encCfg.Codec,
		s.scopedKeeper,
		s.channelKeeper,
		s.portKeeper,
		s.connectionKeeper,
		s.clientKeeper,
		s.transferKeeper,
		s.bankKeeper,
		s.stakingKeeper,
		s.slashingKeeper,
		s.authKeeper,
		s.multiStakingKeeper,
	)

	s.codec = encCfg.Codec
	s.ctx = ctx
	s.keeper = &keeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
