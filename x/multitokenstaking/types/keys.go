package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/celinium-network/restaking_protocol/utils"
)

const (
	// ModuleName is the name of the multistaking module
	ModuleName = "multistaking"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the multistaking module
	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	// Key for the denom white list which allow used for multistaking
	DenomWhiteListKey = []byte{0x11}

	// Prefix for key which used in `{denom + validator_address} => MTStakingAgent's ID`
	MTStakingAgentIDPrefix = []byte{0x21}

	// Prefix for Key which used in `agent_address => MTStakingAgent`
	AgentPrefix = []byte{0x22}

	// Prefix for key which used in `{agent_address + delegator_address} => MTStakingUnbonding`
	UnbondingPrefix = []byte{0x31}

	UnbondingQueueKey = []byte{0x32}

	// Prefix for key which used in `{agent_address + delegator_address} => shares_amount`
	SharesPrefix = []byte{0x41}

	// Prefix for key used in store total minted
	MintedKey = []byte{0x51}

	// Prefix for key used in store EquivalentMultiplierRecord
	TokenMultiplierPrefix = []byte{0x61}
)

func GetMTStakingAgentIDKey(denom, valAddr string) []byte {
	denomBz := utils.BytesLengthPrefix([]byte(denom))
	valAddrBz := utils.BytesLengthPrefix([]byte(valAddr))

	prefixLen := len(MTStakingAgentIDPrefix)
	denomBzLen := len(denomBz)
	valAddrBzLen := len(valAddrBz)

	bz := make([]byte, prefixLen+denomBzLen+valAddrBzLen)

	copy(bz[:prefixLen], MTStakingAgentIDPrefix)
	copy(bz[prefixLen:prefixLen+denomBzLen], denomBz)
	copy(bz[prefixLen+denomBzLen:], valAddrBz)

	return bz
}

func GetMTStakingAgentKey(agentAddr string) []byte {
	return append(AgentPrefix, []byte(agentAddr)...)
}

func GetMTStakingSharesKey(agentAddr string, delegator string) []byte {
	agentAddBz := address.MustLengthPrefix([]byte(agentAddr))
	delegatorBz := utils.BytesLengthPrefix([]byte(delegator))
	prefixLen := len(SharesPrefix)

	bz := make([]byte, prefixLen+len(agentAddBz)+len(delegatorBz))
	copy(bz[:prefixLen], SharesPrefix)
	copy(bz[prefixLen:prefixLen+len(agentAddBz)], agentAddBz)
	copy(bz[prefixLen+len(agentAddBz):], delegatorBz)

	return bz
}

func GetMTStakingUnbondingKey(agentAddr string, delegator string) []byte {
	agentAddBz := address.MustLengthPrefix([]byte(agentAddr))
	delegatorBz := address.MustLengthPrefix([]byte(delegator))
	prefixLen := len(UnbondingPrefix)

	bz := make([]byte, prefixLen+len(agentAddBz)+len(delegatorBz))

	copy(bz[:prefixLen], UnbondingPrefix)
	copy(bz[prefixLen:prefixLen+len(agentAddBz)], agentAddBz)
	copy(bz[prefixLen+len(agentAddBz):], delegatorBz)

	return bz
}

func GetMTStakingUnbondingDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(UnbondingQueueKey, bz...)
}

func GetMTStakingMintedKey() []byte {
	return MintedKey
}

func GetMTTokenMultiplierKey(denom string) []byte {
	return append(TokenMultiplierPrefix, denom...)
}
