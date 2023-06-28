package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
		mtskingtypes.SharesPrefix,
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
	return [][]byte{
		mtskingtypes.GetMTStakingAgentIDKey("arg1", "arg2"),
		mtskingtypes.GetMTStakingAgentKey("arg1"),
		mtskingtypes.GetMTStakingSharesKey("arg1", "arg2"),
		mtskingtypes.GetMTStakingUnbondingKey("arg1", "arg2"),
		mtskingtypes.GetMTStakingUnbondingDelegationTimeKey(time.Now()),
		mtskingtypes.GetMTStakingMintedKey(),
		mtskingtypes.GetMTTokenMultiplierKey("arg1"),
	}
}
