package types

import (
	errorsmod "cosmossdk.io/errors"
)

var ErrRestakingChannelNotFound = errorsmod.Register(ModuleName, 1, "can't found restaking protocol ibc channel")
