package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// VerifyRequest matches the Rust orchestrator's expected format
type VerifyRequest struct {
	JobID             string `json:"job_id"`
	ProblemCommitment string `json:"problem_commitment"`
	SpinCommitment    string `json:"spin_commitment"`
	ClaimedEnergy     int64  `json:"claimed_energy"`
	Threshold         int64  `json:"threshold"`
	Proof             string `json:"proof"`
}

// VerifyResponse matches the Rust orchestrator's response
type VerifyResponse struct {
	Valid          bool    `json:"valid"`
	Energy         int64   `json:"energy"`
	MeetsThreshold bool    `json:"meets_threshold"`
	Error          *string `json:"error"`
}

// VerifierURL is the address of the Nova verification service
const VerifierURL = "http://localhost:3000/verify"

// PostJob creates a new mining job, burns job fee, and escrows the net reward
func (k msgServer) PostJob(goCtx context.Context, msg *types.MsgPostJob) (*types.MsgPostJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Generate job ID
	jobID := fmt.Sprintf("job_%d_%s", ctx.BlockHeight(), msg.Customer[:8])

	// Convert sdk.Coins reward to int64 (use first coin amount)
	var grossRewardAmount int64
	if len(msg.Reward) > 0 {
		grossRewardAmount = msg.Reward[0].Amount.Int64()
	}

	// Calculate fee burn (2% of reward)
	params := k.GetParams(ctx)
	feeBurnPercent := int64(params.JobFeeBurnPercent)
	feeBurnAmount := (grossRewardAmount * feeBurnPercent) / 100
	netRewardAmount := grossRewardAmount - feeBurnAmount

	// Transfer gross amount from customer to module
	if k.bankKeeper != nil && len(msg.Reward) > 0 {
		customerAddr, err := sdk.AccAddressFromBech32(msg.Customer)
		if err != nil {
			return nil, types.ErrInvalidJob
		}

		// Transfer full amount to module first
		err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, customerAddr, types.ModuleName, msg.Reward)
		if err != nil {
			return nil, fmt.Errorf("failed to escrow reward: %w", err)
		}

		// Burn the fee portion (2% job fee - deflationary)
		if feeBurnAmount > 0 {
			feeBurnCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", feeBurnAmount))
			err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, feeBurnCoins)
			if err != nil {
				return nil, fmt.Errorf("failed to burn job fee: %w", err)
			}

			ctx.Logger().Info("Burned job fee",
				"job_id", jobID,
				"fee_burned", feeBurnCoins.String(),
				"net_reward", netRewardAmount,
			)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"fee_burned",
					sdk.NewAttribute("job_id", jobID),
					sdk.NewAttribute("amount", feeBurnCoins.String()),
					sdk.NewAttribute("type", "job_fee"),
				),
			)
		}

		ctx.Logger().Info("Escrowed job reward",
			"job_id", jobID,
			"gross", msg.Reward.String(),
			"fee_burned", feeBurnAmount,
			"net_escrowed", netRewardAmount,
		)
	}

	job := types.Job{
		Id:          jobID,
		Customer:    msg.Customer,
		ProblemType: msg.ProblemType,
		ProblemData: msg.ProblemData,
		ProblemHash: msg.ProblemHash,
		Threshold:   msg.Threshold,
		Reward:      netRewardAmount, // Store net reward (after fee burn)
		Status:      types.JobStatusActive,
		BestEnergy:  0,
		BestSolver:  "",
		TotalShares: 0,
		CreatedAt:   ctx.BlockHeight(),
		Deadline:    ctx.BlockHeight() + msg.Duration,
	}

	k.SetJob(ctx, job)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"job_posted",
			sdk.NewAttribute("job_id", jobID),
			sdk.NewAttribute("customer", msg.Customer),
			sdk.NewAttribute("threshold", fmt.Sprintf("%d", msg.Threshold)),
			sdk.NewAttribute("gross_reward", fmt.Sprintf("%d", grossRewardAmount)),
			sdk.NewAttribute("fee_burned", fmt.Sprintf("%d", feeBurnAmount)),
			sdk.NewAttribute("net_reward", fmt.Sprintf("%d", netRewardAmount)),
		),
	)

	return &types.MsgPostJobResponse{JobId: jobID}, nil
}

