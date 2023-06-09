package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	testutilkeeper "github.com/celinium-network/restaking_protocol/testutil/keeper"
	coordkeeper "github.com/celinium-network/restaking_protocol/x/restaking/coordinator/keeper"
	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
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
