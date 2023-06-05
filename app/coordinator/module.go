package coordinator

import (
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// bankModule defines a custom wrapper around the x/bank module's AppModuleBasic
// implementation to provide custom default genesis state.

var DefaultUnbondingTime = stakingtypes.DefaultUnbondingTime

// type bankModule struct {
// 	bank.AppModuleBasic
// }

// // DefaultGenesis returns custom x/bank module genesis state.
// func (bankModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
// 	genState := banktypes.DefaultGenesisState()

// 	return cdc.MustMarshalJSON(genState)
// }

// // stakingModule wraps the x/staking module in order to overwrite specific
// // ModuleManager APIs.
// type stakingModule struct {
// 	staking.AppModuleBasic
// }

// // DefaultGenesis returns custom x/staking module genesis state.
// func (stakingModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
// 	stakingParams := stakingtypes.DefaultParams()
// 	stakingParams.BondDenom = params.DefaultBondDenom
// 	stakingParams.UnbondingTime = DefaultUnbondingTime

// 	return cdc.MustMarshalJSON(&stakingtypes.GenesisState{
// 		Params: stakingParams,
// 	})
// }

// type crisisModule struct {
// 	crisis.AppModuleBasic
// }

// // DefaultGenesis returns custom x/crisis module genesis state.
// func (crisisModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
// 	return cdc.MustMarshalJSON(&crisistypes.GenesisState{
// 		ConstantFee: sdk.NewCoin(params.DefaultBondDenom, sdk.NewInt(1000)),
// 	})
// }

// type mintModule struct {
// 	mint.AppModuleBasic
// }

// // DefaultGenesis returns custom x/mint module genesis state.
// func (mintModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
// 	genState := minttypes.DefaultGenesisState()
// 	genState.Params.MintDenom = params.DefaultBondDenom

// 	return cdc.MustMarshalJSON(genState)
// }
