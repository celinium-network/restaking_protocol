package types_test

import (
	"testing"

	"cosmossdk.io/math"

	mtskingtypes "github.com/celinium-network/restaking_protocol/x/multitokenstaking/types"
	"github.com/stretchr/testify/require"
)

func TestAgentCalculateShareFromCoin(t *testing.T) {
	tests := []struct {
		describe             string
		agentStakedAmount    string
		agentShares          string
		inputCoinAmount      string
		expectedOutputShares string
	}{
		{
			"basic",
			"1000",
			"1000",
			"100",
			"100",
		},
		{
			"overflow test",
			"340282366920938463463374607431768211455", // u128 max
			"340282366920938463463374607431768211455",
			"1000000000000",
			"1000000000000",
		},
		{
			"shares greater then staked amount",
			"11000000000000000000000",
			"10000000000000000000000",
			"10000000000000000000",
			"9090909090909090909",
		},
		{
			"staked amount greater then shares",
			"10000000000000000000000",
			"11000000000000000000000",
			"900000000000000000000",
			"990000000000000000000", // 11000000000000000000000 * 900000000000000000000/ 10000000000000000000000
		},
	}

	for _, test := range tests {
		stakedAmount, ok := math.NewIntFromString(test.agentStakedAmount)
		require.True(t, ok)
		sharesAMount, ok := math.NewIntFromString(test.agentShares)
		require.True(t, ok)
		inputCoinAmount, ok := math.NewIntFromString(test.inputCoinAmount)
		require.True(t, ok)
		expectedOutputAmount, ok := math.NewIntFromString(test.expectedOutputShares)
		require.True(t, ok)

		agent := mtskingtypes.MTStakingAgent{
			StakedAmount: stakedAmount,
			Shares:       sharesAMount,
		}

		res := agent.CalculateShareFromCoin(inputCoinAmount)
		require.True(t, res.Equal(expectedOutputAmount), test.describe)
	}
}

func TestAgentCalculateCoinFromShare(t *testing.T) {
	tests := []struct {
		describe                 string
		agentStakedAmount        string
		agentShares              string
		inputSharesAmount        string
		expectedOutputCoinAmount string
	}{
		{
			"basic",
			"1000",
			"1000",
			"100",
			"100",
		},
		{
			"overflow test",
			"340282366920938463463374607431768211455", // u128 max
			"340282366920938463463374607431768211455",
			"1000000000000",
			"1000000000000",
		},
		{
			"shares greater then staked amount",
			"11000000000000000000000",
			"10000000000000000000000",
			"10000000000000000000",
			"11000000000000000000", // 11000000000000000000000 * 10000000000000000000/ 10000000000000000000000
		},
		{
			"staked amount greater then shares",
			"10000000000000000000000",
			"11000000000000000000000",
			"900000000000000000000",
			"818181818181818181818", // 10000000000000000000000 * 900000000000000000000/ 11000000000000000000000
		},
	}

	for _, test := range tests {
		stakedAmount, ok := math.NewIntFromString(test.agentStakedAmount)
		require.True(t, ok)
		sharesAMount, ok := math.NewIntFromString(test.agentShares)
		require.True(t, ok)
		inputSharesAmount, ok := math.NewIntFromString(test.inputSharesAmount)
		require.True(t, ok)
		expectedOutputAmount, ok := math.NewIntFromString(test.expectedOutputCoinAmount)
		require.True(t, ok)

		agent := mtskingtypes.MTStakingAgent{
			StakedAmount: stakedAmount,
			Shares:       sharesAMount,
		}

		res := agent.CalculateCoinFromShare(inputSharesAmount)
		require.True(t, res.Equal(expectedOutputAmount), test.describe)
	}
}
