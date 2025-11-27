package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/keeper"
	"nexus/x/mining/types"
)

func TestDockingEpochRewards(t *testing.T) {
	t.Log("=== DOCKING EPOCH REWARD DISTRIBUTION ===")

	k, ctx := setupKeeper(t)

	// Create test miners
	miner1 := sdk.AccAddress([]byte("miner1_address_____"))
	miner2 := sdk.AccAddress([]byte("miner2_address_____"))
	miner3 := sdk.AccAddress([]byte("miner3_address_____"))

	t.Log("\n--- Simulating ligand submissions ---")

	// Miner1: 10 ligands with 5 bonds each (multiplier 1.0)
	// 10 × 1.0 × 1000 = 10,000 shares
	for i := 0; i < 10; i++ {
		k.AddMinerEpochShares(ctx, miner1, 5)
	}
	t.Logf("Miner1: 10 ligands × 5 bonds = %d shares", k.GetMinerEpochShares(ctx, miner1))

	// Miner2: 5 ligands with 10 bonds each (multiplier 2.0)
	// 5 × 2.0 × 1000 = 10,000 shares
	for i := 0; i < 5; i++ {
		k.AddMinerEpochShares(ctx, miner2, 10)
	}
	t.Logf("Miner2: 5 ligands × 10 bonds = %d shares", k.GetMinerEpochShares(ctx, miner2))

	// Miner3: 20 ligands with 0 bonds each (multiplier 0.5)
	// 20 × 0.5 × 1000 = 10,000 shares
	for i := 0; i < 20; i++ {
		k.AddMinerEpochShares(ctx, miner3, 0)
	}
	t.Logf("Miner3: 20 ligands × 0 bonds = %d shares", k.GetMinerEpochShares(ctx, miner3))

	totalShares := k.GetDockingTotalShares(ctx)
	t.Logf("\nTotal shares: %d", totalShares)

	// Calculate expected percentages
	m1Share := float64(k.GetMinerEpochShares(ctx, miner1)) * 100 / float64(totalShares)
	m2Share := float64(k.GetMinerEpochShares(ctx, miner2)) * 100 / float64(totalShares)
	m3Share := float64(k.GetMinerEpochShares(ctx, miner3)) * 100 / float64(totalShares)

	t.Logf("\n--- Expected reward distribution ---")
	t.Logf("Miner1: %.2f%% of epoch emission", m1Share)
	t.Logf("Miner2: %.2f%% of epoch emission", m2Share)
	t.Logf("Miner3: %.2f%% of epoch emission", m3Share)

	// Each should have ~33.33%
	if m1Share < 33.0 || m1Share > 34.0 {
		t.Errorf("Miner1 should have ~33%%, got %.2f%%", m1Share)
	}
	if m2Share < 33.0 || m2Share > 34.0 {
		t.Errorf("Miner2 should have ~33%%, got %.2f%%", m2Share)
	}
	if m3Share < 33.0 || m3Share > 34.0 {
		t.Errorf("Miner3 should have ~33%%, got %.2f%%", m3Share)
	}

	// Simulate epoch emission calculation
	emissionRate := k.GetCurrentEmissionRate(ctx)
	epochEmission := emissionRate * keeper.DockingEpochMinutes

	t.Logf("\n--- Epoch emission calculation ---")
	t.Logf("Emission rate: %d unex/minute", emissionRate)
	t.Logf("Epoch duration: %d minutes", keeper.DockingEpochMinutes)
	t.Logf("Total epoch emission: %d unex (%.2f NEX)", epochEmission, float64(epochEmission)/1e9)

	// Calculate individual rewards
	m1Reward := (k.GetMinerEpochShares(ctx, miner1) * epochEmission) / totalShares
	m2Reward := (k.GetMinerEpochShares(ctx, miner2) * epochEmission) / totalShares
	m3Reward := (k.GetMinerEpochShares(ctx, miner3) * epochEmission) / totalShares

	t.Logf("\n--- Individual rewards ---")
	t.Logf("Miner1 reward: %d unex (%.4f NEX)", m1Reward, float64(m1Reward)/1e9)
	t.Logf("Miner2 reward: %d unex (%.4f NEX)", m2Reward, float64(m2Reward)/1e9)
	t.Logf("Miner3 reward: %d unex (%.4f NEX)", m3Reward, float64(m3Reward)/1e9)

	t.Log("\n✓ Equal work = equal reward, regardless of strategy!")
	t.Log("  (10 easy ligands = 5 medium ligands = 20 trivial ligands)")
}

