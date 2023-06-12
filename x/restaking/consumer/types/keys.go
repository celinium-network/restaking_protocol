package types

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
	ValidatorSetUpdateIDKey = iota

	ValidatorSetChangeSet

	CoordinatorChannelID

	OperatorAddressPrefix
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

func OperatorAddressKey(validatorPK []byte, operatorAddress string) []byte {
	return append([]byte{OperatorAddressPrefix}, append([]byte(operatorAddress), validatorPK...)...)
}
