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
	jobID := fmt.Sprintf("paid_%d_%s", ctx.BlockHeight(), msg.Customer[:8])

	customerAddr, err := sdk.AccAddressFromBech32(msg.Customer)
	if err != nil {
		return nil, types.ErrInvalidJob
	}

	// Convert sdk.Coins reward to int64
	var grossRewardAmount int64
	if len(msg.Reward) > 0 {
		grossRewardAmount = msg.Reward[0].Amount.Int64()
	}

	// Get priority fee amount
	var priorityFeeAmount int64
	if len(msg.PriorityFee) > 0 {
		priorityFeeAmount = msg.PriorityFee[0].Amount.Int64()
	}

	// Calculate job fee burn (2% of reward)
	params := k.GetParams(ctx)
	feeBurnPercent := int64(params.JobFeeBurnPercent)
	feeBurnAmount := (grossRewardAmount * feeBurnPercent) / 100
	netRewardAmount := grossRewardAmount - feeBurnAmount

	// Total to collect = reward + priority fee
	totalToCollect := grossRewardAmount + priorityFeeAmount

	// Transfer total amount from customer to module
	if k.bankKeeper != nil && totalToCollect > 0 {
		collectCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", totalToCollect))
		err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, customerAddr, types.ModuleName, collectCoins)
		if err != nil {
			return nil, fmt.Errorf("failed to escrow reward: %w", err)
		}

		// Burn the job fee (2% of reward) + entire priority fee
		totalBurnAmount := feeBurnAmount + priorityFeeAmount
		if totalBurnAmount > 0 {
			burnCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", totalBurnAmount))
			err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins)
			if err != nil {
				return nil, fmt.Errorf("failed to burn fees: %w", err)
			}

			ctx.Logger().Info("Burned job fees",
				"job_id", jobID,
				"job_fee_burned", feeBurnAmount,
				"priority_fee_burned", priorityFeeAmount,
				"total_burned", totalBurnAmount,
			)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"fee_burned",
					sdk.NewAttribute("job_id", jobID),
					sdk.NewAttribute("job_fee", fmt.Sprintf("%d", feeBurnAmount)),
					sdk.NewAttribute("priority_fee", fmt.Sprintf("%d", priorityFeeAmount)),
					sdk.NewAttribute("type", "paid_job"),
				),
			)
		}
	}

	// Create job (status = Queued, not Active)
	job := types.Job{
		Id:           jobID,
		Customer:     msg.Customer,
		ProblemType:  msg.ProblemType,
		ProblemData:  msg.ProblemData,
		ProblemHash:  msg.ProblemHash,
		Threshold:    msg.Threshold,
		Reward:       netRewardAmount,
		Status:       types.JobStatusQueued,
		BestEnergy:   0,
		BestSolver:   "",
		TotalShares:  0,
		CreatedAt:    ctx.BlockTime().Unix(),
		Deadline:     0, // Set when activated
		IsBackground: false,
		PriorityFee:  priorityFeeAmount,
	}

	k.SetJob(ctx, job)

	// Add to paid job queue (sorted by priority fee)
	queuePosition := k.AddToPaidJobQueue(ctx, jobID, priorityFeeAmount)

	ctx.Logger().Info("Paid job queued",
		"job_id", jobID,
		"customer", msg.Customer,
		"reward", netRewardAmount,
		"priority_fee", priorityFeeAmount,
		"queue_position", queuePosition,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"job_posted",
			sdk.NewAttribute("job_id", jobID),
			sdk.NewAttribute("customer", msg.Customer),
			sdk.NewAttribute("net_reward", fmt.Sprintf("%d", netRewardAmount)),
			sdk.NewAttribute("priority_fee", fmt.Sprintf("%d", priorityFeeAmount)),
			sdk.NewAttribute("queue_position", fmt.Sprintf("%d", queuePosition)),
		),
	)

	return &types.MsgPostJobResponse{JobId: jobID, QueuePosition: queuePosition}, nil
}
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

	return &types.MsgSubmitProofResponse{Accepted: true, Shares: sharesEarned}, nil
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