func TestDockingEpochAdvance(t *testing.T) {
	t.Log("=== DOCKING EPOCH ADVANCEMENT ===")

	k, ctx := setupKeeper(t)

	miner := sdk.AccAddress([]byte("test_miner_address_"))

	// Initialize the epoch start time FIRST
	_ = k.GetDockingEpochStart(ctx) // This sets it to current time

	// Add some shares
	k.AddMinerEpochShares(ctx, miner, 5)
	k.AddMinerEpochShares(ctx, miner, 10)

	initialEpoch := k.GetDockingEpochNumber(ctx)
	initialShares := k.GetMinerEpochShares(ctx, miner)
	initialTotal := k.GetDockingTotalShares(ctx)

	t.Logf("Epoch %d: Miner has %d shares, total %d", initialEpoch, initialShares, initialTotal)

	// Advance time by 11 minutes (past the 10-minute epoch)
	newTime := ctx.BlockTime().Add(11 * time.Minute)
	newCtx := ctx.WithBlockTime(newTime)

	t.Logf("Time advanced from minute %d to minute %d", 
		ctx.BlockTime().Unix()/60, newCtx.BlockTime().Unix()/60)

	// Process epoch (would normally be called in EndBlocker)
	err := k.ProcessDockingEpoch(newCtx)
	if err != nil {
		t.Fatalf("ProcessDockingEpoch failed: %v", err)
	}

	newEpoch := k.GetDockingEpochNumber(newCtx)
	newShares := k.GetMinerEpochShares(newCtx, miner)
	newTotal := k.GetDockingTotalShares(newCtx)

	t.Logf("Epoch %d: Miner has %d shares, total %d", newEpoch, newShares, newTotal)

	if newEpoch != initialEpoch+1 {
		t.Errorf("Epoch should advance from %d to %d, got %d", initialEpoch, initialEpoch+1, newEpoch)
	}

	if newShares != 0 {
		t.Errorf("Miner shares should reset to 0, got %d", newShares)
	}

	if newTotal != 0 {
		t.Errorf("Total shares should reset to 0, got %d", newTotal)
	}

	t.Log("✓ Epoch advances and shares reset correctly!")
}

func TestBondMultiplierWeighting(t *testing.T) {
	t.Log("=== BOND MULTIPLIER WEIGHTING ===")

	testCases := []struct {
		bonds      int
		multiplier float64
		desc       string
	}{
		{0, 0.5, "trivial"},
		{5, 1.0, "baseline"},
		{10, 2.0, "medium complexity"},
		{13, 3.0, "maximum complexity"},
	}

	t.Logf("\nBond Multipliers:")
	for _, tc := range testCases {
		mult := types.GetBondMultiplier(tc.bonds)
		if mult != tc.multiplier {
			t.Errorf("Expected multiplier %.1f for %d bonds, got %.1f", tc.multiplier, tc.bonds, mult)
		}
		t.Logf("  %2d bonds: %.1fx (%s)", tc.bonds, mult, tc.desc)
	}

	t.Log("\n--- Work equivalence examples ---")
	t.Log("To earn equal shares:")
	t.Log("  1 × 13-bond ligand (3.0x) = 6 × 0-bond ligands (0.5x)")
	t.Log("  1 × 10-bond ligand (2.0x) = 2 × 5-bond ligands (1.0x)")

	t.Log("\n✓ Complex ligands reward more per submission!")
}
