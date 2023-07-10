package types_test

import (
	"fmt"
	"testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
)

var PKs = simtestutil.CreateTestPubKeys(5)

func TestParseValidatorOperatorKey(t *testing.T) {
	simAccounts := simtestutil.CreateIncrementalAccounts(1)
	operator := simAccounts[0]

	valAddr := sdk.ValAddress(PKs[0].Address().Bytes())
	operatorAddr := operator.String()

	key := types.OperatorAddressKey(operator, valAddr)

	addrFromParse := types.ParseValidatorOperatorKey(key)

	if string(addrFromParse) != operatorAddr {
		panic(fmt.Sprintf("%s not equal %s", addrFromParse, operatorAddr))
	}
}
