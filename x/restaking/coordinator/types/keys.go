package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	ConsumerValidatorUpdatesPrefix

	ConsumerRestakingTokens

	ConsumerRewardTokens

	OperatorPrefix

	OperatorIDByteKey

	PortByteKey

	DelegationRecordPrefix

	OperatorSharesPrefix
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

func ConsumerValidatorSetKey(chainID string) []byte {
	return append([]byte{ConsumerValidatorUpdatesPrefix}, []byte(chainID)...)
}

func ConsumerRestakingTokensKey(chainID string) []byte {
	return append([]byte{ConsumerRestakingTokens}, []byte(chainID)...)
}

func ConsumerRewardTokensKey(chainID string) []byte {
	return append([]byte{ConsumerRewardTokens}, []byte(chainID)...)
}

func OperatorKey(operatorAddr string) []byte {
	return append([]byte{OperatorPrefix}, []byte(operatorAddr)...)
}

func DelegationRecordKey(blockHeight uint64) []byte {
	sdk.Uint64ToBigEndian(blockHeight)
	return append([]byte{DelegationRecordPrefix}, sdk.Uint64ToBigEndian(blockHeight)...)
}

func OperatorSharesKey(ownerAddr, operatorAddr string) []byte {
	// TODO address string in key should has length as prefix ?
	return append([]byte{OperatorSharesPrefix}, []byte(ownerAddr+operatorAddr)...)
}
