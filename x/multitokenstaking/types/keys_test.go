package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mtskingtypes "github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
)

func TestNoDuplicates(t *testing.T) {
	prefixes := getAllKeyPrefixes()
	seen := [][]byte{}

	for _, prefix := range prefixes {
		require.NotContains(t, seen, prefix, "Duplicate key prefix: %v", prefix)
		seen = append(seen, prefix)
	}
}

func getAllKeyPrefixes() [][]byte {
	return [][]byte{
		mtskingtypes.DenomWhiteListKey,
		mtskingtypes.AgentPrefix,
		mtskingtypes.UnbondingPrefix,
		mtskingtypes.UnbondingQueueKey,
		mtskingtypes.DelegationPrefix,
		mtskingtypes.MintedKey,
		mtskingtypes.TokenMultiplierPrefix,
	}
}

func TestNoPrefixOverlap(t *testing.T) {
	keys := getAllFullyDefinedKeys()
	seenPrefixes := []byte{}
	for _, key := range keys {
		require.NotContains(t, seenPrefixes, key[0], "Duplicate key prefix: %v", key[0])
		seenPrefixes = append(seenPrefixes, key[0])
	}
}

func getAllFullyDefinedKeys() [][]byte {
	pks := simtestutil.CreateTestPubKeys(5)
	accounts := simtestutil.CreateIncrementalAccounts(2)

	valAddr := sdk.ValAddress(pks[0].Address())

	return [][]byte{
		mtskingtypes.GetMTStakingAgentAddressKey("arg1", valAddr),
		mtskingtypes.GetMTStakingAgentKey(accounts[0]),
		mtskingtypes.GetMTStakingDelegationKey(accounts[0], accounts[1]),
		mtskingtypes.GetMTStakingUnbondingKey(accounts[0], accounts[1]),
		mtskingtypes.GetMTStakingUnbondingDelegationTimeKey(time.Now()),
		mtskingtypes.GetMTStakingMintedKey(),
		mtskingtypes.GetMTTokenMultiplierKey("arg1"),
	}
}
