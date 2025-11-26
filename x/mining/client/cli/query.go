package cli

import (
	"context"
	"fmt"
	"strconv"

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
		CmdQueryJobs(),
		CmdQueryMiner(),
		CmdQueryParams(),
		CmdQueryActiveJob(),
		CmdQueryQueueStatus(),
		CmdQueryEmissionInfo(),
	)

	return cmd
}

func CmdQueryJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-job [job-id]",
		Short: "Get details of a specific job",
		Long: `Query detailed information about a specific job by its ID.

Example:
  nexusd query mining get-job paid_12345_abcd1234`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryJobRequest{
				JobId: args[0],
			}

			res, err := queryClient.Job(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryJobs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-jobs",
		Short: "List all jobs with optional status filter",
		Long: `Query a list of all jobs in the system.

Optional flags:
  --status: Filter by job status (0=pending, 1=active, 2=completed, 3=expired, 4=cancelled, 5=queued)

Example:
  nexusd query mining list-jobs
  nexusd query mining list-jobs --status 1`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			status, err := cmd.Flags().GetUint32("status")
			if err != nil {
				return err
			}

			req := &types.QueryJobsRequest{
				Status: status,
			}

			res, err := queryClient.Jobs(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Uint32("status", 999, "Filter by job status (0-5, 999=all)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryMiner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-miner [address]",
		Short: "Get miner statistics and pending rewards",
		Long: `Query detailed statistics for a miner including:
- Total shares earned across all jobs
- Pending rewards (claimable)
- Jobs participated in
- Total rewards claimed

Example:
  nexusd query mining get-miner nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryMinerStatsRequest{
				MinerAddress: args[0],
			}

			res, err := queryClient.MinerStats(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-params",
		Short: "Get module parameters",
		Long: `Query the current mining module parameters including:
- Miner share percentage (default: 80%)
- Validator share percentage (default: 20%)
- Checkpoint interval (blocks)
- Minimum proof period
- Job fee burn percentage
- Transaction fee burn percentage
- Background emission rate
- Minimum job reward
- Maximum job duration

Example:
  nexusd query mining get-params`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryParamsRequest{}

			res, err := queryClient.Params(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryActiveJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-active-job",
		Short: "Get the currently active background job",
		Long: `Query the currently active background job that miners can work on.

This shows:
- Job ID
- Problem type and size
- Threshold requirement
- Current best energy
- Total shares earned
- Time remaining

Example:
  nexusd query mining get-active-job`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryActiveJobRequest{}

			res, err := queryClient.ActiveJob(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryQueueStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-queue-status",
		Short: "Show paid and public job queue status",
		Long: `Query the current status of job queues:
- Paid job queue length (priority fee sorted)
- Public job queue length (random selection)
- Next jobs to be activated
- Queue positions

Example:
  nexusd query mining get-queue-status`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryQueueStatusRequest{}

			res, err := queryClient.QueueStatus(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryEmissionInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-emission-info",
		Short: "Show current emission rate and epoch information",
		Long: `Query the current emission schedule information:
- Current epoch (1-8+)
- Emission rate per minute
- Total emission escrow
- Genesis time
- Time until next epoch

The emission schedule follows a halving pattern:
- Epoch 1 (Years 0-2): 100% = 35,950 NEX/min
- Epoch 2 (Years 2-4): 50%
- Epoch 3 (Years 4-6): 25%
- Epoch 4 (Years 6-8): 12.5%
- Epoch 5 (Years 8-10): 6.25%
- Epoch 6 (Years 10-12): 3.125%
- Epoch 7 (Years 12-14): 1.5625%
- Epoch 8+ (Years 14+): 1.5% perpetual

Example:
  nexusd query mining get-emission-info`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryEmissionInfoRequest{}

			res, err := queryClient.EmissionInfo(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryShares() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shares [job-id] [miner-address]",
		Short: "Query shares for a miner on a specific job",
		Long: `Query the number of shares a miner has earned on a specific job.

Example:
  nexusd query mining shares paid_12345_abcd1234 nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryMinerSharesRequest{
				JobId: args[0],
				Miner: args[1],
			}

			res, err := queryClient.MinerShares(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryCheckpoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint [id]",
		Short: "Query a specific checkpoint by ID",
		Long: `Query detailed information about a checkpoint.

Example:
  nexusd query mining checkpoint 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid checkpoint ID: %w", err)
			}

			req := &types.QueryCheckpointRequest{
				Id: id,
			}

			res, err := queryClient.Checkpoint(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryLatestCheckpoint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-checkpoint",
		Short: "Query the latest checkpoint",
		Long: `Query the most recent checkpoint created.

Example:
  nexusd query mining latest-checkpoint`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryLatestCheckpointRequest{}

			res, err := queryClient.LatestCheckpoint(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
