package types

const (
	// ModuleName is the name of the restaking coordinator module
	ModuleName = "restakingCoordinator"

	// StoreKey is the string store representation
	StoreKey = ModuleName

	// RouterKey is the msg router key for the restaking coordinator module
	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

const (
	ConsumerAdditionProposalPrefix byte = iota

	ConsumerClientIDPrefix

	ConsumerValidatorSetPrefix

	PortByteKey
)

func ConsumerAdditionProposalKey(chainID string) []byte {
	return append([]byte{ConsumerAdditionProposalPrefix}, []byte(chainID)...)
}

func ConsumerClientIDKey(chainID string) []byte {
	return append([]byte{ConsumerClientIDPrefix}, []byte(chainID)...)
}

func ConsumerValidatorSetKey(chainID string) []byte {
	return append([]byte{ConsumerValidatorSetPrefix}, []byte(chainID)...)
}

func PortKey() []byte {
	return []byte{PortByteKey}
}
