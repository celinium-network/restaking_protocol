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
	ErrInsufficientDelegation     = errorsmod.Register(ModuleName, 7, "")
	ErrIBCCallbackNotExisted      = errorsmod.Register(ModuleName, 8, "")
	ErrUnknownIBCCallbackType     = errorsmod.Register(ModuleName, 9, "")
	ErrIBCCallback                = errorsmod.Register(ModuleName, 10, "Some error was generated in process ibc callback handler")
	ErrMismatchStatus             = errorsmod.Register(ModuleName, 11, "")
	ErrNoUnbondingDelegation      = errorsmod.Register(ModuleName, 12, "")
)
