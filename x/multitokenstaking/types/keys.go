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

	// Prefix for key which used in `{validator_address + denom } => MTStakingAgent's ID`
	MTStakingAgentAddressPrefix = []byte{0x21}

	// Prefix for Key which used in `agent_address => MTStakingAgent`
	AgentPrefix = []byte{0x22}

	// Prefix for key which used in `{agent_address + delegator_address} => MTStakingUnbonding`
	UnbondingPrefix = []byte{0x31}

	UnbondingQueueKey = []byte{0x32}

	// Prefix for key which used in `{agent_address + delegator_address} => shares_amount`
	DelegationPrefix = []byte{0x41}

	// Prefix for key used in store total minted
	MintedKey = []byte{0x51}

	// Prefix for key used in store EquivalentMultiplierRecord
	TokenMultiplierPrefix = []byte{0x61}

	// Prefix for key used in store blockHeight of delegator withdraw reward
	WithdrawRewardPrefix = []byte{0x71}
)

func GetMTStakingAgentAddressKey(denom string, valAddr sdk.ValAddress) []byte {
	denomBz := utils.BytesLengthPrefix([]byte(denom))

	prefixLen := len(MTStakingAgentAddressPrefix)
	denomBzLen := len(denomBz)
	valAddrBzLen := len(valAddr)

	bz := make([]byte, prefixLen+denomBzLen+valAddrBzLen)

	copy(bz[:prefixLen], MTStakingAgentAddressPrefix)
	copy(bz[prefixLen:prefixLen+valAddrBzLen], valAddr)
	copy(bz[prefixLen+valAddrBzLen:], denomBz)

	return bz
}

func GetMTStakingAgentKey(agentAddr sdk.AccAddress) []byte {
	return append(AgentPrefix, agentAddr...)
}

func GetMTStakingDelegationKey(agentAddr, delegator sdk.AccAddress) []byte {
	agentAddBz := address.MustLengthPrefix([]byte(agentAddr))
	delegatorBz := address.MustLengthPrefix(delegator)
	prefixLen := len(DelegationPrefix)

	bz := make([]byte, prefixLen+len(agentAddBz)+len(delegatorBz))
	copy(bz[:prefixLen], DelegationPrefix)
	copy(bz[prefixLen:prefixLen+len(agentAddBz)], agentAddBz)
	copy(bz[prefixLen+len(agentAddBz):], delegatorBz)

	return bz
}

func GetMTStakingUnbondingKey(agentAddr, delegator sdk.AccAddress) []byte {
	agentAddBz := address.MustLengthPrefix(agentAddr)
	delegatorBz := address.MustLengthPrefix(delegator)
	prefixLen := len(UnbondingPrefix)

	bz := make([]byte, prefixLen+len(agentAddBz)+len(delegatorBz))

	copy(bz[:prefixLen], UnbondingPrefix)
	copy(bz[prefixLen:prefixLen+len(agentAddBz)], agentAddBz)
	copy(bz[prefixLen+len(agentAddBz):], delegatorBz)

	return bz
}

func GetMTStakingUnbondingByAgentIndexKey(agentAddr sdk.AccAddress) []byte {
	agentAddBz := address.MustLengthPrefix(agentAddr)
	prefixLen := len(UnbondingPrefix)

	bz := make([]byte, prefixLen+len(agentAddBz))

	copy(bz[:prefixLen], UnbondingPrefix)
	copy(bz[prefixLen:prefixLen+len(agentAddBz)], agentAddBz)

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

func GetMTWithdrawRewardHeightKey(delegator, agent sdk.AccAddress) []byte {
	dbz := address.MustLengthPrefix(delegator)
	abz := address.MustLengthPrefix(agent)
	return append(WithdrawRewardPrefix, append(dbz, abz...)...)
}
