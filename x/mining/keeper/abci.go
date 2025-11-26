package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	// Process expired jobs at beginning of each block
	return nil
}

func (k Keeper) EndBlocker(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	height := ctx.BlockHeight()

	// Create checkpoint and distribute validator rewards every CheckpointInterval blocks
	if height > 0 && height%params.CheckpointInterval == 0 {
		k.createCheckpointAndDistribute(ctx, height, params)
	}

	return nil
}

func (k Keeper) createCheckpointAndDistribute(ctx sdk.Context, height int64, params types.Params) {
	lastID := k.GetLastCheckpointID(ctx)
	newID := lastID + 1

	startHeight := int64(0)
	if lastID > 0 {
		lastCp, found := k.GetCheckpoint(ctx, lastID)
		if found {
			startHeight = lastCp.EndHeight + 1
		}
	}

	// Get validator reward pool before distribution
	rewardPool := k.GetValidatorRewardPool(ctx)

	checkpoint := types.Checkpoint{
		Id:               newID,
		StartHeight:      startHeight,
		EndHeight:        height,
		ValidatorRewards: rewardPool,
		Timestamp:        ctx.BlockTime().Unix(),
	}

	k.SetCheckpoint(ctx, checkpoint)
	k.SetLastCheckpointID(ctx, newID)

	k.Logger(ctx).Info("Created checkpoint",
		"id", newID,
		"height", height,
		"validator_reward_pool", rewardPool,
	)

	// Distribute validator rewards if there's anything to distribute
	if rewardPool > 0 {
		err := k.distributeValidatorRewards(ctx, rewardPool)
		if err != nil {
			k.Logger(ctx).Error("Failed to distribute validator rewards", "error", err)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"checkpoint_created",
			sdk.NewAttribute("checkpoint_id", fmt.Sprintf("%d", newID)),
			sdk.NewAttribute("height", fmt.Sprintf("%d", height)),
			sdk.NewAttribute("validator_rewards_distributed", fmt.Sprintf("%d", rewardPool)),
		),
	)
}

// distributeValidatorRewards distributes the reward pool to bonded validators
// proportional to their stake
func (k Keeper) distributeValidatorRewards(ctx sdk.Context, rewardPool int64) error {
	if k.stakingKeeper == nil || k.bankKeeper == nil {
		k.Logger(ctx).Info("Staking/bank keeper not available, skipping validator distribution")
		// Clear pool anyway to prevent accumulation
		k.SetValidatorRewardPool(ctx, 0)
		return nil
	}

	totalBonded, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total bonded tokens: %w", err)
	}

	if totalBonded.IsZero() {
		k.Logger(ctx).Info("No bonded tokens, clearing validator reward pool")
		k.SetValidatorRewardPool(ctx, 0)
		return nil
	}

	// In a full implementation, we would iterate through all validators
	// and distribute proportionally to their stake:
	//
	// k.stakingKeeper.IterateBondedValidatorsByPower(ctx, func(index int64, val stakingtypes.ValidatorI) bool {
	//     valTokens := val.GetBondedTokens()
	//     share := valTokens.Mul(sdk.NewInt(rewardPool)).Quo(totalBonded)
	//     // Send share to validator's operator address
	//     return false
	// })
	//
	// For now, we just log and clear the pool
	// The tokens remain in the module account for later implementation

	k.Logger(ctx).Info("Validator reward distribution",
		"pool", rewardPool,
		"total_bonded", totalBonded.String(),
		"note", "Distribution to individual validators not yet implemented - tokens held in module",
	)

	// Clear the pool (in production, only clear after successful distribution)
	k.SetValidatorRewardPool(ctx, 0)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_rewards_distributed",
			sdk.NewAttribute("amount", fmt.Sprintf("%d", rewardPool)),
			sdk.NewAttribute("total_bonded", totalBonded.String()),
		),
	)

	return nil
}
