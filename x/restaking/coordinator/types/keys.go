package types

import (
	"encoding/binary"
	time "time"

	"github.com/celinium-network/restaking_protocol/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the restaking coordinator module
	ModuleName = "restakingCoordinator"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the restaking coordinator module
	RouterKey = ModuleName

	QuerierRoute = ModuleName

	StringListSplitter = "|"
)

const (
	ConsumerAdditionProposalPrefix byte = iota

	ConsumerClientIDPrefix

	ConsumerValidatorListPrefix

	ConsumerValidatorPrefix

	ConsumerRestakingTokens

	ConsumerRewardTokens

	ConsumerClientToChannelPrefix

	OperatorPrefix

	OperatorIDByteKey

	PortByteKey

	DelegationRecordPrefix

	UndelegationRecordPrefix

	UnbondingDelegationKey

	OperatorSharesPrefix

	IBCCallbackPrefix

	UnbondingIDKey

	UnbondingIndexKey

	UnbondingQueueKey
)

func PortKey() []byte {
	return []byte{PortByteKey}
}

func ConsumerAdditionProposalKey(chainID string) []byte {
	return append([]byte{ConsumerAdditionProposalPrefix}, []byte(chainID)...)
}

func ConsumerClientIDKey(chainID string) []byte {
	return append([]byte{ConsumerClientIDPrefix}, []byte(chainID)...)
}

func ConsumerValidatorListKey(chainID string) []byte {
	return append([]byte{ConsumerValidatorListPrefix}, []byte(chainID)...)
}

func ConsumerValidatorKey(chainID string, pk []byte) []byte {
	return append([]byte{ConsumerValidatorPrefix},
		append(utils.BytesLengthPrefix([]byte(chainID)), address.MustLengthPrefix(pk)...)...)
}

func ConsumerRestakingTokensKey(chainID string) []byte {
	return append([]byte{ConsumerRestakingTokens}, []byte(chainID)...)
}

func ConsumerRewardTokensKey(chainID string) []byte {
	return append([]byte{ConsumerRewardTokens}, []byte(chainID)...)
}

func ConsumerClientToChannelKey(clientID string) []byte {
	return append([]byte{ConsumerClientToChannelPrefix}, []byte(clientID)...)
}

func OperatorKey(operatorAddr string) []byte {
	return append([]byte{OperatorPrefix}, []byte(operatorAddr)...)
}

func DelegationRecordKey(blockHeight uint64, operatorAddr string) []byte {
	bz := sdk.Uint64ToBigEndian(blockHeight)
	return append([]byte{DelegationRecordPrefix}, []byte(operatorAddr+string(bz))...)
}

func UndelegationRecordKey(blockHeight uint64, operatorAddr string) []byte {
	bz := sdk.Uint64ToBigEndian(blockHeight)
	return append([]byte{UndelegationRecordPrefix}, []byte(operatorAddr+string(bz))...)
}

func OperatorSharesKey(ownerAddr, operatorAddr string) []byte {
	// TODO address string in key should has length as prefix ?
	return append([]byte{OperatorSharesPrefix}, []byte(ownerAddr+operatorAddr)...)
}

func IBCCallbackKey(channelID, portID string, seq uint64) []byte {
	bz := sdk.Uint64ToBigEndian(seq)
	return append([]byte{IBCCallbackPrefix}, []byte(channelID+portID+string(bz))...)
}

func GetUBDKey(delAddr sdk.AccAddress, opAddr sdk.AccAddress) []byte {
	return append(GetUBDsKey(delAddr.Bytes()), address.MustLengthPrefix(opAddr)...)
}

func GetUBDsKey(delAddr sdk.AccAddress) []byte {
	return append([]byte{UnbondingDelegationKey}, address.MustLengthPrefix(delAddr)...)
}

func GetUnbondingIndexKey(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return append([]byte{UnbondingIndexKey}, bz...)
}

func GetUnbondingDelegationTimeKey(timestamp time.Time) []byte {
	bz := sdk.FormatTimeBytes(timestamp)
	return append([]byte{UnbondingQueueKey}, bz...)
}
