package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/docking"
	"nexus/x/mining/types"
)

// CreateDockingJob creates a new molecular docking job
func (k msgServer) CreateDockingJob(goCtx context.Context, msg *types.MsgCreateDockingJob) (*types.MsgCreateDockingJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	count := k.IncrementDockingJobCount(ctx)
	jobId := fmt.Sprintf("dock_%d_%d", ctx.BlockHeight(), count)

	job := types.DockingJob{
		Id:            jobId,
		TargetHash:    msg.TargetHash,
		ProteinPDB:    msg.ProteinPDB,
		TotalLigands:  msg.TotalLigands,
		DockedCount:   0,
		HitCount:      0,
		CenterX:       msg.CenterX,
		CenterY:       msg.CenterY,
		CenterZ:       msg.CenterZ,
		SizeX:         30,
		SizeY:         30,
		SizeZ:         30,
		IsBackground:  msg.IsBackground,
		Status:        types.DockingJobStatusActive,
		CreatedAt:     ctx.BlockTime().Unix(),
		Deadline:      ctx.BlockTime().Add(7 * 24 * time.Hour).Unix(),
		NextLigandIdx: 0,
	}

	// Handle paid jobs - escrow rewards
	if !msg.IsBackground && len(msg.Reward) > 0 {
		creatorAddr, _ := sdk.AccAddressFromBech32(msg.Creator)
		if k.bankKeeper != nil {
			err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, msg.Reward)
			if err != nil {
				return nil, fmt.Errorf("failed to escrow reward: %w", err)
			}
		}
		job.RewardPool = msg.Reward[0].Amount.Int64()
	}

	k.SetDockingJob(ctx, job)

	// If background job, set as active
	if msg.IsBackground {
		k.SetActiveDockingJob(ctx, jobId)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"docking_job_created",
			sdk.NewAttribute("job_id", jobId),
			sdk.NewAttribute("target_hash", msg.TargetHash),
			sdk.NewAttribute("total_ligands", fmt.Sprintf("%d", msg.TotalLigands)),
			sdk.NewAttribute("is_background", fmt.Sprintf("%t", msg.IsBackground)),
		),
	)

	return &types.MsgCreateDockingJobResponse{JobId: jobId}, nil
}

// ClaimDockingJob assigns a batch of ligands to a miner
func (k msgServer) ClaimDockingJob(goCtx context.Context, msg *types.MsgClaimDockingJob) (*types.MsgClaimDockingJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var job types.DockingJob
	var found bool

	if msg.JobId != "" {
		job, found = k.GetDockingJob(ctx, msg.JobId)
	} else {
		// Get active background job
		job, found = k.GetActiveDockingJob(ctx)
	}

	if !found {
		return nil, types.ErrJobNotFound
	}

	if job.Status != types.DockingJobStatusActive {
		return nil, types.ErrJobNotActive
	}

	// Check if all ligands are assigned
	if job.NextLigandIdx >= job.TotalLigands {
		return nil, fmt.Errorf("all ligands assigned for job %s", job.Id)
	}

	// Assign batch
	batchSize := int64(msg.BatchSize)
	if batchSize <= 0 {
		batchSize = 100 // default batch
	}

	startLigand := job.NextLigandIdx
	endLigand := startLigand + batchSize
	if endLigand > job.TotalLigands {
		endLigand = job.TotalLigands
	}

	// Update job
	job.NextLigandIdx = endLigand
	k.SetDockingJob(ctx, job)

	// Record claim
	claim := types.DockingClaim{
		Miner:       msg.Miner,
		JobId:       job.Id,
		StartLigand: startLigand,
		EndLigand:   endLigand,
		ClaimedAt:   ctx.BlockTime().Unix(),
	}
	k.SetDockingClaim(ctx, claim)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"docking_job_claimed",
			sdk.NewAttribute("job_id", job.Id),
			sdk.NewAttribute("miner", msg.Miner),
			sdk.NewAttribute("start_ligand", fmt.Sprintf("%d", startLigand)),
			sdk.NewAttribute("end_ligand", fmt.Sprintf("%d", endLigand)),
		),
	)

	return &types.MsgClaimDockingJobResponse{
		JobId:       job.Id,
		ProteinPDB:  job.ProteinPDB,
		TargetHash:  job.TargetHash,
		StartLigand: startLigand,
		EndLigand:   endLigand,
		CenterX:     job.CenterX,
		CenterY:     job.CenterY,
		CenterZ:     job.CenterZ,
		SizeX:       job.SizeX,
		SizeY:       job.SizeY,
		SizeZ:       job.SizeZ,
	}, nil
}

