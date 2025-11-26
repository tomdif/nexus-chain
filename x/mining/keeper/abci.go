package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"nexus/x/mining/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	// 1. Process emissions (accumulate tokens every minute into escrow)
	if err := k.ProcessEmissions(ctx); err != nil {
		k.Logger(ctx).Error("Failed to process emissions", "error", err)
	}

	// 2. Check and generate background job if needed
	// Priority: random public job from queue, then synthetic generation
	k.CheckAndGenerateBackgroundJob(ctx)

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

	// Get emission escrow for logging
	emissionEscrow := k.GetEmissionEscrow(ctx)

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
		"emission_escrow", emissionEscrow,
		"problem_size", k.GetCurrentProblemSize(ctx),
	)

	// Distribute validator rewards if there's anything to distribute
	if rewardPool > 0 {
		distributed, err := k.distributeValidatorRewardsInternal(ctx, rewardPool)
		if err != nil {
			k.Logger(ctx).Error("Failed to distribute validator rewards", "error", err)
		} else {
			k.Logger(ctx).Info("Distributed validator rewards",
				"total_pool", rewardPool,
				"distributed", distributed,
			)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"checkpoint_created",
			sdk.NewAttribute("checkpoint_id", fmt.Sprintf("%d", newID)),
			sdk.NewAttribute("height", fmt.Sprintf("%d", height)),
			sdk.NewAttribute("validator_rewards_distributed", fmt.Sprintf("%d", rewardPool)),
			sdk.NewAttribute("emission_escrow", fmt.Sprintf("%d", emissionEscrow)),
		),
	)
}

// distributeValidatorRewardsInternal distributes the reward pool to bonded validators
// proportional to their stake. Returns total amount distributed.
func (k Keeper) distributeValidatorRewardsInternal(ctx sdk.Context, rewardPool int64) (int64, error) {
	if k.stakingKeeper == nil || k.bankKeeper == nil {
		k.Logger(ctx).Info("Staking/bank keeper not available, skipping validator distribution")
		k.SetValidatorRewardPool(ctx, 0)
		return 0, nil
	}

	totalBonded, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get total bonded tokens: %w", err)
	}

	if totalBonded.IsZero() {
		k.Logger(ctx).Info("No bonded tokens, clearing validator reward pool")
		k.SetValidatorRewardPool(ctx, 0)
		return 0, nil
	}

	var totalDistributed int64
	rewardPoolInt := math.NewInt(rewardPool)

	// Iterate through all bonded validators and distribute proportionally
	err = k.stakingKeeper.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		valTokens := validator.GetBondedTokens()
		if valTokens.IsZero() {
			return false
		}

		share := valTokens.Mul(rewardPoolInt).Quo(totalBonded)
		if share.IsZero() {
			return false
		}

		shareAmount := share.Int64()
		valAddrStr := validator.GetOperator()

		valAddr, addrErr := sdk.ValAddressFromBech32(valAddrStr)
		if addrErr != nil {
			k.Logger(ctx).Error("Invalid validator address", "address", valAddrStr, "error", addrErr)
			return false
		}
		accAddr := sdk.AccAddress(valAddr)

		rewardCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", shareAmount))
		sendErr := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddr, rewardCoins)
		if sendErr != nil {
			k.Logger(ctx).Error("Failed to send validator reward",
				"validator", valAddrStr,
				"amount", shareAmount,
				"error", sendErr,
			)
			return false
		}

		totalDistributed += shareAmount

		k.Logger(ctx).Debug("Distributed validator reward",
			"validator", valAddrStr,
			"tokens", valTokens.String(),
			"share", shareAmount,
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"validator_reward_paid",
				sdk.NewAttribute("validator", valAddrStr),
				sdk.NewAttribute("amount", fmt.Sprintf("%d", shareAmount)),
			),
		)

		return false
	})

	if err != nil {
		return totalDistributed, fmt.Errorf("error iterating validators: %w", err)
	}

	k.SetValidatorRewardPool(ctx, 0)

	remainder := rewardPool - totalDistributed
	if remainder > 0 {
		k.Logger(ctx).Debug("Reward distribution remainder", "remainder", remainder)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_rewards_distributed",
			sdk.NewAttribute("total_pool", fmt.Sprintf("%d", rewardPool)),
			sdk.NewAttribute("total_distributed", fmt.Sprintf("%d", totalDistributed)),
			sdk.NewAttribute("remainder", fmt.Sprintf("%d", remainder)),
		),
	)

	return totalDistributed, nil
}
