package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"nexus/x/mining/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Mining query commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryJob(),
		CmdQueryShares(),
	)

	return cmd
}

func CmdQueryJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job [job-id]",
		Short: "Query a job by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			cmd.Printf("Querying job: %s\n", args[0])
			cmd.Printf("Client context chain-id: %s\n", clientCtx.ChainID)
			
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryShares() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shares [job-id] [miner-address]",
		Short: "Query shares for a miner on a job",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			cmd.Printf("Querying shares for job %s, miner %s\n", args[0], args[1])
			cmd.Printf("Client context chain-id: %s\n", clientCtx.ChainID)
			
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
