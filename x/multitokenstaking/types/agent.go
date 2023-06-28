package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ma MTStakingAgent) CalculateSharesFromTokenAmount(tokenAmt math.Int) math.Int {
	if ma.StakedAmount.IsZero() {
		return tokenAmt
	}

	return sdk.NewDecFromInt(tokenAmt).QuoInt(ma.StakedAmount).MulInt(ma.Shares).TruncateInt()
}

func (ma MTStakingAgent) CalculateCoins(shareAmt math.Int) math.Int {
	if ma.StakedAmount.IsZero() {
		return math.ZeroInt()
	}

	return sdk.NewDecFromInt(shareAmt).QuoInt(ma.Shares).MulInt(ma.StakedAmount).TruncateInt()
}
