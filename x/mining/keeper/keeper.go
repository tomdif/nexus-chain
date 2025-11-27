package keeper

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"nexus/x/mining/types"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	memKey        storetypes.StoreKey
	stakingKeeper types.StakingKeeper
	bankKeeper    types.BankKeeper
	authority     string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	stakingKeeper types.StakingKeeper,
	bankKeeper types.BankKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		stakingKeeper: stakingKeeper,
		bankKeeper:    bankKeeper,
		authority:     authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("params"))
	if bz == nil {
		return types.DefaultParams()
	}
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set([]byte("params"), bz)
	return nil
}

func (k Keeper) GetJob(ctx sdk.Context, jobId string) (types.Job, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.JobKeyPrefix, []byte(jobId)...)
	bz := store.Get(key)
	if bz == nil {
		return types.Job{}, false
	}
	var job types.Job
	k.cdc.MustUnmarshal(bz, &job)
	return job, true
}

func (k Keeper) SetJob(ctx sdk.Context, job types.Job) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.JobKeyPrefix, []byte(job.Id)...)
	bz := k.cdc.MustMarshal(&job)
	store.Set(key, bz)
}

func (k Keeper) GetCheckpoint(ctx sdk.Context, id uint64) (types.Checkpoint, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.CheckpointKeyPrefix, uint64ToBytes(id)...)
	bz := store.Get(key)
	if bz == nil {
		return types.Checkpoint{}, false
	}
	var cp types.Checkpoint
	k.cdc.MustUnmarshal(bz, &cp)
	return cp, true
}

func (k Keeper) SetCheckpoint(ctx sdk.Context, cp types.Checkpoint) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.CheckpointKeyPrefix, uint64ToBytes(cp.Id)...)
	bz := k.cdc.MustMarshal(&cp)
	store.Set(key, bz)
}

func (k Keeper) GetLastCheckpointID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastCheckpointIDKey)
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) SetLastCheckpointID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastCheckpointIDKey, uint64ToBytes(id))
}

func (k Keeper) GetShares(ctx sdk.Context, miner sdk.AccAddress, jobId string) int64 {
	store := ctx.KVStore(k.storeKey)
	key := append(types.ShareKeyPrefix, append(miner.Bytes(), []byte(jobId)...)...)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) SetShares(ctx sdk.Context, miner sdk.AccAddress, jobId string, shares int64) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.ShareKeyPrefix, append(miner.Bytes(), []byte(jobId)...)...)
	store.Set(key, uint64ToBytes(uint64(shares)))
}

func (k Keeper) GetValidatorMiningRecord(ctx sdk.Context, valAddr sdk.ValAddress) (types.ValidatorMiningRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.ValidatorRecordKeyPrefix, valAddr.Bytes()...)
	bz := store.Get(key)
	if bz == nil {
		return types.ValidatorMiningRecord{}, false
	}
	var record types.ValidatorMiningRecord
	k.cdc.MustUnmarshal(bz, &record)
	return record, true
}

func (k Keeper) SetValidatorMiningRecord(ctx sdk.Context, record types.ValidatorMiningRecord) {
	store := ctx.KVStore(k.storeKey)
	valAddr, _ := sdk.ValAddressFromBech32(record.ValidatorAddr)
	key := append(types.ValidatorRecordKeyPrefix, valAddr.Bytes()...)
	bz := k.cdc.MustMarshal(&record)
	store.Set(key, bz)
}

func uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}


// ============================================
// Collaborative Mining Storage Methods
// ============================================

// GetWorkSubmission retrieves a work submission by ID
func (k Keeper) GetWorkSubmission(ctx sdk.Context, submissionId string) (types.WorkSubmission, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.WorkSubmissionKeyPrefix, []byte(submissionId)...)
	bz := store.Get(key)
	if bz == nil {
		return types.WorkSubmission{}, false
	}
	var submission types.WorkSubmission
	k.cdc.MustUnmarshal(bz, &submission)
	return submission, true
}

// SetWorkSubmission stores a work submission
func (k Keeper) SetWorkSubmission(ctx sdk.Context, submission types.WorkSubmission) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.WorkSubmissionKeyPrefix, []byte(submission.Id)...)
	bz := k.cdc.MustMarshal(&submission)
	store.Set(key, bz)
}

// GetWorkShares retrieves work shares for a miner on a job
func (k Keeper) GetWorkShares(ctx sdk.Context, miner sdk.AccAddress, jobId string) int64 {
	store := ctx.KVStore(k.storeKey)
	key := append(types.WorkShareKeyPrefix, append(miner.Bytes(), []byte(jobId)...)...)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

// SetWorkShares sets work shares for a miner on a job
func (k Keeper) SetWorkShares(ctx sdk.Context, miner sdk.AccAddress, jobId string, shares int64) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.WorkShareKeyPrefix, append(miner.Bytes(), []byte(jobId)...)...)
	store.Set(key, uint64ToBytes(uint64(shares)))
}

// GetBonusShares retrieves bonus shares for a miner on a job
func (k Keeper) GetBonusShares(ctx sdk.Context, miner sdk.AccAddress, jobId string) int64 {
	store := ctx.KVStore(k.storeKey)
	key := append(types.BonusShareKeyPrefix, append(miner.Bytes(), []byte(jobId)...)...)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

// SetBonusShares sets bonus shares for a miner on a job
func (k Keeper) SetBonusShares(ctx sdk.Context, miner sdk.AccAddress, jobId string, shares int64) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.BonusShareKeyPrefix, append(miner.Bytes(), []byte(jobId)...)...)
	store.Set(key, uint64ToBytes(uint64(shares)))
}

// GetJobEpoch retrieves the current epoch for a job
func (k Keeper) GetJobEpoch(ctx sdk.Context, jobId string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := append(types.EpochKeyPrefix, []byte(jobId)...)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// SetJobEpoch sets the current epoch for a job
func (k Keeper) SetJobEpoch(ctx sdk.Context, jobId string, epoch uint64) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.EpochKeyPrefix, []byte(jobId)...)
	store.Set(key, uint64ToBytes(epoch))
}

// IterateWorkSubmissions iterates over all work submissions for a job
func (k Keeper) IterateWorkSubmissions(ctx sdk.Context, jobId string, cb func(submission types.WorkSubmission) bool) {
	store := ctx.KVStore(k.storeKey)
	prefix := append(types.WorkSubmissionKeyPrefix, []byte(jobId)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var submission types.WorkSubmission
		k.cdc.MustUnmarshal(iterator.Value(), &submission)
		if cb(submission) {
			break
		}
	}
}
