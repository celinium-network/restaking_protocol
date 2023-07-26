package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/celinium-network/restaking_protocol/x/restaking/coordinator/types"
)

func GetQueryCmd() *cobra.Command {
	restakingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the restaking module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	restakingQueryCmd.AddCommand(
		GetCmdQueryOperators(),
		GetCmdQueryOperator(),
		GetCmdQueryOperatorDelegations(),
		GetCmdQueryDelegation(),
		GetCmdQueryDelegatorDelegations(),
		GetCmdQueryDelegatorOperators(),
		GetCmdQueryDelegatorUnbondingDelegations(),
		GetCmdQueryOperatorUnbondingDelegations(),
		GetCmdQueryUnbondingDelegation(),
	)

	return restakingQueryCmd
}

func GetCmdQueryOperators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `operators`,
		Short: `query all operators`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.Operators(cmd.Context(), &types.QueryOperatorsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	return cmd
}

func GetCmdQueryOperator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `operator [operator-address]`,
		Short: `query operator by address`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			result, err := queryClient.Operator(cmd.Context(), &types.QueryOperatorRequest{
				OperatorAddress: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&result.Operator)
		},
	}

	return cmd
}

func GetCmdQueryOperatorDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `operator-delegations [operator-address]`,
		Short: `query delegation of operator`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			result, err := queryClient.OperatorDelegations(cmd.Context(), &types.QueryOperatorDelegationsRequest{
				OperatorAddress: args[0],
				Pagination:      pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	return cmd
}

func GetCmdQueryDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `delegation [delegator-address] [operator-address]`,
		Short: `query delegation of delegator operator pair`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			result, err := queryClient.Delegation(cmd.Context(), &types.QueryDelegationRequest{
				OperatorAddress:  args[0],
				DelegatorAddress: args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	return cmd
}

func GetCmdQueryDelegatorDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `delegation [delegator-address]`,
		Short: `query all delegations of the delegator`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.DelegatorDelegations(cmd.Context(), &types.QueryDelegatorDelegationsRequest{
				DelegatorAddr: args[0],
				Pagination:    pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	return cmd
}

func GetCmdQueryDelegatorOperators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `delegator-operators [delegator-address]`,
		Short: `query all operators which the delegator has delegated`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.DelegatorOperators(cmd.Context(), &types.QueryDelegatorOperatorsRequest{
				DelegatorAddr: args[0],
				Pagination:    pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	return cmd
}

func GetCmdQueryDelegatorUnbondingDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `delegator-unbonding [delegator-address]`,
		Short: `query all unbonding delegations of the delegator`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.DelegatorUnbondingDelegations(cmd.Context(), &types.QueryDelegatorUnbondingDelegationsRequest{
				DelegatorAddr: args[0],
				Pagination:    pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	return cmd
}

func GetCmdQueryOperatorUnbondingDelegations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `operator-unbonding [operator-address]`,
		Short: `query all unbonding delegations of the operator`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.OperatorUnbondingDelegations(cmd.Context(), &types.QueryOperatorUnbondingDelegationsRequest{
				OperatorAddress: args[0],
				Pagination:      pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	return cmd
}

func GetCmdQueryUnbondingDelegation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `unbonding [delegator-address] [operator-address]`,
		Short: `query all unbonding delegations of the delegator operator pair`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			result, err := queryClient.UnbondingDelegation(cmd.Context(), &types.QueryUnbondingDelegationRequest{
				DelegatorAddress: args[0],
				OperatorAddress:  args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	return cmd
}
