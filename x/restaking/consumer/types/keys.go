package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the restaking coordinator module
	ModuleName = "restakingConsumer"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the restaking coordinator module
	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

const (
	ValidatorSetUpdateIDKey byte = iota

	ValidatorSetChangeSet

	PendingValidatorSlashListKey

	CoordinatorChannelID

	OperatorAddressPrefix

	NotifyUpdateValidators
)

func GetValidatorSetUpdateIDKey() []byte {
	return []byte{ValidatorSetUpdateIDKey}
}

func GetPendingValidatorChangeSetKey() []byte {
	return []byte{ValidatorSetChangeSet}
}

func GetCoordinatorChannelIDKey() []byte {
	return []byte{CoordinatorChannelID}
}

func OperatorAddressKey(operatorAddress sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	// key: byte{prefix}| byte{valAddrLen}| valAddr| operatorAddr
	return append([]byte{OperatorAddressPrefix}, append(address.MustLengthPrefix(valAddr), operatorAddress...)...)
}

func GetPendingConsumerSlashListKey() []byte {
	return []byte{PendingValidatorSlashListKey}
}

func NotifyUpdateValidatorKey() []byte {
	return []byte{NotifyUpdateValidators}
}

func ParseValidatorOperatorKey(key []byte) []byte {
	valAddrLen := key[1]
	prefixLen := valAddrLen + 2
	return key[prefixLen:]
}
