package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

// Emission schedule constants
const (
	// BaseEmissionPerMinute is the starting emission rate (Epoch 1)
	// 35,950 NEX per minute = 35,950,000,000 unex per minute
	BaseEmissionPerMinute = 35_950_000_000 // in unex (smallest unit)

	// EpochDurationMinutes is 2 years in minutes
	// 2 years * 365.25 days * 24 hours * 60 minutes
	EpochDurationMinutes = 2 * 365.25 * 24 * 60 // ~1,051,200 minutes

	// MinEmissionPercent is the perpetual floor (1.5% of base = epoch 8+)
	MinEmissionPercent = 15 // 1.5% represented as 15/1000

	// BlocksPerMinute with 2-second block time
	BlocksPerMinute = 30
)

// EmissionEpoch defines the emission rate for each epoch
type EmissionEpoch struct {
	StartMinute int64 // Minutes since genesis
	EndMinute   int64 // Minutes until end of epoch (-1 for perpetual)
	RatePercent int64 // Percent of base rate (1000 = 100%)
}

// GetEmissionEpochs returns the emission schedule
func GetEmissionEpochs() []EmissionEpoch {
	epochMins := int64(EpochDurationMinutes)
	return []EmissionEpoch{
		{StartMinute: 0, EndMinute: epochMins, RatePercent: 1000},                     // Epoch 1: 100%
		{StartMinute: epochMins, EndMinute: epochMins * 2, RatePercent: 500},          // Epoch 2: 50%
		{StartMinute: epochMins * 2, EndMinute: epochMins * 3, RatePercent: 250},      // Epoch 3: 25%
		{StartMinute: epochMins * 3, EndMinute: epochMins * 4, RatePercent: 125},      // Epoch 4: 12.5%
		{StartMinute: epochMins * 4, EndMinute: epochMins * 5, RatePercent: 62},       // Epoch 5: 6.25%
		{StartMinute: epochMins * 5, EndMinute: epochMins * 6, RatePercent: 31},       // Epoch 6: 3.125%
		{StartMinute: epochMins * 6, EndMinute: epochMins * 7, RatePercent: 15},       // Epoch 7: 1.5625%
		{StartMinute: epochMins * 7, EndMinute: -1, RatePercent: MinEmissionPercent},  // Epoch 8+: 1.5% perpetual
	}
}

// Storage keys
var (
	EmissionEscrowKey       = []byte("emission_escrow")
	LastEmissionMinuteKey   = []byte("last_emission_minute")
	GenesisMinuteKey        = []byte("genesis_minute")
	CurrentJobStartMinuteKey = []byte("current_job_start_minute")
)

// GetEmissionEscrow returns the accumulated emission escrow
func (k Keeper) GetEmissionEscrow(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(EmissionEscrowKey)
	if bz == nil {
		return 0
	}
	return int64(bytesToUint64(bz))
}

// SetEmissionEscrow sets the emission escrow amount
func (k Keeper) SetEmissionEscrow(ctx sdk.Context, amount int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(EmissionEscrowKey, uint64ToBytes(uint64(amount)))
}

// GetLastEmissionMinute returns the last minute at which emissions were added
func (k Keeper) GetLastEmissionMinute(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(LastEmissionMinuteKey)
	if bz == nil {
		return 0
	}
	return int64(bytesToUint64(bz))
}

// SetLastEmissionMinute sets the last emission minute
func (k Keeper) SetLastEmissionMinute(ctx sdk.Context, minute int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(LastEmissionMinuteKey, uint64ToBytes(uint64(minute)))
}

// GetGenesisMinute returns the genesis minute (for epoch calculation)
func (k Keeper) GetGenesisMinute(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GenesisMinuteKey)
	if bz == nil {
		// Set genesis minute on first access
		genesisMinute := ctx.BlockTime().Unix() / 60
		k.SetGenesisMinute(ctx, genesisMinute)
		return genesisMinute
	}
	return int64(bytesToUint64(bz))
}

// SetGenesisMinute sets the genesis minute
func (k Keeper) SetGenesisMinute(ctx sdk.Context, minute int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(GenesisMinuteKey, uint64ToBytes(uint64(minute)))
}

// GetCurrentJobStartMinute returns when the current job started (for emission calc)
func (k Keeper) GetCurrentJobStartMinute(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(CurrentJobStartMinuteKey)
	if bz == nil {
		return 0
	}
	return int64(bytesToUint64(bz))
}

// SetCurrentJobStartMinute sets when the current job started
func (k Keeper) SetCurrentJobStartMinute(ctx sdk.Context, minute int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(CurrentJobStartMinuteKey, uint64ToBytes(uint64(minute)))
}

