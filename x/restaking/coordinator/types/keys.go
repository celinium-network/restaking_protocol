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

	WithdrawRecordPrefix

	OperatorLastRewardPeriodPrefix

	OperatorHistoricalRewardPrefix

	ConsumerTransferRewardPrefix

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

func ConsumerValidatorKey(chainID string, valAddr string) []byte {
	return append([]byte{ConsumerValidatorPrefix}, append(utils.BytesLengthPrefix([]byte(chainID)), address.MustLengthPrefix([]byte(valAddr))...)...)
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

func OperatorKey(operatorAccAddr sdk.AccAddress) []byte {
	return append([]byte{OperatorPrefix}, operatorAccAddr...)
}

func DelegationRecordKey(blockHeight uint64, operatorAccAddr sdk.AccAddress) []byte {
	bz := sdk.Uint64ToBigEndian(blockHeight)
	return append([]byte{DelegationRecordPrefix}, (append(operatorAccAddr, bz...))...)
}

func UndelegationRecordKey(blockHeight uint64, operatorAccAddr sdk.AccAddress) []byte {
	bz := sdk.Uint64ToBigEndian(blockHeight)
	return append([]byte{UndelegationRecordPrefix}, append(operatorAccAddr, bz...)...)
}

func OperatorWithdrawRecordKey(blockHeight uint64, operatorAccAddr sdk.AccAddress) []byte {
	bz := sdk.Uint64ToBigEndian(blockHeight)
	return append([]byte{WithdrawRecordPrefix}, append(operatorAccAddr, bz...)...)
}

func OperatorLastRewardPeriodKey(operatorAccAddr sdk.AccAddress) []byte {
	return append([]byte{OperatorLastRewardPeriodPrefix}, operatorAccAddr...)
}

func OperatorHistoricalRewardKey(period uint64, operatorAccAddr sdk.AccAddress) []byte {
	bz := sdk.Uint64ToBigEndian(period)
	return append([]byte{OperatorHistoricalRewardPrefix}, append(bz, operatorAccAddr...)...)
}

func ConsumerTransferRewardKey(destChannel, destPort string, sequence uint64) []byte {
	bz := utils.BytesLengthPrefix([]byte(destChannel))
	bz = append(bz, utils.BytesLengthPrefix([]byte(destChannel))...)
	bz = append(bz, sdk.Uint64ToBigEndian(sequence)...)

	return append([]byte{ConsumerTransferRewardPrefix}, bz...)
}

func OperatorSharesKey(ownerAccAddr, operatorAccAddr sdk.AccAddress) []byte {
	// TODO address string in key should has length as prefix ?
	return append([]byte{OperatorSharesPrefix}, append(ownerAccAddr, operatorAccAddr...)...)
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
