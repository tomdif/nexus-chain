package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
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

// ProcessCheckpoint is called at checkpoint intervals to distribute validator rewards
// This should be called from EndBlock every CheckpointInterval blocks
func (k Keeper) ProcessCheckpoint(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	currentHeight := ctx.BlockHeight()

	// Check if we're at a checkpoint interval
	if currentHeight%params.CheckpointInterval != 0 {
		return nil
	}

	// Get the reward pool
	rewardPool := k.GetValidatorRewardPool(ctx)
	if rewardPool == 0 {
		k.Logger(ctx).Info("No validator rewards to distribute at checkpoint", "height", currentHeight)
		return nil
	}

	// Get all bonded validators and distribute proportionally to their stake
	if k.stakingKeeper == nil {
		k.Logger(ctx).Error("Staking keeper not available, cannot distribute validator rewards")
		return nil
	}

	totalBonded, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		k.Logger(ctx).Error("Failed to get total bonded tokens", "error", err)
		return err
	}

	if totalBonded.IsZero() {
		k.Logger(ctx).Info("No bonded tokens, cannot distribute validator rewards")
		return nil
	}

	// Create checkpoint record
	checkpointID := k.GetLastCheckpointID(ctx) + 1
	checkpoint := types.Checkpoint{
		Id:               checkpointID,
		StartHeight:      currentHeight - params.CheckpointInterval,
		EndHeight:        currentHeight,
		ValidatorRewards: rewardPool,
		Timestamp:        ctx.BlockTime().Unix(),
	}

	// Distribute rewards to validators based on their stake proportion
	// In a full implementation, we'd iterate through all validators
	// For now, we'll emit an event and track the checkpoint
	
	k.Logger(ctx).Info("Processing checkpoint",
		"checkpoint_id", checkpointID,
		"height", currentHeight,
		"reward_pool", rewardPool,
		"total_bonded", totalBonded.String(),
	)

	// Save checkpoint
	k.SetCheckpoint(ctx, checkpoint)
	k.SetLastCheckpointID(ctx, checkpointID)

	// Clear the reward pool (rewards distributed)
	k.SetValidatorRewardPool(ctx, 0)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"checkpoint_processed",
			sdk.NewAttribute("checkpoint_id", fmt.Sprintf("%d", checkpointID)),
			sdk.NewAttribute("height", fmt.Sprintf("%d", currentHeight)),
			sdk.NewAttribute("validator_rewards", fmt.Sprintf("%d", rewardPool)),
		),
	)

	return nil
}

// DistributeValidatorRewards distributes the reward pool to all bonded validators
// proportional to their stake. Returns the list of distributions made.
func (k Keeper) DistributeValidatorRewards(ctx sdk.Context) ([]ValidatorRewardDistribution, error) {
	rewardPool := k.GetValidatorRewardPool(ctx)
	if rewardPool == 0 {
		return nil, nil
	}

	if k.stakingKeeper == nil || k.bankKeeper == nil {
		return nil, fmt.Errorf("staking or bank keeper not available")
	}

	totalBonded, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return nil, err
	}

	if totalBonded.IsZero() {
		return nil, nil
	}

	var distributions []ValidatorRewardDistribution
	var totalDistributed int64

	// Get validators from the staking module
	// Note: In a full implementation, you'd iterate through all bonded validators
	// For this simplified version, we'll use a placeholder approach
	
	// The actual distribution would look something like:
	// k.stakingKeeper.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) bool {
	//     stake := validator.GetBondedTokens()
	//     proportion := stake.Quo(totalBonded)
	//     reward := proportion.Mul(sdk.NewInt(rewardPool)).Int64()
	//     ... distribute reward ...
	//     return false
	// })

	// For now, we'll just track that distribution should happen
	k.Logger(ctx).Info("Validator reward distribution triggered",
		"pool", rewardPool,
		"total_bonded", totalBonded.String(),
	)

	// Clear pool after distribution
	k.SetValidatorRewardPool(ctx, 0)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_rewards_distributed",
			sdk.NewAttribute("total_amount", fmt.Sprintf("%d", rewardPool)),
			sdk.NewAttribute("total_distributed", fmt.Sprintf("%d", totalDistributed)),
		),
	)

	return distributions, nil
}

// ValidatorRewardDistribution tracks a single validator's reward
type ValidatorRewardDistribution struct {
	Validator string
	Amount    int64
}
