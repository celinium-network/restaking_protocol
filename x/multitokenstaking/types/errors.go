package types

import (
	sdkioerrors "cosmossdk.io/errors"
)

var (
	ErrForbidStakingDenom    = sdkioerrors.Register(ModuleName, 1, "The denom is forbidden in multistaking module")
	ErrNotExistedValidator   = sdkioerrors.Register(ModuleName, 2, "The validator is not exist")
	ErrInsufficientShares    = sdkioerrors.Register(ModuleName, 3, "The shares is insufficient")
	ErrNotExistedAgent       = sdkioerrors.Register(ModuleName, 4, "The validator has't multistaking agent")
	ErrNoUnbondingDelegation = sdkioerrors.Register(ModuleName, 5, "The unbonding delegation is not existed")
	ErrNoShares              = sdkioerrors.Register(ModuleName, 6, "The user has't shares in this agent")
)
