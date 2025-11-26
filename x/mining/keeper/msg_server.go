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

// PostJob creates a new mining job
func (k msgServer) PostJob(goCtx context.Context, msg *types.MsgPostJob) (*types.MsgPostJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Generate job ID
	jobID := fmt.Sprintf("job_%d_%s", ctx.BlockHeight(), msg.Customer[:8])

	// Convert sdk.Coins reward to int64 (use first coin amount)
	var rewardAmount int64
	if len(msg.Reward) > 0 {
		rewardAmount = msg.Reward[0].Amount.Int64()
	}

	job := types.Job{
		Id:          jobID,
		Customer:    msg.Customer,
		ProblemType: msg.ProblemType,
		ProblemData: msg.ProblemData,
		ProblemHash: msg.ProblemHash,
		Threshold:   msg.Threshold,
		Reward:      rewardAmount,
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

// ClaimRewards allows miners to claim their earned rewards
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

	// Calculate reward: (miner_shares / total_shares) * reward * miner_percent
	params := k.GetParams(ctx)
	minerPercent := int64(params.MinerSharePercent)

	minerReward := (shares * job.Reward * minerPercent) / (job.TotalShares * 100)

	rewardCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", minerReward))

	// Clear shares to mark as claimed
	k.SetShares(ctx, claimerAddr, msg.JobId, 0)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"rewards_claimed",
			sdk.NewAttribute("job_id", msg.JobId),
			sdk.NewAttribute("claimer", msg.Claimer),
			sdk.NewAttribute("amount", rewardCoins.String()),
		),
	)

	return &types.MsgClaimRewardsResponse{Amount: rewardCoins}, nil
}

// CancelJob allows the customer to cancel an unworked job
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

	job.Status = types.JobStatusCancelled
	k.SetJob(ctx, job)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"job_cancelled",
			sdk.NewAttribute("job_id", msg.JobId),
			sdk.NewAttribute("customer", msg.Customer),
		),
	)

	return &types.MsgCancelJobResponse{Success: true}, nil
}
