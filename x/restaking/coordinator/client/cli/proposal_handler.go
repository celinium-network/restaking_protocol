package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

var ConsumerAdditionProposalHandler = govclient.NewProposalHandler(SubmitConsumerAdditionProposalTxCmd)

func SubmitConsumerAdditionProposalTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `consumer-addition [proposal_file]`,
		Short: `Submit a consumer addition proposal`,
		Long: `
Submit a consumer addition proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.
Example:
$ <appdx> gov tx submit-proposal consumer-addition <path/to/proposal.json> ...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := ParseConsumerAdditionProposalJSON(args[0])
			if err != nil {
				return err
			}

			govContext := types.ConsumerAdditionProposal{
				Title:                 proposal.Title,
				Description:           proposal.Description,
				ChainId:               proposal.ChainID,
				UnbondingPeriod:       proposal.UnbondingPeriod,
				TimeoutPeriod:         proposal.TimeoutPeriod,
				TransferTimeoutPeriod: proposal.TransferTimeoutPeriod,
				RestakingTokens:       proposal.RestakingTokens,
				RewardTokens:          proposal.RewardTokens,
				TransferChannelId:     proposal.TransferChannelID,
			}

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(&govContext, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	return cmd
}

type ConsumerAdditionProposalJSON struct {
	Title                 string        `json:"title"`
	Description           string        `json:"description"`
	ChainID               string        `json:"chain_id"`
	UnbondingPeriod       time.Duration `json:"unbonding_period"`
	TimeoutPeriod         time.Duration `json:"timeout_period"`
	TransferTimeoutPeriod time.Duration `json:"transfer_timeout_period"`
	RestakingTokens       []string      `json:"restaking_tokens"`
	RewardTokens          []string      `json:"reward_tokens"`
	TransferChannelID     string        `json:"transfer_channel_id"`

	Deposit string `json:"deposit"`
}

func ParseConsumerAdditionProposalJSON(proposalFile string) (ConsumerAdditionProposalJSON, error) {
	proposal := ConsumerAdditionProposalJSON{}

	contents, err := os.ReadFile(filepath.Clean(proposalFile))
	if err != nil {
		return proposal, err
	}

	if err := json.Unmarshal(contents, &proposal); err != nil {
		return proposal, err
	}

	return proposal, nil
}
