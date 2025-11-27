package keeper

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

func (k Keeper) GetDockingJob(ctx sdk.Context, jobId string) (types.DockingJob, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.DockingJobKeyPrefix, []byte(jobId)...)
	bz := store.Get(key)
	if bz == nil {
		return types.DockingJob{}, false
	}
	var job types.DockingJob
	k.cdc.MustUnmarshal(bz, &job)
	return job, true
}

func (k Keeper) SetDockingJob(ctx sdk.Context, job types.DockingJob) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.DockingJobKeyPrefix, []byte(job.Id)...)
	bz := k.cdc.MustMarshal(&job)
	store.Set(key, bz)
}

func (k Keeper) GetDockingResult(ctx sdk.Context, jobId, ligandId string) (types.DockingResult, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.DockingResultKeyPrefix, []byte(fmt.Sprintf("%s_%s", jobId, ligandId))...)
	bz := store.Get(key)
	if bz == nil {
		return types.DockingResult{}, false
	}
	var result types.DockingResult
	k.cdc.MustUnmarshal(bz, &result)
	return result, true
}

func (k Keeper) SetDockingResult(ctx sdk.Context, result types.DockingResult) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.DockingResultKeyPrefix, []byte(fmt.Sprintf("%s_%s", result.JobId, result.LigandId))...)
	bz := k.cdc.MustMarshal(&result)
	store.Set(key, bz)
}

func (k Keeper) GetDockingClaim(ctx sdk.Context, miner sdk.AccAddress, jobId string) (types.DockingClaim, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.DockingClaimKeyPrefix, append(miner.Bytes(), []byte(jobId)...)...)
	bz := store.Get(key)
	if bz == nil {
		return types.DockingClaim{}, false
	}
	var claim types.DockingClaim
	k.cdc.MustUnmarshal(bz, &claim)
	return claim, true
}

func (k Keeper) SetDockingClaim(ctx sdk.Context, claim types.DockingClaim) {
	store := ctx.KVStore(k.storeKey)
	minerAddr, _ := sdk.AccAddressFromBech32(claim.Miner)
	key := append(types.DockingClaimKeyPrefix, append(minerAddr.Bytes(), []byte(claim.JobId)...)...)
	bz := k.cdc.MustMarshal(&claim)
	store.Set(key, bz)
}

func (k Keeper) GetActiveDockingJob(ctx sdk.Context) (types.DockingJob, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("active_docking_job"))
	if bz == nil {
		return types.DockingJob{}, false
	}
	return k.GetDockingJob(ctx, string(bz))
}

func (k Keeper) SetActiveDockingJob(ctx sdk.Context, jobId string) {
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte("active_docking_job"), []byte(jobId))
}

func (k Keeper) GetDockingJobCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("docking_job_count"))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) IncrementDockingJobCount(ctx sdk.Context) uint64 {
	count := k.GetDockingJobCount(ctx) + 1
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte("docking_job_count"), uint64ToBytes(count))
	return count
}

func (k Keeper) GetMinerDockingRewards(ctx sdk.Context, miner sdk.AccAddress) int64 {
	store := ctx.KVStore(k.storeKey)
	key := append([]byte("miner_dock_rewards_"), miner.Bytes()...)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) AddMinerDockingRewards(ctx sdk.Context, miner sdk.AccAddress, amount int64) {
	current := k.GetMinerDockingRewards(ctx, miner)
	store := ctx.KVStore(k.storeKey)
	key := append([]byte("miner_dock_rewards_"), miner.Bytes()...)
	store.Set(key, uint64ToBytes(uint64(current+amount)))
}
