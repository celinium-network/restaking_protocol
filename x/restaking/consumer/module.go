package consumer

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/keeper"
	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
)

var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModule           = AppModule{}
	_ module.BeginBlockAppModule = AppModule{}
	_ module.EndBlockAppModule   = AppModule{}
)

type AppModuleBasic struct {
	cdc codec.Codec
}

// DefaultGenesis implements module.AppModuleBasic
func (AppModuleBasic) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	return nil
}

// GetQueryCmd implements module.AppModuleBasic
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// GetTxCmd implements module.AppModuleBasic
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// Name implements module.AppModuleBasic
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterGRPCGatewayRoutes implements module.AppModuleBasic
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterInterfaces implements module.AppModuleBasic
func (AppModuleBasic) RegisterInterfaces(codectypes.InterfaceRegistry) {
}

// RegisterLegacyAminoCodec implements module.AppModuleBasic
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {
}

// ValidateGenesis implements module.AppModuleBasic
func (AppModuleBasic) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
	return nil
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// EndBlock implements module.EndBlockAppModule
func (AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}

// BeginBlock implements module.BeginBlockAppModule
func (AppModule) BeginBlock(sdk.Context, abci.RequestBeginBlock) {
}

// ExportGenesis implements module.AppModule
func (AppModule) ExportGenesis(sdk.Context, codec.JSONCodec) json.RawMessage {
	return nil
}

// InitGenesis implements module.AppModule
func (AppModule) InitGenesis(sdk.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

// ConsensusVersion implements module.AppModule
func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// LegacyQuerierHandler implements module.AppModule
func (AppModule) LegacyQuerierHandler(*codec.LegacyAmino) func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
	return nil
}

// QuerierRoute implements module.AppModule
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterInvariants implements module.AppModule
func (AppModule) RegisterInvariants(sdk.InvariantRegistry) {
}

// RegisterServices implements module.AppModule
func (AppModule) RegisterServices(module.Configurator) {
}