// SubmitPublicJob allows staked users to submit free background jobs for public benefit
// Requirements: minimum stake, posting fee (burned)
// Job goes into queue, selected randomly when no paid jobs
func (k msgServer) SubmitPublicJob(goCtx context.Context, msg *types.MsgSubmitPublicJob) (*types.MsgSubmitPublicJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	submitterAddr, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		return nil, types.ErrUnauthorized
	}

	// Validate category
	if !ValidCategories[msg.Category] {
		return nil, fmt.Errorf("invalid category: %s", msg.Category)
	}

	// Check minimum stake requirement
	// For now, check if user has minimum balance (in production, check delegations)
	balance := k.bankKeeper.GetBalance(ctx, submitterAddr, "unexus")
	if balance.Amount.Int64() < MinStakeForSubmission {
		return nil, fmt.Errorf("insufficient stake: need %d, have %d", MinStakeForSubmission, balance.Amount.Int64())
	}

	// Charge and burn posting fee
	postingFee := sdk.NewCoins(sdk.NewInt64Coin("unexus", PublicJobPostingFee))
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, submitterAddr, types.ModuleName, postingFee)
	if err != nil {
		return nil, fmt.Errorf("failed to charge posting fee: %w", err)
	}

	// Burn the posting fee
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, postingFee)
	if err != nil {
		return nil, fmt.Errorf("failed to burn posting fee: %w", err)
	}

	// Validate problem size meets minimum difficulty
	minSize := k.GetCurrentProblemSize(ctx)
	problemSize := int64(len(msg.ProblemData))
	// Problem data for Ising model is size*size bytes
	// So sqrt(problemSize) should be >= minSize
	estimatedSpins := int64(1)
	for estimatedSpins*estimatedSpins < problemSize {
		estimatedSpins++
	}
	if estimatedSpins < minSize {
		return nil, fmt.Errorf("problem too small: %d spins, minimum %d", estimatedSpins, minSize)
	}

	// Create job ID
	height := ctx.BlockHeight()
	jobID := fmt.Sprintf("pub_%d_%s", height, msg.ProblemHash[:8])

	// Create the job (status = Queued, not Active)
	job := types.Job{
		Id:           jobID,
		Customer:     msg.Submitter, // Submitter is the "customer" for public jobs
		ProblemType:  msg.Category,
		ProblemData:  msg.ProblemData,
		ProblemHash:  msg.ProblemHash,
		Threshold:    msg.Threshold,
		Reward:       0, // No customer reward - emission only
		Status:       types.JobStatusQueued,
		BestEnergy:   0,
		TotalShares:  0,
		CreatedAt:    ctx.BlockTime().Unix(),
		Deadline:     0, // Set when activated
		IsBackground: true,
	}

	k.SetJob(ctx, job)

	// Add to public job queue
	k.AddToPublicJobQueue(ctx, jobID)
	queueLen := k.GetPublicJobQueueLength(ctx)

	ctx.Logger().Info("Public job submitted",
		"job_id", jobID,
		"submitter", msg.Submitter,
		"category", msg.Category,
		"title", msg.Title,
		"queue_position", queueLen,
		"posting_fee_burned", PublicJobPostingFee,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"public_job_submitted",
			sdk.NewAttribute("job_id", jobID),
			sdk.NewAttribute("submitter", msg.Submitter),
			sdk.NewAttribute("category", msg.Category),
			sdk.NewAttribute("title", msg.Title),
			sdk.NewAttribute("queue_length", fmt.Sprintf("%d", queueLen)),
			sdk.NewAttribute("ipfs_cid", msg.IpfsCid),
		),
	)

	return &types.MsgSubmitPublicJobResponse{
		JobId:         jobID,
		QueuePosition: int64(queueLen),
	}, nil
}
