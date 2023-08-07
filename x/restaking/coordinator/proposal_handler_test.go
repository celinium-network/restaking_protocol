package coordinator_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/client/cli"
)

func TestAddConsumerProposal(t *testing.T) {
	proposal := cli.ConsumerAdditionProposalJSON{
		Title:                 "add restaking consumer",
		Description:           "description",
		ChainID:               "consumer",
		UnbondingPeriod:       stakingtypes.DefaultUnbondingTime,
		TimeoutPeriod:         time.Minute * 10,
		TransferTimeoutPeriod: time.Minute * 10,
		RestakingTokens:       []string{"CNTR"},
		RewardTokens:          []string{"NCT"},
		TransferChannelID:     "channel-0",
		Deposit:               "10000000CNTR",
	}

	jsonStr, _ := json.Marshal(proposal)

	os.WriteFile("./proposal.json", jsonStr, 0644)
}
