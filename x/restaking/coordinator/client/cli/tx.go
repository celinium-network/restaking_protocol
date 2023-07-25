package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func NewTxCommand() *cobra.Command {
	stakingTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Restaking Coordinator transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingTxCmd.AddCommand(
		NewRegisterOperatorCmd(),
		NewDelegateCmd(),
		NewUndelegateCmd(),
		NewWithdrawRewardCmd(),
	)

	return stakingTxCmd
}

func NewRegisterOperatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: `register-operator [consumer-chain-ids] [consumer-validator-addresses] 
[restaking-denom]`,
		Short: `register a new operator for restaking. eg register-operator "chain0,chain1" "chain0x123,chain10x234" 
token`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := clientCtx.GetFromAddress().String()
			chainIDs := strings.Split(args[0], ",")
			addressed := strings.Split(args[1], ",")

			msg := types.MsgRegisterOperatorRequest{
				ConsumerChainIDs:           chainIDs,
				ConsumerValidatorAddresses: addressed,
				RestakingDenom:             args[2],
				Sender:                     sender,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	return cmd
}

func NewDelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `delegate [delegator-address] [operator-address] [amount]`,
		Short: `delegate to the operator. eg: delegate [celi0x123] [celi0x23] 100000`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("can't new int from %s", args[2])
			}

			msg := types.MsgDelegateRequest{
				DelegatorAddress: args[0],
				OperatorAddress:  args[1],
				Amount:           amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	return cmd
}

func NewUndelegateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `undelegate [delegator-address] [operator-address] [amount]`,
		Short: `undelegate from the operator. eg: delegate [celi0x123] [celi0x23] 100000`,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("can't new int from %s", args[2])
			}

			msg := types.MsgUndelegateRequest{
				DelegatorAddress: args[0],
				OperatorAddress:  args[1],
				Amount:           amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	return cmd
}

func NewWithdrawRewardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `withdraw-reward [delegator-address] [operator-address]`,
		Short: `withdraw reward from the operator. eg: withdraw-reward [celi0x123] [celi0x23]`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgWithdrawRewardRequest{
				DelegatorAddress: args[0],
				OperatorAddress:  args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	return cmd
}
