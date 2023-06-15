package integration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	rsconsumer "github.com/celinium-network/restaking_protocol/app/consumer"
	rscoordinator "github.com/celinium-network/restaking_protocol/app/coordinator"
	"github.com/celinium-network/restaking_protocol/app/params"

	restakingtypes "github.com/celinium-network/restaking_protocol/x/restaking/types"
	//
)

type SetupRestakingCoordinator func(t *testing.T) (
	coord *ibctesting.Coordinator,
	RestakingCoordinatorChain *ibctesting.TestChain,
	RestakingCoordinatorApp *rscoordinator.App,
)

type SetupRestakingConsumer func(*testing.T, *ibctesting.Coordinator, int) (
	*ibctesting.TestChain,
	*rsconsumer.App,
)

type IntegrationTestSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator

	setupRestakingCoordinator SetupRestakingCoordinator
	setupRestakingConsumer    SetupRestakingConsumer

	rsCoordinatorChain *ibctesting.TestChain
	rsCoordinatorApp   *rscoordinator.App

	rsConsumerChain *ibctesting.TestChain
	rsConsumerApp   *rsconsumer.App

	path         *ibctesting.Path
	transferPath *ibctesting.Path
}

func NewIntegrationTestSuite() *IntegrationTestSuite {
	suite := new(IntegrationTestSuite)

	suite.setupRestakingCoordinator = func(t *testing.T) (
		*ibctesting.Coordinator,
		*ibctesting.TestChain,
		*rscoordinator.App,
	) {
		t.Helper()

		ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
			db := dbm.NewMemDB()
			encCdc := rscoordinator.MakeEncodingConfig()
			appOptions := make(simtestutil.AppOptionsMap, 0)
			appOptions[flags.FlagHome] = rscoordinator.DefaultNodeHome
			appOptions[server.FlagInvCheckPeriod] = false

			app := rscoordinator.New(log.NewNopLogger(), db, nil, true, nil, "", 0, encCdc, appOptions)
			genesisState := rscoordinator.NewDefaultGenesisState(encCdc.Marshaler)
			return app, genesisState
		}
		coordinator := ibctesting.NewCoordinator(t, 0)

		rscoordinatorChain := ibctesting.NewTestChain(t, coordinator, "coordinator")

		coordinatorApp, ok := rscoordinatorChain.App.(*rscoordinator.App)
		if !ok {
			t.Fatal("coordinator chain has mismatch app")
		}

		ctx := rscoordinatorChain.GetContext()
		coordinatorApp.RestakingCoordinatorKeeper.InitGenesis(ctx, nil)

		return coordinator, rscoordinatorChain, coordinatorApp
	}

	suite.setupRestakingConsumer = func(t *testing.T, coordinator *ibctesting.Coordinator, chainIndex int) (
		*ibctesting.TestChain,
		*rsconsumer.App,
	) {
		ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
			db := dbm.NewMemDB()
			encCdc := rsconsumer.MakeEncodingConfig()
			appOptions := make(simtestutil.AppOptionsMap, 0)
			appOptions[flags.FlagHome] = rscoordinator.DefaultNodeHome
			appOptions[server.FlagInvCheckPeriod] = false

			app := rsconsumer.New(log.NewNopLogger(), db, nil, true, nil, "", 0, encCdc, appOptions)
			genesisState := rsconsumer.NewDefaultGenesisState(encCdc.Marshaler)
			return app, genesisState
		}

		chainID := ibctesting.GetChainID(chainIndex)

		consumerChian := ibctesting.NewTestChain(t, coordinator, chainID)
		consumerApp, ok := consumerChian.App.(*rsconsumer.App)
		if !ok {
			t.Fatal("consumer chain has mismatch app")
		}
		ctx := consumerChian.GetContext()
		consumerApp.RestakingConsumerKeeper.InitGenesis(ctx, nil)
		// TODO the params of coordinator and consumer maybe in different package
		consumerApp.MultiStakingKeeper.SetMultiStakingDenom(ctx, params.DefaultBondDenom)

		return consumerChian, consumerApp
	}

	return suite
}

func (s *IntegrationTestSuite) SetupTest() {
	s.coordinator, s.rsCoordinatorChain, s.rsCoordinatorApp = s.setupRestakingCoordinator(s.T())
	s.rsConsumerChain, s.rsConsumerApp = s.setupRestakingConsumer(s.T(), s.coordinator, 0)

	s.path = ibctesting.NewPath(s.rsConsumerChain, s.rsCoordinatorChain)
	s.path.EndpointA.ChannelConfig.PortID = restakingtypes.ConsumerPortID
	s.path.EndpointB.ChannelConfig.PortID = restakingtypes.CoordinatorPortID
	s.path.EndpointA.ChannelConfig.Version = restakingtypes.Version
	s.path.EndpointB.ChannelConfig.Version = restakingtypes.Version
	s.path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	s.path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED

	s.transferPath = ibctesting.NewPath(s.rsConsumerChain, s.rsCoordinatorChain)
	s.transferPath.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	s.transferPath.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	s.transferPath.EndpointA.ChannelConfig.Version = ibctransfertypes.Version
	s.transferPath.EndpointB.ChannelConfig.Version = ibctransfertypes.Version
}

func getCoordinatorApp(chain *ibctesting.TestChain) *rscoordinator.App {
	app := chain.App.(*rscoordinator.App)
	return app
}

func getConsumerApp(chain *ibctesting.TestChain) *rsconsumer.App {
	app := chain.App.(*rsconsumer.App)
	return app
}

// func copyConnectionAndClientToPath(path *ibctesting.Path, pathToCopy *ibctesting.Path) *ibctesting.Path {
// 	path.EndpointA.ClientID = pathToCopy.EndpointA.ClientID
// 	path.EndpointB.ClientID = pathToCopy.EndpointB.ClientID
// 	path.EndpointA.ConnectionID = pathToCopy.EndpointA.ConnectionID
// 	path.EndpointB.ConnectionID = pathToCopy.EndpointB.ConnectionID
// 	path.EndpointA.ClientConfig = pathToCopy.EndpointA.ClientConfig
// 	path.EndpointB.ClientConfig = pathToCopy.EndpointB.ClientConfig
// 	path.EndpointA.ConnectionConfig = pathToCopy.EndpointA.ConnectionConfig
// 	path.EndpointB.ConnectionConfig = pathToCopy.EndpointB.ConnectionConfig
// 	return path
// }
