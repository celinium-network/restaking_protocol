package types_test

import (
	"bytes"
	"fmt"
	"testing"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/consumer/types"
)

var PKs = simtestutil.CreateTestPubKeys(5)

func TestParseValidatorOperatorKey(t *testing.T) {
	simAccounts := simtestutil.CreateIncrementalAccounts(1)
	operatorAccAddr := simAccounts[0]

	valAddr := sdk.ValAddress(PKs[0].Address().Bytes())
	key := types.OperatorAddressKey(operatorAccAddr, valAddr)
	addrFromParse := types.ParseValidatorOperatorKey(key)
	if !bytes.Equal(addrFromParse, operatorAccAddr) {
		panic(fmt.Sprintf("%s not equal %s", addrFromParse, operatorAccAddr))
	}
}