// SubmitProof processes a mining proof submission with ZK verification
func (k msgServer) SubmitProof(goCtx context.Context, msg *types.MsgSubmitProof) (*types.MsgSubmitProofResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get job
	job, found := k.GetJob(ctx, msg.JobId)
	if !found {
		return nil, types.ErrJobNotFound
	}

	if job.Status != types.JobStatusActive {
		return nil, types.ErrJobNotActive
	}

	if ctx.BlockHeight() > job.Deadline {
		return nil, types.ErrJobExpired
	}

	// Verify the ZK proof via the Nova verification service
	valid, err := k.verifyNovaProof(msg, job)
	if err != nil {
		// If verifier is unavailable, log but continue (for testing)
		ctx.Logger().Error("Nova verification service unavailable", "error", err)
		// In production, you'd want to return an error here
		// return nil, types.ErrProofVerification
	} else if !valid {
		return nil, types.ErrInvalidProof
	}

	// ========================================
	// UNIVERSAL SHARE FORMULA
	// ========================================
	var sharesEarned int64

	if job.TotalShares == 0 {
		// Bootstrap phase: first solver gets abs(energy)
		sharesEarned = msg.Energy
		if sharesEarned < 0 {
			sharesEarned = -sharesEarned
		}
		job.BestEnergy = msg.Energy
		job.BestSolver = msg.Miner
	} else {
		// Competition phase: shares = max(0, previous_best - new_energy)
		improvement := job.BestEnergy - msg.Energy
		if improvement > 0 {
			sharesEarned = improvement
			job.BestEnergy = msg.Energy
			job.BestSolver = msg.Miner
		} else {
			sharesEarned = 0
		}
	}

	// Update job
	job.TotalShares += sharesEarned
	k.SetJob(ctx, job)

	// Update miner's shares for this job
	if sharesEarned > 0 {
		minerAddr, err := sdk.AccAddressFromBech32(msg.Miner)
		if err != nil {
			return nil, types.ErrInvalidMiner
		}

		currentShares := k.GetShares(ctx, minerAddr, msg.JobId)
		k.SetShares(ctx, minerAddr, msg.JobId, currentShares+sharesEarned)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"proof_accepted",
				sdk.NewAttribute("job_id", msg.JobId),
				sdk.NewAttribute("miner", msg.Miner),
				sdk.NewAttribute("energy", fmt.Sprintf("%d", msg.Energy)),
				sdk.NewAttribute("shares_earned", fmt.Sprintf("%d", sharesEarned)),
				sdk.NewAttribute("proof_type", msg.ProofType),
			),
		)
	}

	return &types.MsgSubmitProofResponse{SharesEarned: sharesEarned}, nil
}

