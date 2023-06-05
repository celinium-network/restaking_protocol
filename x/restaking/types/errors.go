package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrDuplicateConsumerChain    = errorsmod.Register(ModuleName, 1, "consumer chain already exists")
	ErrInvalidChannelFlow        = errorsmod.Register(ModuleName, 2, "invalid message sent to channel end")
	ErrInvalidVersion            = errorsmod.Register(ModuleName, 3, "invalid restaking version")
	ErrUnauthorizedConsumerChain = errorsmod.Register(ModuleName, 4, "consumer chain is not authorized")
)
