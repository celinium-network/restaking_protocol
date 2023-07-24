package types

import "cosmossdk.io/math"

func (o Operator) TokensFromShares(shares math.Int) math.Int {
	return o.RestakedAmount.Mul(shares).Quo(o.Shares)
}
