package keeper

import (
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"nexus/x/mining/types"
)

const (
	QueryParams       = "params"
	QueryActiveJob    = "active-job"
	QueryQueueStatus  = "queue-status"
	QueryEmissionInfo = "emission-info"
	QueryJobs         = "jobs"
)

type LegacyQuerier func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error)

func NewQuerier(k Keeper) LegacyQuerier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case QueryParams:
			return queryParams(ctx, k)
		case QueryActiveJob:
			return queryActiveJob(ctx, k)
		case QueryQueueStatus:
			return queryQueueStatus(ctx, k)
		case QueryEmissionInfo:
			return queryEmissionInfo(ctx, k)
		case QueryJobs:
			return queryJobs(ctx, k)
		default:
			return nil, fmt.Errorf("unknown query path: %s", path[0])
		}
	}
}

func queryParams(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)
	return json.Marshal(params)
}

func queryActiveJob(ctx sdk.Context, k Keeper) ([]byte, error) {
	currentJobID := k.GetCurrentJobID(ctx)
	if currentJobID == "" {
		return json.Marshal(map[string]interface{}{
			"job":            nil,
			"time_remaining": 0,
			"problem_size":   k.GetCurrentProblemSize(ctx),
		})
	}

	job, found := k.GetJob(ctx, currentJobID)
	if !found {
		return json.Marshal(map[string]interface{}{
			"job":            nil,
			"time_remaining": 0,
			"problem_size":   k.GetCurrentProblemSize(ctx),
		})
	}

	timeRemaining := job.Deadline - ctx.BlockTime().Unix()
	if timeRemaining < 0 {
		timeRemaining = 0
	}

	return json.Marshal(map[string]interface{}{
		"job":            job,
		"time_remaining": timeRemaining,
		"problem_size":   k.GetCurrentProblemSize(ctx),
	})
}

func queryQueueStatus(ctx sdk.Context, k Keeper) ([]byte, error) {
	paidQueueLength := k.GetPaidJobQueueLength(ctx)
	publicQueueLength := k.GetPublicJobQueueLength(ctx)

	var nextPaidJobs []types.QueuedJobInfo
	paidQueue := k.GetPaidJobQueue(ctx)
	for i := 0; i < len(paidQueue) && i < 5; i++ {
		entry := paidQueue[i]
		job, found := k.GetJob(ctx, entry.JobID)
		if found {
			nextPaidJobs = append(nextPaidJobs, types.QueuedJobInfo{
				JobId:       job.Id,
				Customer:    job.Customer,
				PriorityFee: job.PriorityFee,
				Reward:      job.Reward,
			})
		}
	}

	publicQueue := k.GetPublicJobQueue(ctx)
	nextPublicJobs := publicQueue
	if len(nextPublicJobs) > 5 {
		nextPublicJobs = nextPublicJobs[:5]
	}

	return json.Marshal(map[string]interface{}{
		"paid_queue_length":   paidQueueLength,
		"public_queue_length": publicQueueLength,
		"next_paid_jobs":      nextPaidJobs,
		"next_public_jobs":    nextPublicJobs,
	})
}

func queryEmissionInfo(ctx sdk.Context, k Keeper) ([]byte, error) {
	currentEpoch := k.GetCurrentEpoch(ctx)
	emissionRate := k.GetCurrentEmissionRate(ctx)
	emissionEscrow := k.GetEmissionEscrow(ctx)
	genesisMinute := k.GetGenesisMinute(ctx)
	currentMinute := ctx.BlockTime().Unix() / 60

	minutesSinceGenesis := currentMinute - genesisMinute
	epochDuration := int64(EpochDurationMinutes)

	var minutesIntoEpoch int64
	var minutesUntilNext int64

	if currentEpoch < 8 {
		epochStartMinute := int64(currentEpoch-1) * epochDuration
		minutesIntoEpoch = minutesSinceGenesis - epochStartMinute
		minutesUntilNext = epochDuration - minutesIntoEpoch
	} else {
		epochStartMinute := int64(7) * epochDuration
		minutesIntoEpoch = minutesSinceGenesis - epochStartMinute
		minutesUntilNext = -1
	}

	var nextEpochRate int64
	if currentEpoch < 8 {
		epochs := GetEmissionEpochs()
		if currentEpoch < len(epochs) {
			nextEpochRate = (BaseEmissionPerMinute * epochs[currentEpoch].RatePercent) / 1000
		}
	} else {
		nextEpochRate = emissionRate
	}

	return json.Marshal(map[string]interface{}{
		"current_epoch":      currentEpoch,
		"emission_rate":      emissionRate,
		"emission_escrow":    emissionEscrow,
		"genesis_time":       genesisMinute * 60,
		"minutes_into_epoch": minutesIntoEpoch,
		"minutes_until_next": minutesUntilNext,
		"epoch_duration":     epochDuration,
		"next_epoch_rate":    nextEpochRate,
	})
}

func queryJobs(ctx sdk.Context, k Keeper) ([]byte, error) {
	var jobs []types.Job
	k.IterateJobs(ctx, func(job types.Job) bool {
		jobs = append(jobs, job)
		return false
	})
	return json.Marshal(jobs)
}