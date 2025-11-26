package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"nexus/x/mining/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Mining transaction commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdPostJob(),
		CmdSubmitProof(),
		CmdClaimRewards(),
	)

	return cmd
}

func CmdPostJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-job [problem-hash] [threshold] [reward-amount] [duration-blocks]",
		Short: "Post a new optimization job",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			threshold, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			rewardAmt, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			duration, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgPostJob{
				Customer:    clientCtx.GetFromAddress().String(),
				ProblemHash: args[0],
				Threshold:   threshold,
				Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", rewardAmt)),
				Duration:    duration,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSubmitProof() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proof [job-id] [solution-hash] [energy] [proof-hex]",
		Short: "Submit a proof for a job",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			energy, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgSubmitProof{
				Miner:        clientCtx.GetFromAddress().String(),
				JobId:        args[0],
				SolutionHash: args[1],
				Energy:       energy,
				Proof:        []byte(args[3]),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdClaimRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards [job-id]",
		Short: "Claim rewards for a completed job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgClaimRewards{
				Claimer: clientCtx.GetFromAddress().String(),
				JobId:   args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
