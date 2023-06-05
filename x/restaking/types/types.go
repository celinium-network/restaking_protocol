package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
)

type (
	TendermintABCIValidatorUpdate = abci.ValidatorUpdate
	ValidatorSet                  = tmtypes.ValidatorSet
)
