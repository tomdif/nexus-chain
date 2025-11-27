package cli

import (
	"encoding/hex"
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
		CmdCancelJob(),
		CmdSubmitPublicJob(),
	)

	return cmd
}

func CmdPostJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-job [problem-hash] [threshold] [reward-amount]",
		Short: "Post a new paid optimization job",
		Long: `Post a new paid optimization job to the network.

The job will be queued and activated based on priority fee.
A 2% fee is burned from the reward amount.

Example:
  nexusd tx mining post-job \
    0000000000000000000000000000000000000000000000000000000000000001 \
    -1000 \
    1000000 \
    --priority-fee 100000 \
    --duration 86400 \
    --from mykey`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			problemHash := args[0]

			threshold, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			rewardAmt, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			priorityFeeAmt, err := cmd.Flags().GetInt64("priority-fee")
			if err != nil {
				return err
			}

			duration, err := cmd.Flags().GetInt64("duration")
			if err != nil {
				return err
			}

			quantumSafe, err := cmd.Flags().GetBool("quantum-safe")
			if err != nil {
				return err
			}

			msg := &types.MsgPostJob{
				Customer:    clientCtx.GetFromAddress().String(),
				ProblemType: "ising",
				ProblemHash: problemHash,
				Threshold:   threshold,
				Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", rewardAmt)),
				PriorityFee: sdk.NewCoins(sdk.NewInt64Coin("unexus", priorityFeeAmt)),
				Duration:    duration,
				QuantumSafe: quantumSafe,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Int64("priority-fee", 0, "Priority fee in unexus (higher = faster activation)")
	cmd.Flags().Int64("duration", 86400, "Job duration in seconds (default: 24 hours)")
	cmd.Flags().Bool("quantum-safe", false, "Require quantum-safe STARK proofs")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdSubmitProof() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proof [job-id] [solution-hash] [energy] [proof-hex]",
		Short: "Submit a zero-knowledge proof for a job",
		Long: `Submit a zero-knowledge proof demonstrating a solution to an optimization job.

The proof will be verified by the network's Nova verification service.
Shares are awarded based on energy improvement over previous best solution.

Example:
  nexusd tx mining submit-proof \
    paid_12345_abcd1234 \
    0000000000000000000000000000000000000000000000000000000000000002 \
    -1500 \
    deadbeef01020304 \
    --proof-type nova \
    --from mykey`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jobId := args[0]
			solutionHash := args[1]

			energy, err := cmd.Flags().GetInt64("energy")
			if err != nil {
				return err
			}

			proofHex := args[2]
			proofBytes, err := hex.DecodeString(proofHex)
			if err != nil {
				return err
			}

			proofType, err := cmd.Flags().GetString("proof-type")
			if err != nil {
				return err
			}

			msg := &types.MsgSubmitProof{
				Miner:        clientCtx.GetFromAddress().String(),
				JobId:        jobId,
				SolutionHash: solutionHash,
				Energy:       energy,
				Proof:        proofBytes,
				ProofType:    proofType,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("proof-type", "nova", "Proof type: nova or stark")
	cmd.Flags().Int64("energy", 0, "Energy value of the solution (can be negative)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdClaimRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards [job-id]",
		Short: "Claim mining rewards for a completed job",
		Long: `Claim your earned mining rewards for a completed job.

Rewards include:
- 80% of your proportional share of the customer's payment
- 80% of your proportional share of emission rewards (time-based)

The remaining 20% goes to validators.

Example:
  nexusd tx mining claim-rewards paid_12345_abcd1234 --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jobId := args[0]

			msg := &types.MsgClaimRewards{
				Claimer: clientCtx.GetFromAddress().String(),
				JobId:   jobId,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdCancelJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-job [job-id]",
		Short: "Cancel a queued job and receive refund",
		Long: `Cancel a job that has not yet received any mining work.

You will receive a refund of the net reward amount (after the 2% fee was burned).
The priority fee is NOT refunded as it was already burned.

Can only cancel if:
- You are the job customer
- No shares have been earned yet (TotalShares = 0)

Example:
  nexusd tx mining cancel-job paid_12345_abcd1234 --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			jobId := args[0]

			msg := &types.MsgCancelJob{
				Customer: clientCtx.GetFromAddress().String(),
				JobId:    jobId,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSubmitPublicJob() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-public-job [title] [category] [problem-hash] [threshold] [ipfs-cid]",
		Short: "Submit a free public benefit job",
		Long: `Submit a free optimization job for public benefit (research, science, etc).

Requirements:
- Minimum stake: 10,000 NEX
- Posting fee: 100 NEX (burned)
- Problem must meet minimum difficulty

The job will be queued and randomly selected when no paid jobs are active.
Miners earn emission rewards only (no customer payment).

Valid categories:
- protein_folding
- molecular_docking
- materials_science
- logistics
- scheduling
- graph_optimization
- machine_learning
- cryptography
- physics_simulation
- other

Example:
  nexusd tx mining submit-public-job \
    "Protein Folding - COVID-19 Spike" \
    protein_folding \
    0000000000000000000000000000000000000000000000000000000000000001 \
    -5000 \
    QmXyz123... \
    --description "Folding simulation for COVID-19 research" \
    --from mykey`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			title := args[0]
			category := args[1]
			problemHash := args[2]

			threshold, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			ipfsCid := args[4]

			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return err
			}

			msg := &types.MsgSubmitPublicJob{
				Submitter:   clientCtx.GetFromAddress().String(),
				Title:       title,
				Category:    category,
				ProblemHash: problemHash,
				Threshold:   threshold,
				IpfsCid:     ipfsCid,
				Description: description,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("description", "", "Optional detailed description of the problem")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
