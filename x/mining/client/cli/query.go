package cli

import (
	"encoding/json"
	"fmt"

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
		CmdQueryParams(),
		CmdQueryActiveJob(),
		CmdQueryEmissionInfo(),
		CmdQueryJobs(),
	)

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-params",
		Short: "Get module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Query the store directly
			res, _, err := clientCtx.QueryStore([]byte("params"), types.StoreKey)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("params not found")
			}

			var params types.Params
			if err := clientCtx.Codec.Unmarshal(res, &params); err != nil {
				return err
			}

			// Print as JSON
			out, _ := json.MarshalIndent(params, "", "  ")
			fmt.Println(string(out))
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryActiveJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-active-job",
		Short: "Get the currently active background job",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Query current job ID
			res, _, err := clientCtx.QueryStore([]byte("current_job_id"), types.StoreKey)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				fmt.Println(`{"job": null, "message": "No active job"}`)
				return nil
			}

			jobID := string(res)

			// Query the job
			// Query the job - use proper key prefix (0x01)
			jobKey := append([]byte{0x01}, []byte(jobID)...)
			jobRes, _, err := clientCtx.QueryStore(jobKey, types.StoreKey)
			if err != nil {
				return err
			}

			if len(jobRes) == 0 {
				fmt.Printf(`{"job_id": "%s", "message": "Job not found"}`, jobID)
				return nil
			}

			var job types.Job
			if err := clientCtx.Codec.Unmarshal(jobRes, &job); err != nil {
				return err
			}

			out, _ := json.MarshalIndent(job, "", "  ")
			fmt.Println(string(out))
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryEmissionInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-emission-info",
		Short: "Show current emission epoch and rate",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Query genesis minute
			res, _, err := clientCtx.QueryStore([]byte("genesis_minute"), types.StoreKey)
			if err != nil {
				return err
			}

			info := map[string]interface{}{
				"genesis_minute_stored": len(res) > 0,
				"raw_bytes":             len(res),
			}

			// Query emission escrow
			escrowRes, _, _ := clientCtx.QueryStore([]byte("emission_escrow"), types.StoreKey)
			info["escrow_stored"] = len(escrowRes) > 0

			out, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(out))
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryJobs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-jobs",
		Short: "List all jobs (note: iterating store not supported via CLI)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(`{"message": "Job listing requires gRPC/API server. Check individual jobs with get-active-job."}`)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