// GetCurrentEmissionRate returns the emission rate for the current epoch
func (k Keeper) GetCurrentEmissionRate(ctx sdk.Context) int64 {
	genesisMinute := k.GetGenesisMinute(ctx)
	currentMinute := ctx.BlockTime().Unix() / 60
	minutesSinceGenesis := currentMinute - genesisMinute

	epochs := GetEmissionEpochs()
	for _, epoch := range epochs {
		if epoch.EndMinute == -1 || minutesSinceGenesis < epoch.EndMinute {
			// Found the current epoch
			return (BaseEmissionPerMinute * epoch.RatePercent) / 1000
		}
	}

	// Fallback to perpetual rate
	return (BaseEmissionPerMinute * MinEmissionPercent) / 1000
}

// GetCurrentEpoch returns the current epoch number (1-8+)
func (k Keeper) GetCurrentEpoch(ctx sdk.Context) int {
	genesisMinute := k.GetGenesisMinute(ctx)
	currentMinute := ctx.BlockTime().Unix() / 60
	minutesSinceGenesis := currentMinute - genesisMinute

	epochs := GetEmissionEpochs()
	for i, epoch := range epochs {
		if epoch.EndMinute == -1 || minutesSinceGenesis < epoch.EndMinute {
			return i + 1
		}
	}
	return 8 // Perpetual epoch
}

// ProcessEmissions is called every block to accumulate emissions into escrow
func (k Keeper) ProcessEmissions(ctx sdk.Context) error {
	currentMinute := ctx.BlockTime().Unix() / 60
	lastMinute := k.GetLastEmissionMinute(ctx)

	// Initialize on first call
	if lastMinute == 0 {
		k.SetLastEmissionMinute(ctx, currentMinute)
		k.GetGenesisMinute(ctx) // Initialize genesis minute
		return nil
	}

	// Check if a new minute has passed
	if currentMinute <= lastMinute {
		return nil
	}

	// Calculate minutes elapsed (handle multiple minutes if blocks were slow)
	minutesElapsed := currentMinute - lastMinute

	// Get current emission rate
	emissionRate := k.GetCurrentEmissionRate(ctx)

	// Calculate total emissions to add
	emissionsToAdd := emissionRate * minutesElapsed

	// Mint new tokens to module account
	if k.bankKeeper != nil && emissionsToAdd > 0 {
		emissionCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", emissionsToAdd))
		err := k.bankKeeper.MintCoins(ctx, types.ModuleName, emissionCoins)
		if err != nil {
			k.Logger(ctx).Error("Failed to mint emission coins", "error", err)
			return err
		}

		// Add to escrow
		currentEscrow := k.GetEmissionEscrow(ctx)
		k.SetEmissionEscrow(ctx, currentEscrow+emissionsToAdd)

		k.Logger(ctx).Debug("Emissions accumulated",
			"minutes_elapsed", minutesElapsed,
			"rate_per_minute", emissionRate,
			"emissions_added", emissionsToAdd,
			"total_escrow", k.GetEmissionEscrow(ctx),
			"epoch", k.GetCurrentEpoch(ctx),
		)
	}

	k.SetLastEmissionMinute(ctx, currentMinute)

	return nil
}

// CalculateEmissionReward calculates the emission reward for solving a job
// Based on minutes elapsed since job was posted
func (k Keeper) CalculateEmissionReward(ctx sdk.Context, job types.Job) int64 {
	currentMinute := ctx.BlockTime().Unix() / 60
	jobStartMinute := job.CreatedAt / 60

	minutesElapsed := currentMinute - jobStartMinute
	if minutesElapsed < 1 {
		minutesElapsed = 1 // Minimum 1 minute
	}

	// Get the average emission rate over the job duration
	// For simplicity, use current rate (jobs typically don't span epochs)
	emissionRate := k.GetCurrentEmissionRate(ctx)

	return emissionRate * minutesElapsed
}

// ClaimEmissionReward releases emission escrow to the miner for solving a job
func (k Keeper) ClaimEmissionReward(ctx sdk.Context, job types.Job) int64 {
	emissionReward := k.CalculateEmissionReward(ctx, job)
	currentEscrow := k.GetEmissionEscrow(ctx)

	// Cap at available escrow
	if emissionReward > currentEscrow {
		emissionReward = currentEscrow
	}

	// Deduct from escrow (tokens already in module from minting)
	k.SetEmissionEscrow(ctx, currentEscrow-emissionReward)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"emission_reward_claimed",
			sdk.NewAttribute("job_id", job.Id),
			sdk.NewAttribute("minutes_to_solve", fmt.Sprintf("%d", (ctx.BlockTime().Unix()-job.CreatedAt)/60)),
			sdk.NewAttribute("emission_reward", fmt.Sprintf("%d", emissionReward)),
			sdk.NewAttribute("remaining_escrow", fmt.Sprintf("%d", k.GetEmissionEscrow(ctx))),
		),
	)

	return emissionReward
}
