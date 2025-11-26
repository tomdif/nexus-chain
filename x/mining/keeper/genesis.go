package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

// InitGenesis initializes the mining module's state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, gs types.GenesisState) {
	// Set params
	k.SetParams(ctx, gs.Params)

	// Set jobs
	for _, job := range gs.Jobs {
		k.SetJob(ctx, job)
	}

	// Set checkpoints
	for _, cp := range gs.Checkpoints {
		k.SetCheckpoint(ctx, cp)
	}
	k.SetLastCheckpointID(ctx, gs.LastCheckpointID)

	// Set validator reward pool
	k.SetValidatorRewardPool(ctx, gs.ValidatorRewardPool)

	// Set emission escrow
	k.SetEmissionEscrow(ctx, gs.EmissionEscrow)

	// Set difficulty/problem size
	k.SetCurrentProblemSize(ctx, gs.CurrentProblemSize)

	// Initialize active job count
	k.SetActiveJobCount(ctx, 0)

	// Log initialization
	k.Logger(ctx).Info("Mining module initialized",
		"params", gs.Params,
		"problem_size", gs.CurrentProblemSize,
	)
}

// ExportGenesis exports the mining module's state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	// Collect all jobs
	jobs := []types.Job{}
	k.IterateJobs(ctx, func(job types.Job) bool {
		jobs = append(jobs, job)
		return false
	})

	// Collect all checkpoints
	checkpoints := []types.Checkpoint{}
	lastCpID := k.GetLastCheckpointID(ctx)
	for i := uint64(1); i <= lastCpID; i++ {
		if cp, found := k.GetCheckpoint(ctx, i); found {
			checkpoints = append(checkpoints, cp)
		}
	}

	return &types.GenesisState{
		Params:              k.GetParams(ctx),
		Jobs:                jobs,
		Checkpoints:         checkpoints,
		ValidatorRewardPool: k.GetValidatorRewardPool(ctx),
		EmissionEscrow:      k.GetEmissionEscrow(ctx),
		LastCheckpointID:    lastCpID,
		CurrentProblemSize:  k.GetCurrentProblemSize(ctx),
		BackgroundJobCount:  k.GetBackgroundJobCount(ctx),
	}
}

// IterateJobs iterates over all jobs
func (k Keeper) IterateJobs(ctx sdk.Context, fn func(job types.Job) bool) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.JobKeyPrefix)
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var job types.Job
		k.cdc.MustUnmarshal(iterator.Value(), &job)
		if fn(job) {
			break
		}
	}
}
