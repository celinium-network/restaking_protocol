package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	testutilkeeper "github.com/celinium-network/restaking_protocol/testutil/keeper/multitokenstaking"
	mtstakingkeeper "github.com/celinium-network/restaking_protocol/x/multitokenstaking/keeper"
	mtstakingtypes "github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx             sdk.Context
	codec           codec.Codec
	mtStakingKeeper *mtstakingkeeper.Keeper

	accountKeeper     *testutilkeeper.MockAccountKeeper
	bankKeeper        *testutilkeeper.MockBankKeeper
	epochKeeper       *testutilkeeper.MockEpochKeeper
	stakingKeeper     *testutilkeeper.MockStakingKeeper
	distributerKeeper *testutilkeeper.MockDistributionKeeper
}

func (s *KeeperTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(mtstakingtypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, sdk.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	ctrl := gomock.NewController(s.T())
	s.accountKeeper = testutilkeeper.NewMockAccountKeeper(ctrl)
	s.bankKeeper = testutilkeeper.NewMockBankKeeper(ctrl)
	s.epochKeeper = testutilkeeper.NewMockEpochKeeper(ctrl)
	s.stakingKeeper = testutilkeeper.NewMockStakingKeeper(ctrl)
	s.distributerKeeper = testutilkeeper.NewMockDistributionKeeper(ctrl)

	keeper := mtstakingkeeper.NewKeeper(
		encCfg.Codec,
		key,
		s.accountKeeper,
		s.bankKeeper,
		s.epochKeeper,
		s.stakingKeeper,
		s.distributerKeeper,
	)

	s.mtStakingKeeper = &keeper
	s.ctx = ctx
	s.codec = encCfg.Codec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
