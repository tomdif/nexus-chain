package keeper

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
	storetypes "cosmossdk.io/store/types"
)

// Docking epoch constants
const (
	DockingEpochMinutes = 10 // 10-minute epochs
)

// DockingEpochState tracks work shares for an epoch
type DockingEpochState struct {
	EpochNumber    uint64            // Current epoch number
	TotalShares    int64             // Total weighted shares this epoch
	MinerShares    map[string]int64  // Miner address -> their weighted shares
	StartMinute    int64             // When this epoch started
}

// Storage keys for docking epochs
var (
	DockingEpochKey        = []byte("docking_epoch")
	DockingEpochSharesKey  = []byte("docking_epoch_shares_")
	DockingTotalSharesKey  = []byte("docking_total_shares")
	DockingEpochNumberKey  = []byte("docking_epoch_number")
	DockingEpochStartKey   = []byte("docking_epoch_start")
)

// GetDockingEpochNumber returns current docking epoch
func (k Keeper) GetDockingEpochNumber(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(DockingEpochNumberKey)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) SetDockingEpochNumber(ctx sdk.Context, epoch uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(DockingEpochNumberKey, uint64ToBytes(epoch))
}

// GetDockingEpochStart returns when current epoch started (unix minute)
func (k Keeper) GetDockingEpochStart(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(DockingEpochStartKey)
	if bz == nil {
		// Initialize on first access
		minute := ctx.BlockTime().Unix() / 60
		k.SetDockingEpochStart(ctx, minute)
		return minute
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) SetDockingEpochStart(ctx sdk.Context, minute int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(DockingEpochStartKey, uint64ToBytes(uint64(minute)))
}

// GetDockingTotalShares returns total weighted shares for current epoch
func (k Keeper) GetDockingTotalShares(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(DockingTotalSharesKey)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) SetDockingTotalShares(ctx sdk.Context, shares int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(DockingTotalSharesKey, uint64ToBytes(uint64(shares)))
}

// GetMinerEpochShares returns a miner's weighted shares for current epoch
func (k Keeper) GetMinerEpochShares(ctx sdk.Context, miner sdk.AccAddress) int64 {
	store := ctx.KVStore(k.storeKey)
	key := append(DockingEpochSharesKey, miner.Bytes()...)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return int64(binary.BigEndian.Uint64(bz))
}

func (k Keeper) SetMinerEpochShares(ctx sdk.Context, miner sdk.AccAddress, shares int64) {
	store := ctx.KVStore(k.storeKey)
	key := append(DockingEpochSharesKey, miner.Bytes()...)
	store.Set(key, uint64ToBytes(uint64(shares)))
}

// AddMinerEpochShares adds weighted shares for a ligand submission
// weight = bond_multiplier * 1000 (to avoid float)
func (k Keeper) AddMinerEpochShares(ctx sdk.Context, miner sdk.AccAddress, rotatableBonds int) int64 {
	// Calculate weighted share (multiply by 1000 to keep precision)
	multiplier := types.GetBondMultiplier(rotatableBonds)
	weightedShare := int64(multiplier * 1000)

	// Add to miner's shares
	currentShares := k.GetMinerEpochShares(ctx, miner)
	k.SetMinerEpochShares(ctx, miner, currentShares+weightedShare)

	// Add to total shares
	totalShares := k.GetDockingTotalShares(ctx)
	k.SetDockingTotalShares(ctx, totalShares+weightedShare)

	return weightedShare
}

// ProcessDockingEpoch checks if epoch should end and distributes rewards
// Called in EndBlocker
func (k Keeper) ProcessDockingEpoch(ctx sdk.Context) error {
	currentMinute := ctx.BlockTime().Unix() / 60
	epochStart := k.GetDockingEpochStart(ctx)

	// Check if 10 minutes have passed
	if currentMinute-epochStart < DockingEpochMinutes {
		return nil
	}

	// Time to end epoch and distribute rewards
	return k.EndDockingEpochAndDistribute(ctx)
}

// EndDockingEpochAndDistribute distributes epoch emission to miners
func (k Keeper) EndDockingEpochAndDistribute(ctx sdk.Context) error {
	epochNumber := k.GetDockingEpochNumber(ctx)
	totalShares := k.GetDockingTotalShares(ctx)

	if totalShares == 0 {
		// No work done this epoch, just advance
		k.advanceDockingEpoch(ctx)
		return nil
	}

	// Calculate emission for this epoch (10 minutes worth)
	emissionRate := k.GetCurrentEmissionRate(ctx) // unex per minute
	epochEmission := emissionRate * DockingEpochMinutes

	// Distribute proportionally to all miners who contributed
	k.distributeDockingRewards(ctx, epochEmission, totalShares)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"docking_epoch_end",
			sdk.NewAttribute("epoch", fmt.Sprintf("%d", epochNumber)),
			sdk.NewAttribute("total_shares", fmt.Sprintf("%d", totalShares)),
			sdk.NewAttribute("emission_distributed", fmt.Sprintf("%d", epochEmission)),
		),
	)

	k.Logger(ctx).Info("Docking epoch ended",
		"epoch", epochNumber,
		"total_shares", totalShares,
		"emission", epochEmission,
	)

	// Advance to next epoch
	k.advanceDockingEpoch(ctx)

	return nil
}

// distributeDockingRewards distributes emission proportionally
func (k Keeper) distributeDockingRewards(ctx sdk.Context, totalEmission int64, totalShares int64) {
	store := ctx.KVStore(k.storeKey)
	
	// Iterate through all miners with shares
	iterator := storetypes.KVStorePrefixIterator(store, DockingEpochSharesKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		minerShares := int64(binary.BigEndian.Uint64(iterator.Value()))
		if minerShares == 0 {
			continue
		}

		// Calculate proportional reward: (minerShares / totalShares) * totalEmission
		reward := (minerShares * totalEmission) / totalShares

		if reward > 0 {
			// Extract miner address from key
			minerBytes := iterator.Key()[len(DockingEpochSharesKey):]
			minerAddr := sdk.AccAddress(minerBytes)

			// Transfer reward from module to miner
			if k.bankKeeper != nil {
				rewardCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", reward))
				err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, minerAddr, rewardCoins)
				if err != nil {
					k.Logger(ctx).Error("Failed to send docking reward", 
						"miner", minerAddr.String(), 
						"reward", reward,
						"error", err,
					)
					continue
				}
			}

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"docking_reward_paid",
					sdk.NewAttribute("miner", minerAddr.String()),
					sdk.NewAttribute("shares", fmt.Sprintf("%d", minerShares)),
					sdk.NewAttribute("reward", fmt.Sprintf("%d", reward)),
					sdk.NewAttribute("share_percent", fmt.Sprintf("%.2f%%", float64(minerShares)*100/float64(totalShares))),
				),
			)
		}
	}
}

// advanceDockingEpoch resets for new epoch
func (k Keeper) advanceDockingEpoch(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// Clear all miner shares
	iterator := storetypes.KVStorePrefixIterator(store, DockingEpochSharesKey)
	defer iterator.Close()
	
	keysToDelete := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		keysToDelete = append(keysToDelete, iterator.Key())
	}
	for _, key := range keysToDelete {
		store.Delete(key)
	}

	// Reset total shares
	k.SetDockingTotalShares(ctx, 0)

	// Increment epoch number
	k.SetDockingEpochNumber(ctx, k.GetDockingEpochNumber(ctx)+1)

	// Set new epoch start
	k.SetDockingEpochStart(ctx, ctx.BlockTime().Unix()/60)
}
