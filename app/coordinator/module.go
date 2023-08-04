package coordinator

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	DefaultUnbondingTime = stakingtypes.DefaultUnbondingTime
	TestNetVotingPeriod  = time.Minute * 10
)

// stakingModule wraps the x/staking module in order to overwrite specific
// ModuleManager APIs.
type stakingModule struct {
	staking.AppModuleBasic
}

// DefaultGenesis returns custom x/staking module genesis state.
func (stakingModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	stakingParams := stakingtypes.DefaultParams()
	stakingParams.BondDenom = DefaultBondDenom
	stakingParams.UnbondingTime = DefaultUnbondingTime

	return cdc.MustMarshalJSON(&stakingtypes.GenesisState{
		Params: stakingParams,
	})
}

func newGovModule() govModule {
	return govModule{gov.NewAppModuleBasic(getGovProposalHandlers())}
}

// govModule is a custom wrapper around the x/gov module's AppModuleBasic
// implementation to provide custom default genesis state.
type govModule struct {
	gov.AppModuleBasic
}

// DefaultGenesis returns custom x/gov module genesis state.
func (govModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genState := govtypes.DefaultGenesisState()
	genState.Params.MinDeposit = sdk.NewCoins(sdk.NewCoin(DefaultBondDenom, sdk.NewInt(10000000)))

	// TODO: How set it in testnet-v0
	genState.Params.VotingPeriod = &TestNetVotingPeriod

	return cdc.MustMarshalJSON(genState)
}

type crisisModule struct {
	crisis.AppModuleBasic
}

// DefaultGenesis returns custom x/crisis module genesis state.
func (crisisModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&crisistypes.GenesisState{
		ConstantFee: sdk.NewCoin(DefaultBondDenom, sdk.NewInt(1000)),
	})
}

type mintModule struct {
	mint.AppModuleBasic
}

// DefaultGenesis returns custom x/mint module genesis state.
func (mintModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genState := minttypes.DefaultGenesisState()
	genState.Params.MintDenom = DefaultBondDenom

	return cdc.MustMarshalJSON(genState)
}