// SubmitDockingResult records a single ligand docking result
func (k msgServer) SubmitDockingResult(goCtx context.Context, msg *types.MsgSubmitDockingResult) (*types.MsgSubmitDockingResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	job, found := k.GetDockingJob(ctx, msg.JobId)
	if !found {
		return nil, types.ErrJobNotFound
	}

	if job.Status != types.DockingJobStatusActive {
		return nil, types.ErrJobNotActive
	}

	// Check for duplicate
	_, exists := k.GetDockingResult(ctx, msg.JobId, msg.LigandId)
	if exists {
		return nil, fmt.Errorf("result already submitted for ligand %s", msg.LigandId)
	}

	// Calculate reward based on rotatable bonds
	reward := types.CalculateDockingReward(int(msg.RotatableBonds))

	// Check if this is a hit (binding score below threshold)
	isHit := msg.BindingScore < types.DockingHitThreshold

	// Record result
	result := types.DockingResult{
		Id:             fmt.Sprintf("%s_%s", msg.JobId, msg.LigandId),
		JobId:          msg.JobId,
		LigandId:       msg.LigandId,
		LigandSMILES:   msg.LigandSMILES,
		BindingScore:   msg.BindingScore,
		RotatableBonds: msg.RotatableBonds,
		Miner:          msg.Miner,
		Reward:         reward,
		IsHit:          isHit,
		BlockHeight:    ctx.BlockHeight(),
		SubmittedAt:    ctx.BlockTime().Unix(),
	}
	k.SetDockingResult(ctx, result)

	// Update job counters
	job.DockedCount++
	if isHit {
		job.HitCount++
	}

	// Check if job is complete
	if job.DockedCount >= job.TotalLigands {
		job.Status = types.DockingJobStatusCompleted
	}
	k.SetDockingJob(ctx, job)

	// Track miner rewards (for later distribution)
	minerAddr, _ := sdk.AccAddressFromBech32(msg.Miner)
	k.AddMinerDockingRewards(ctx, minerAddr, reward)

	// For background jobs, mint and distribute reward immediately
	if job.IsBackground && k.bankKeeper != nil {
		// Rewards tracked for later distribution via emission
		// Mint from emission (handled in ABCI) - for now just track
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"docking_result_submitted",
			sdk.NewAttribute("job_id", msg.JobId),
			sdk.NewAttribute("ligand_id", msg.LigandId),
			sdk.NewAttribute("miner", msg.Miner),
			sdk.NewAttribute("binding_score", fmt.Sprintf("%.2f", msg.BindingScore)),
			sdk.NewAttribute("reward", fmt.Sprintf("%d", reward)),
			sdk.NewAttribute("is_hit", fmt.Sprintf("%t", isHit)),
		),
	)

	if isHit {
		ctx.Logger().Info("Drug candidate found!",
			"job_id", msg.JobId,
			"ligand_id", msg.LigandId,
			"binding_score", msg.BindingScore,
			"miner", msg.Miner,
		)
	}

	return &types.MsgSubmitDockingResultResponse{
		Accepted: true,
		Reward:   reward,
		IsHit:    isHit,
	}, nil
}

// CreateBackgroundDockingJob auto-generates a docking job from AlphaFold + PubChem
func (k Keeper) CreateBackgroundDockingJob(ctx sdk.Context, targetIndex int) (string, error) {
	target := docking.GetNextBackgroundTarget(targetIndex)

	// Fetch protein from AlphaFold DB
	protein, err := docking.FetchProtein(target)
	if err != nil {
		return "", fmt.Errorf("failed to fetch protein %s: %w", target.UniprotID, err)
	}

	count := k.IncrementDockingJobCount(ctx)
	jobId := fmt.Sprintf("bg_%s_%d", target.UniprotID, count)

	job := types.DockingJob{
		Id:            jobId,
		ProteinId:     target.UniprotID,
		TargetHash:    protein.PDBHash,
		ProteinPDB:    protein.PDBContent,
		TotalLigands:  10000, // Default batch size
		DockedCount:   0,
		HitCount:      0,
		CenterX:       protein.BindingSite.CenterX,
		CenterY:       protein.BindingSite.CenterY,
		CenterZ:       protein.BindingSite.CenterZ,
		SizeX:         30,
		SizeY:         30,
		SizeZ:         30,
		IsBackground:  true,
		Status:        types.DockingJobStatusActive,
		CreatedAt:     ctx.BlockTime().Unix(),
		Deadline:      ctx.BlockTime().Add(7 * 24 * time.Hour).Unix(),
		NextLigandIdx: 0,
		License:       protein.License,
	}

	k.SetDockingJob(ctx, job)
	k.SetActiveDockingJob(ctx, jobId)

	ctx.Logger().Info("Background docking job created",
		"job_id", jobId,
		"protein", target.Name,
		"disease", target.Disease,
		"license", protein.License,
	)

	return jobId, nil
}
