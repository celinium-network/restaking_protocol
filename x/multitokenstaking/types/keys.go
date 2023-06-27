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
	MTStakingDenomWhiteListKey = []byte{0x11}

	// Prefix for key which used in `{denom + validator_address} => MTStakingAgent's ID`
	MTStakingAgentIDPrefix = []byte{0x21}

	// Prefix for Key which used in `agent_address => MTStakingAgent`
	MTStakingAgentPrefix = []byte{0x22}

	// Prefix for key which used in `{agent_address + delegator_address} => MTStakingUnbonding`
	MTStakingUnbondingPrefix = []byte{0x31}

	MTStakingUnbondingQueueKey = []byte{0x32}

	// Prefix for key which used in `{agent_address + delegator_address} => shares_amount`
	MTStakingSharesPrefix = []byte{0x41}
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
	return append(MTStakingAgentPrefix, []byte(agentAddr)...)
}

func GetMTStakingSharesKey(agentAddr string, delegator string) []byte {
	agentAddBz := address.MustLengthPrefix([]byte(agentAddr))
	delegatorBz := utils.BytesLengthPrefix([]byte(delegator))
	prefixLen := len(MTStakingSharesPrefix)

	bz := make([]byte, prefixLen+len(agentAddBz)+len(delegatorBz))
	copy(bz[:prefixLen], MTStakingSharesPrefix)
	copy(bz[prefixLen:prefixLen+len(agentAddBz)], agentAddBz)
	copy(bz[prefixLen+len(agentAddBz):], delegatorBz)

	return bz
}

func GetMTStakingUnbondingKey(agentAddr string, delegator string) []byte {
	agentAddBz := address.MustLengthPrefix([]byte(agentAddr))
	delegatorBz := address.MustLengthPrefix([]byte(delegator))
	prefixLen := len(MTStakingUnbondingPrefix)

	bz := make([]byte, prefixLen+len(agentAddBz)+len(delegatorBz))

	copy(bz[:prefixLen], MTStakingUnbondingPrefix)
	copy(bz[prefixLen:prefixLen+len(agentAddBz)], agentAddBz)
	copy(bz[prefixLen+len(agentAddBz):], delegatorBz)

	return bz
}

func GetMTStakingUnbondingDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append(MTStakingUnbondingQueueKey, bz...)
}

func (ubd *MTStakingUnbonding) RemoveEntry(i int64) {
	ubd.Entries = append(ubd.Entries[:i], ubd.Entries[i+1:]...)
}

func (e MTStakingUnbondingEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}
