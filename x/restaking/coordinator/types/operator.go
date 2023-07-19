package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (o Operator) TokensFromShares(shares sdk.Int) sdk.Int {
	return o.RestakedAmount.Mul(shares).Quo(o.Shares)
}
