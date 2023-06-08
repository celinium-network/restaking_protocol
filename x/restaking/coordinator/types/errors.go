package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrAdditionalProposalNotFound = errorsmod.Register(ModuleName, 1, "add consumer proposal not found")
	ErrMismatchParams             = errorsmod.Register(ModuleName, 2, "parameters of msg are't match")
	ErrUnknownConsumer            = errorsmod.Register(ModuleName, 3, "")
	ErrNotExistedValidator        = errorsmod.Register(ModuleName, 4, "")
	ErrUnsupportedRestakingToken  = errorsmod.Register(ModuleName, 5, "The consumer do't support the restaking token")
	ErrUnknownOperator            = errorsmod.Register(ModuleName, 6, "")
)