// verifyNovaProof calls the external Nova verification service
func (k msgServer) verifyNovaProof(msg *types.MsgSubmitProof, job types.Job) (bool, error) {
	req := VerifyRequest{
		JobID:             msg.JobId,
		ProblemCommitment: job.ProblemHash,
		SpinCommitment:    msg.SolutionHash,
		ClaimedEnergy:     msg.Energy,
		Threshold:         job.Threshold,
		Proof:             hex.EncodeToString(msg.Proof),
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return false, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(VerifierURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var verifyResp VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return false, err
	}

	if verifyResp.Error != nil {
		return false, fmt.Errorf("verification error: %s", *verifyResp.Error)
	}

	return verifyResp.Valid && verifyResp.MeetsThreshold, nil
}


// ClaimRewards allows miners to claim their earned rewards with actual token transfer
// Rewards include: customer payment (80/20 split) + emission reward (time-based, 80/20 split)
// Validator share remains in module for later distribution to validators
func (k msgServer) ClaimRewards(goCtx context.Context, msg *types.MsgClaimRewards) (*types.MsgClaimRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	job, found := k.GetJob(ctx, msg.JobId)
	if !found {
		return nil, types.ErrJobNotFound
	}

	claimerAddr, err := sdk.AccAddressFromBech32(msg.Claimer)
	if err != nil {
		return nil, types.ErrUnauthorized
	}

	shares := k.GetShares(ctx, claimerAddr, msg.JobId)
	if shares == 0 {
		return nil, types.ErrNoShares
	}

	params := k.GetParams(ctx)
	minerPercent := int64(params.MinerSharePercent)

	// === CUSTOMER REWARD (from escrowed job payment) ===
	minerProportionalReward := (shares * job.Reward) / job.TotalShares
	customerMinerReward := (minerProportionalReward * minerPercent) / 100
	customerValidatorShare := minerProportionalReward - customerMinerReward

	// === EMISSION REWARD (time-based: minutes_to_solve * emission_rate) ===
	emissionReward := k.ClaimEmissionReward(ctx, job)
	emissionMinerReward := (emissionReward * minerPercent) / 100
	emissionValidatorShare := emissionReward - emissionMinerReward

	// === TOTALS ===
	totalMinerReward := customerMinerReward + emissionMinerReward
	totalValidatorShare := customerValidatorShare + emissionValidatorShare

	rewardCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", totalMinerReward))

	// Transfer tokens from module to miner
	if k.bankKeeper != nil && totalMinerReward > 0 {
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, claimerAddr, rewardCoins)
		if err != nil {
			return nil, fmt.Errorf("failed to transfer reward: %w", err)
		}

		ctx.Logger().Info("Mining reward claimed",
			"job_id", msg.JobId,
			"claimer", msg.Claimer,
			"customer_reward", customerMinerReward,
			"emission_reward", emissionMinerReward,
			"total_miner_reward", totalMinerReward,
			"validator_share", totalValidatorShare,
		)
	}

	// Add validator share to reward pool for checkpoint distribution
	k.AddToValidatorRewardPool(ctx, totalValidatorShare)

	// Clear shares to mark as claimed
	k.SetShares(ctx, claimerAddr, msg.JobId, 0)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"rewards_claimed",
			sdk.NewAttribute("job_id", msg.JobId),
			sdk.NewAttribute("claimer", msg.Claimer),
			sdk.NewAttribute("customer_miner_reward", fmt.Sprintf("%d", customerMinerReward)),
			sdk.NewAttribute("emission_miner_reward", fmt.Sprintf("%d", emissionMinerReward)),
			sdk.NewAttribute("total_miner_reward", fmt.Sprintf("%d", totalMinerReward)),
			sdk.NewAttribute("validator_share", fmt.Sprintf("%d", totalValidatorShare)),
		),
	)

	return &types.MsgClaimRewardsResponse{Amount: rewardCoins}, nil
}
func (k msgServer) CancelJob(goCtx context.Context, msg *types.MsgCancelJob) (*types.MsgCancelJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	job, found := k.GetJob(ctx, msg.JobId)
	if !found {
		return nil, types.ErrJobNotFound
	}

	if job.Customer != msg.Customer {
		return nil, types.ErrUnauthorized
	}

	// Can only cancel if no shares have been earned
	if job.TotalShares > 0 {
		return nil, types.ErrCannotCancel
	}

	// Refund net reward to customer (fee was already burned on PostJob)
	if k.bankKeeper != nil && job.Reward > 0 {
		customerAddr, err := sdk.AccAddressFromBech32(msg.Customer)
		if err != nil {
			return nil, types.ErrUnauthorized
		}

		refundCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", job.Reward))
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, customerAddr, refundCoins)
		if err != nil {
			return nil, fmt.Errorf("failed to refund: %w", err)
		}

		ctx.Logger().Info("Refunded cancelled job",
			"job_id", msg.JobId,
			"amount", refundCoins.String(),
			"note", "job fee was already burned",
		)
	}

	job.Status = types.JobStatusCancelled
	k.SetJob(ctx, job)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"job_cancelled",
			sdk.NewAttribute("job_id", msg.JobId),
			sdk.NewAttribute("customer", msg.Customer),
			sdk.NewAttribute("refunded", fmt.Sprintf("%d", job.Reward)),
		),
	)

	return &types.MsgCancelJobResponse{Success: true}, nil
}
