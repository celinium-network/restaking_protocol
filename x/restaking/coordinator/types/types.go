package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

type (
	Int                        = sdkmath.Int
	Dec                        = sdk.Dec
	TendermintLightClientState = ibctm.ClientState
)
