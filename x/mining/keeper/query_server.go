package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

type queryServer struct {
	Keeper
}

func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

func (q queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (q queryServer) Job(goCtx context.Context, req *types.QueryJobRequest) (*types.QueryJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	job, found := q.Keeper.GetJob(ctx, req.JobId)
	if !found {
		return nil, types.ErrJobNotFound
	}
	return &types.QueryJobResponse{Job: job}, nil
}

func (q queryServer) Jobs(goCtx context.Context, req *types.QueryJobsRequest) (*types.QueryJobsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	var jobs []types.Job
	q.Keeper.IterateJobs(ctx, func(job types.Job) bool {
		// Filter by status if specified (999 means all)
		if req.Status == 999 || uint32(job.Status) == req.Status {
			jobs = append(jobs, job)
		}
		return false
	})
	
	return &types.QueryJobsResponse{Jobs: jobs}, nil
}

func (q queryServer) MinerShares(goCtx context.Context, req *types.QueryMinerSharesRequest) (*types.QueryMinerSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	miner, err := sdk.AccAddressFromBech32(req.Miner)
	if err != nil {
		return nil, err
	}
	shares := q.Keeper.GetShares(ctx, miner, req.JobId)
	return &types.QueryMinerSharesResponse{
		Miner:  req.Miner,
		JobId:  req.JobId,
		Shares: shares,
	}, nil
}

func (q queryServer) MinerStats(goCtx context.Context, req *types.QueryMinerStatsRequest) (*types.QueryMinerStatsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	minerAddr, err := sdk.AccAddressFromBech32(req.MinerAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid miner address: %w", err)
	}

	var totalShares int64
	var jobsParticipated int64
	var pendingRewards int64
	var activeJobs []types.MinerJobInfo

	// Iterate through all jobs to find miner's participation
	q.Keeper.IterateJobs(ctx, func(job types.Job) bool {
		shares := q.Keeper.GetShares(ctx, minerAddr, job.Id)
		if shares > 0 {
			totalShares += shares
			jobsParticipated++

			// Calculate pending rewards for this job
			if job.Status == types.JobStatusCompleted || job.Status == types.JobStatusActive {
				params := q.Keeper.GetParams(ctx)
				minerPercent := int64(params.MinerSharePercent)
				
				// Customer reward portion
				minerProportionalReward := (shares * job.Reward) / job.TotalShares
				customerMinerReward := (minerProportionalReward * minerPercent) / 100
				
				// Emission reward portion
				emissionReward := q.Keeper.CalculateEmissionReward(ctx, job)
				emissionMinerReward := (emissionReward * minerPercent) / 100
				
				pendingRewards += customerMinerReward + emissionMinerReward
			}

			// Add to active jobs list if job is still active
			if job.Status == types.JobStatusActive {
				activeJobs = append(activeJobs, types.MinerJobInfo{
					JobId:       job.Id,
					Shares:      shares,
					BestEnergy:  job.BestEnergy,
					TotalShares: job.TotalShares,
				})
			}
		}
		return false
	})

	return &types.QueryMinerStatsResponse{
		MinerAddress:     req.MinerAddress,
		TotalShares:      totalShares,
		PendingRewards:   sdk.NewCoins(sdk.NewInt64Coin("unexus", pendingRewards)),
		JobsParticipated: jobsParticipated,
		TotalClaimed:     sdk.NewCoins(), // TODO: Track claimed rewards
		ActiveJobs:       activeJobs,
	}, nil
}

func (q queryServer) ActiveJob(goCtx context.Context, req *types.QueryActiveJobRequest) (*types.QueryActiveJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	currentJobID := q.Keeper.GetCurrentJobID(ctx)
	if currentJobID == "" {
		return &types.QueryActiveJobResponse{
			Job:           nil,
			TimeRemaining: 0,
			ProblemSize:   q.Keeper.GetCurrentProblemSize(ctx),
		}, nil
	}

	job, found := q.Keeper.GetJob(ctx, currentJobID)
	if !found {
		return &types.QueryActiveJobResponse{
			Job:           nil,
			TimeRemaining: 0,
			ProblemSize:   q.Keeper.GetCurrentProblemSize(ctx),
		}, nil
	}

	timeRemaining := job.Deadline - ctx.BlockTime().Unix()
	if timeRemaining < 0 {
		timeRemaining = 0
	}

	return &types.QueryActiveJobResponse{
		Job:           &job,
		TimeRemaining: timeRemaining,
		ProblemSize:   q.Keeper.GetCurrentProblemSize(ctx),
	}, nil
}

func (q queryServer) QueueStatus(goCtx context.Context, req *types.QueryQueueStatusRequest) (*types.QueryQueueStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	paidQueueLength := int64(q.Keeper.GetPaidJobQueueLength(ctx))
	publicQueueLength := int64(q.Keeper.GetPublicJobQueueLength(ctx))

	// Get next paid jobs (up to 5)
	var nextPaidJobs []types.QueuedJobInfo
	paidQueue := q.Keeper.GetPaidJobQueue(ctx)
	for i := 0; i < len(paidQueue) && i < 5; i++ {
		entry := paidQueue[i]
		job, found := q.Keeper.GetJob(ctx, entry.JobID)
		if found {
			nextPaidJobs = append(nextPaidJobs, types.QueuedJobInfo{
				JobId:       job.Id,
				Customer:    job.Customer,
				PriorityFee: job.PriorityFee,
				Reward:      job.Reward,
			})
		}
	}

	// Get next public jobs (up to 5)
	publicQueue := q.Keeper.GetPublicJobQueue(ctx)
	nextPublicJobs := publicQueue
	if len(nextPublicJobs) > 5 {
		nextPublicJobs = nextPublicJobs[:5]
	}

	return &types.QueryQueueStatusResponse{
		PaidQueueLength:   paidQueueLength,
		PublicQueueLength: publicQueueLength,
		NextPaidJobs:      nextPaidJobs,
		NextPublicJobs:    nextPublicJobs,
	}, nil
}

func (q queryServer) EmissionInfo(goCtx context.Context, req *types.QueryEmissionInfoRequest) (*types.QueryEmissionInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	currentEpoch := q.Keeper.GetCurrentEpoch(ctx)
	emissionRate := q.Keeper.GetCurrentEmissionRate(ctx)
	emissionEscrow := q.Keeper.GetEmissionEscrow(ctx)
	genesisMinute := q.Keeper.GetGenesisMinute(ctx)
	currentMinute := ctx.BlockTime().Unix() / 60

	// Calculate minutes into current epoch
	minutesSinceGenesis := currentMinute - genesisMinute
	epochDuration := int64(EpochDurationMinutes)
	
	var minutesIntoEpoch int64
	var minutesUntilNext int64
	
	if currentEpoch < 8 {
		epochStartMinute := int64(currentEpoch-1) * epochDuration
		minutesIntoEpoch = minutesSinceGenesis - epochStartMinute
		minutesUntilNext = epochDuration - minutesIntoEpoch
	} else {
		// Perpetual epoch
		epochStartMinute := int64(7) * epochDuration
		minutesIntoEpoch = minutesSinceGenesis - epochStartMinute
		minutesUntilNext = -1 // No next epoch
	}

	// Calculate next epoch rate
	var nextEpochRate int64
	if currentEpoch < 8 {
		epochs := GetEmissionEpochs()
		if currentEpoch < len(epochs) {
			nextEpochRate = (BaseEmissionPerMinute * epochs[currentEpoch].RatePercent) / 1000
		}
	} else {
		nextEpochRate = emissionRate // Same rate perpetually
	}

	return &types.QueryEmissionInfoResponse{
		CurrentEpoch:     int32(currentEpoch),
		EmissionRate:     emissionRate,
		EmissionEscrow:   emissionEscrow,
		GenesisTime:      genesisMinute * 60,
		MinutesIntoEpoch: minutesIntoEpoch,
		MinutesUntilNext: minutesUntilNext,
		EpochDuration:    epochDuration,
		NextEpochRate:    nextEpochRate,
	}, nil
}

func (q queryServer) ValidatorMiningRecord(goCtx context.Context, req *types.QueryValidatorMiningRecordRequest) (*types.QueryValidatorMiningRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}
	record, found := q.Keeper.GetValidatorMiningRecord(ctx, valAddr)
	if !found {
		return nil, types.ErrValidatorNotFound
	}
	return &types.QueryValidatorMiningRecordResponse{Record: record}, nil
}

func (q queryServer) Checkpoint(goCtx context.Context, req *types.QueryCheckpointRequest) (*types.QueryCheckpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	checkpoint, found := q.Keeper.GetCheckpoint(ctx, req.Id)
	if !found {
		return nil, types.ErrCheckpointNotFound
	}
	return &types.QueryCheckpointResponse{Checkpoint: checkpoint}, nil
}

func (q queryServer) LatestCheckpoint(goCtx context.Context, req *types.QueryLatestCheckpointRequest) (*types.QueryLatestCheckpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	lastID := q.Keeper.GetLastCheckpointID(ctx)
	if lastID == 0 {
		return nil, types.ErrCheckpointNotFound
	}
	checkpoint, found := q.Keeper.GetCheckpoint(ctx, lastID)
	if !found {
		return nil, types.ErrCheckpointNotFound
	}
	return &types.QueryLatestCheckpointResponse{Checkpoint: checkpoint}, nil
}
