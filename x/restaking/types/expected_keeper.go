package types

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	transfertype "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	multistakingtypes "github.com/celinium-network/restaking_protocol/x/multistaking/types"
)

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (uint64, error)
	WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error
	ChanCloseInit(ctx sdk.Context, portID, channelID string, chanCap *capabilitytypes.Capability) error
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
}

// ConnectionKeeper defines the expected IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (conntypes.ConnectionEnd, bool)
}

// ClientKeeper defines the expected IBC client keeper
type ClientKeeper interface {
	CreateClient(ctx sdk.Context, clientState ibcexported.ClientState, consensusState ibcexported.ConsensusState) (string, error)
	GetClientState(ctx sdk.Context, clientID string) (ibcexported.ClientState, bool)
	GetLatestClientConsensusState(ctx sdk.Context, clientID string) (ibcexported.ConsensusState, bool)
	GetSelfConsensusState(ctx sdk.Context, height ibcexported.Height) (ibcexported.ConsensusState, error)
}

type ScopedKeeper interface {
	GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool)
	AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool
	ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error
}

type StakingKeeper interface {
	GetParams(ctx sdk.Context) stakingtypes.Params
	GetValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate
	UnbondingCanComplete(ctx sdk.Context, id uint64) error
	UnbondingTime(ctx sdk.Context) time.Duration
	GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, found bool)
	GetLastValidatorPower(ctx sdk.Context, operator sdk.ValAddress) (power int64)
	// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
	Jail(sdk.Context, sdk.ConsAddress) // jail a validator
	Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec) math.Int
	Unjail(ctx sdk.Context, addr sdk.ConsAddress)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	IterateLastValidatorPowers(ctx sdk.Context, cb func(addr sdk.ValAddress, power int64) (stop bool))
	PowerReduction(ctx sdk.Context) math.Int
	PutUnbondingOnHold(ctx sdk.Context, id uint64) error
	IterateValidators(ctx sdk.Context, f func(index int64, validator stakingtypes.ValidatorI) (stop bool))
	Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI
	IsValidatorJailed(ctx sdk.Context, addr sdk.ConsAddress) bool
	ValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) stakingtypes.ValidatorI
	Delegation(ctx sdk.Context, addr sdk.AccAddress, valAddr sdk.ValAddress) stakingtypes.DelegationI
	MaxValidators(ctx sdk.Context) uint32
	GetLastTotalPower(ctx sdk.Context) math.Int
	GetLastValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	BondDenom(ctx sdk.Context) (res string)
}

type EvidenceKeeper interface {
	HandleEquivocationEvidence(ctx sdk.Context, evidence *evidencetypes.Equivocation)
}

// SlashingKeeper defines the contract expected to perform ccv slashing
type SlashingKeeper interface {
	JailUntil(sdk.Context, sdk.ConsAddress, time.Time) // called from provider keeper only
	GetValidatorSigningInfo(ctx sdk.Context, address sdk.ConsAddress) (info slashingtypes.ValidatorSigningInfo, found bool)
	DowntimeJailDuration(sdk.Context) time.Duration
	SlashFractionDowntime(sdk.Context) sdk.Dec
	SlashFractionDoubleSign(ctx sdk.Context) (res sdk.Dec)
	Tombstone(sdk.Context, sdk.ConsAddress)
	IsTombstoned(sdk.Context, sdk.ConsAddress) bool
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
}

// AccountKeeper defines the expected account keeper used for simulations
type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, name string) auth.ModuleAccountI
}

type IBCTransferKeeper interface {
	Transfer(
		context.Context,
		*transfertype.MsgTransfer,
	) (*transfertype.MsgTransferResponse, error)
}

type MultiStakingKeeper interface {
	MultiStakingDelegate(ctx sdk.Context, msg multistakingtypes.MsgMultiStakingDelegate) error

	Unbond(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, token sdk.Coin) (math.Int, error)
}
