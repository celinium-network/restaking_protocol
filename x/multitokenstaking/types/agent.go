package types

import (
	"cosmossdk.io/math"
)

func (ma MTStakingAgent) CalculateShareFromCoin(tokenAmt math.Int) math.Int {
	if ma.StakedAmount.IsZero() {
		return tokenAmt
	}

	return tokenAmt.Mul(ma.Shares).Quo(ma.StakedAmount)
}

func (ma MTStakingAgent) CalculateCoinFromShare(shareAmt math.Int) math.Int {
	if ma.StakedAmount.IsZero() {
		return math.ZeroInt()
	}

	return shareAmt.Mul(ma.StakedAmount).Quo(ma.Shares)
}
