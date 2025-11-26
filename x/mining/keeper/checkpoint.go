package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidatorRewardPoolKey is the key for storing accumulated validator rewards
var ValidatorRewardPoolKey = []byte("validator_reward_pool")

// GetValidatorRewardPool returns the accumulated validator rewards waiting for distribution
func (k Keeper) GetValidatorRewardPool(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ValidatorRewardPoolKey)
	if bz == nil {
		return 0
	}
	return int64(bytesToUint64(bz))
}

// SetValidatorRewardPool sets the accumulated validator rewards
func (k Keeper) SetValidatorRewardPool(ctx sdk.Context, amount int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(ValidatorRewardPoolKey, uint64ToBytes(uint64(amount)))
}

// AddToValidatorRewardPool adds to the accumulated validator rewards
func (k Keeper) AddToValidatorRewardPool(ctx sdk.Context, amount int64) {
	current := k.GetValidatorRewardPool(ctx)
	k.SetValidatorRewardPool(ctx, current+amount)
}

// bytesToUint64 converts bytes to uint64
func bytesToUint64(bz []byte) uint64 {
	if len(bz) != 8 {
		return 0
	}
	return uint64(bz[0])<<56 | uint64(bz[1])<<48 | uint64(bz[2])<<40 | uint64(bz[3])<<32 |
		uint64(bz[4])<<24 | uint64(bz[5])<<16 | uint64(bz[6])<<8 | uint64(bz[7])
}

// DistributeValidatorRewards is an exported wrapper for distributeValidatorRewardsInternal
// Used for testing
func (k Keeper) DistributeValidatorRewards(ctx sdk.Context, rewardPool int64) (int64, error) {
	return k.distributeValidatorRewardsInternal(ctx, rewardPool)
}
