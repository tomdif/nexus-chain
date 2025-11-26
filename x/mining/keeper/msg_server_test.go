package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"nexus/x/mining/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"nexus/x/mining/types"
)

func init() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("nexus", "nexuspub")
	config.SetBech32PrefixForValidator("nexusvaloper", "nexusvaloperpub")
	config.SetBech32PrefixForConsensusNode("nexusvalcons", "nexusvalconspub")
	config.Seal()
}

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	return setupKeeperWithBank(t, nil)
}

func setupKeeperWithBank(t *testing.T, bankKeeper types.BankKeeper) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey("mem_mining")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1}, false, log.NewNopLogger())
	k := keeper.NewKeeper(cdc, storeKey, memKey, nil, bankKeeper, "authority")
	k.SetParams(ctx, types.DefaultParams())
	return k, ctx
}

const (
	testCustomer = "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg"
	testMiner    = "nexus109vzpgxnm8mjye50yaf4vj8yp59he3aclran4g"
)

func TestPostJob(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)
	resp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: -100, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	if err != nil { t.Fatalf("PostJob failed: %v", err) }
	t.Logf("Created job: %s", resp.JobId)
	job, found := k.GetJob(ctx, resp.JobId)
	if !found { t.Fatal("Job not found") }
	t.Logf("Job verified: ID=%s", job.Id)
}

func TestSubmitProof(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: -100, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	resp, err := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{
		Miner: testCustomer, JobId: postResp.JobId, Energy: -150, Proof: []byte{0xde, 0xad, 0xbe, 0xef},
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	})
	if err != nil { t.Logf("SubmitProof error: %v", err) } else { t.Logf("Shares: %d", resp.SharesEarned) }
	job, _ := k.GetJob(ctx, postResp.JobId)
	if job.TotalShares > 0 { t.Logf("SUCCESS: Universal Share Formula working") }
}

func TestUniversalShareFormula(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	proof := []byte{0xde, 0xad, 0xbe, 0xef}
	resp1, _ := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{Miner: testCustomer, JobId: postResp.JobId, Energy: -500, Proof: proof, SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002"})
	resp2, _ := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{Miner: testCustomer, JobId: postResp.JobId, Energy: -700, Proof: proof, SolutionHash: "0000000000000000000000000000000000000000000000000000000000000003"})
	resp3, _ := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{Miner: testCustomer, JobId: postResp.JobId, Energy: -650, Proof: proof, SolutionHash: "0000000000000000000000000000000000000000000000000000000000000004"})
	job, _ := k.GetJob(ctx, postResp.JobId)
	if resp1.SharesEarned != 500 || resp2.SharesEarned != 200 || resp3.SharesEarned != 0 || job.TotalShares != 700 {
		t.Errorf("Share formula failed")
	} else { t.Logf("Universal Share Formula verified: 500+200+0=700") }
}

func TestPostJobWithFeeBurn(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)
	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)))
	resp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	job, _ := k.GetJob(ctx, resp.JobId)
	if job.Reward != 980000 { t.Errorf("Job reward: expected 980000, got %d", job.Reward) }
	if bankKeeper.Balances[customerAddr.String()].AmountOf("unexus").Int64() != 9000000 { t.Error("Customer should have 9M") }
	if bankKeeper.ModuleBalances[types.ModuleName].AmountOf("unexus").Int64() != 980000 { t.Error("Module should have 980K") }
	t.Logf("2%% fee burn: Customer paid 1M, module has 980K, 20K burned")
}

func TestClaimRewardsValidatorShareHeld(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)
	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	minerAddr, _ := sdk.AccAddressFromBech32(testMiner)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)))
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{
		Miner: testMiner, JobId: postResp.JobId, Energy: -500, Proof: []byte{0xde, 0xad, 0xbe, 0xef},
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	})
	msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), &types.MsgClaimRewards{Claimer: testMiner, JobId: postResp.JobId})
	minerBal := bankKeeper.Balances[minerAddr.String()].AmountOf("unexus").Int64()
	moduleBal := bankKeeper.ModuleBalances[types.ModuleName].AmountOf("unexus").Int64()
	if minerBal != 784000 { t.Errorf("Miner: expected 784000, got %d", minerBal) } else { t.Logf("Miner got 80%%: %d", minerBal) }
	if moduleBal != 196000 { t.Errorf("Module: expected 196000, got %d", moduleBal) } else { t.Logf("Validator share held: %d", moduleBal) }
}

func TestCancelJobRefund(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)
	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)))
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	msgServer.CancelJob(sdk.WrapSDKContext(ctx), &types.MsgCancelJob{Customer: testCustomer, JobId: postResp.JobId})
	customerBal := bankKeeper.Balances[customerAddr.String()].AmountOf("unexus").Int64()
	moduleBal := bankKeeper.ModuleBalances[types.ModuleName].AmountOf("unexus").Int64()
	if customerBal != 9980000 { t.Errorf("Customer: expected 9980000, got %d", customerBal) } else { t.Logf("Customer refunded net: %d", customerBal) }
	if moduleBal != 0 { t.Errorf("Module: expected 0, got %d", moduleBal) }
}

func TestFullTokenFlow(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)
	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	minerAddr, _ := sdk.AccAddressFromBech32(testMiner)
	initialSupply := int64(10000000)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", initialSupply)))
	t.Logf("Initial: %d", initialSupply)
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{
		Miner: testMiner, JobId: postResp.JobId, Energy: -500, Proof: []byte{0xde, 0xad, 0xbe, 0xef},
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	})
	msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), &types.MsgClaimRewards{Claimer: testMiner, JobId: postResp.JobId})
	customerFinal := bankKeeper.Balances[customerAddr.String()].AmountOf("unexus").Int64()
	minerFinal := bankKeeper.Balances[minerAddr.String()].AmountOf("unexus").Int64()
	moduleFinal := bankKeeper.ModuleBalances[types.ModuleName].AmountOf("unexus").Int64()
	totalCirculating := customerFinal + minerFinal + moduleFinal
	t.Logf("Final: Customer=%d, Miner=%d, Module=%d, Total=%d, Burned=%d", customerFinal, minerFinal, moduleFinal, totalCirculating, initialSupply-totalCirculating)
	if initialSupply-totalCirculating != 20000 { t.Errorf("Burn: expected 20000, got %d", initialSupply-totalCirculating) }
	if moduleFinal != 196000 { t.Errorf("Validator share: expected 196000, got %d", moduleFinal) }
	if minerFinal != 784000 { t.Errorf("Miner: expected 784000, got %d", minerFinal) }
	t.Logf("SUMMARY: 2%% fee burned=20K, 80%% miner=784K, 20%% validators=196K held")
}

func TestInsufficientFunds(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)
	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 100)))
	_, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	if err == nil { t.Error("Expected error") } else { t.Logf("Correctly rejected: %v", err) }
}

func TestValidatorRewardPoolAccumulation(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 100000000)))

	t.Logf("=== VALIDATOR REWARD POOL ACCUMULATION ===")

	// Initial pool should be 0
	pool := k.GetValidatorRewardPool(ctx)
	if pool != 0 {
		t.Errorf("Initial pool: expected 0, got %d", pool)
	}
	t.Logf("Initial pool: %d", pool)

	// Post and complete 3 jobs
	for i := 1; i <= 3; i++ {
		postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
			Customer: testCustomer, 
			ProblemHash: fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000000%d", i),
			Threshold: 1000, 
			Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), 
			Duration: 100,
		})

		msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{
			Miner: testMiner, JobId: postResp.JobId, Energy: -500, 
			Proof: []byte{0xde, 0xad, 0xbe, 0xef},
			SolutionHash: fmt.Sprintf("000000000000000000000000000000000000000000000000000000000000001%d", i),
		})

		msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), &types.MsgClaimRewards{
			Claimer: testMiner, JobId: postResp.JobId,
		})

		pool = k.GetValidatorRewardPool(ctx)
		t.Logf("After job %d: pool = %d", i, pool)
	}

	// Each job: 1M gross -> 980K net -> 196K validator share
	// 3 jobs = 3 * 196K = 588K
	expectedPool := int64(588000)
	if pool != expectedPool {
		t.Errorf("Final pool: expected %d, got %d", expectedPool, pool)
	} else {
		t.Logf("Validator pool accumulated correctly: %d (3 x 196K)", pool)
	}
}

func setupKeeperWithBothKeepers(t *testing.T, bankKeeper *MockBankKeeper, stakingKeeper *MockStakingKeeper) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey("mem_mining")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1}, false, log.NewNopLogger())
	k := keeper.NewKeeper(cdc, storeKey, memKey, stakingKeeper, bankKeeper, "authority")
	k.SetParams(ctx, types.DefaultParams())
	return k, ctx
}

func TestValidatorRewardDistribution(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	stakingKeeper := NewMockStakingKeeper()
	k, ctx := setupKeeperWithBothKeepers(t, bankKeeper, stakingKeeper)

	// Setup 3 validators with different stakes
	// Val1: 50% stake, Val2: 30% stake, Val3: 20% stake
	val1Bytes := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	val1Addr := sdk.ValAddress(val1Bytes).String()
	val2Bytes := []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	val2Addr := sdk.ValAddress(val2Bytes).String()
	val3Bytes := []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
	val3Addr := sdk.ValAddress(val3Bytes).String()
	
	stakingKeeper.AddValidator(val1Addr, 500000) // 50%
	stakingKeeper.AddValidator(val2Addr, 300000) // 30%
	stakingKeeper.AddValidator(val3Addr, 200000) // 20%

	t.Logf("=== VALIDATOR REWARD DISTRIBUTION ===")
	t.Logf("Val1: 500K stake (50%%)")
	t.Logf("Val2: 300K stake (30%%)")
	t.Logf("Val3: 200K stake (20%%)")

	// Set up reward pool (simulating accumulated validator shares)
	rewardPool := int64(1000000) // 1M to distribute
	k.SetValidatorRewardPool(ctx, rewardPool)
	
	// Fund the module account so it can distribute
	bankKeeper.SetModuleBalance(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("unexus", rewardPool)))

	t.Logf("Reward pool: %d", rewardPool)
	t.Logf("Module balance: %s", bankKeeper.ModuleBalances[types.ModuleName])

	// Trigger distribution
	distributed, err := k.DistributeValidatorRewards(ctx, rewardPool)
	if err != nil {
		t.Fatalf("Distribution failed: %v", err)
	}

	// Check pool is cleared
	poolAfter := k.GetValidatorRewardPool(ctx)
	if poolAfter != 0 {
		t.Errorf("Pool should be 0 after distribution, got %d", poolAfter)
	}

	// Check validator balances
	val1AccAddr, _ := sdk.ValAddressFromBech32(val1Addr)
	val2AccAddr, _ := sdk.ValAddressFromBech32(val2Addr)
	val3AccAddr, _ := sdk.ValAddressFromBech32(val3Addr)

	val1Bal := bankKeeper.Balances[sdk.AccAddress(val1AccAddr).String()].AmountOf("unexus").Int64()
	val2Bal := bankKeeper.Balances[sdk.AccAddress(val2AccAddr).String()].AmountOf("unexus").Int64()
	val3Bal := bankKeeper.Balances[sdk.AccAddress(val3AccAddr).String()].AmountOf("unexus").Int64()

	t.Logf("\nDistribution results:")
	t.Logf("  Val1 (50%%): %d", val1Bal)
	t.Logf("  Val2 (30%%): %d", val2Bal)
	t.Logf("  Val3 (20%%): %d", val3Bal)
	t.Logf("  Total distributed: %d", distributed)

	// Expected: 500K, 300K, 200K
	if val1Bal != 500000 {
		t.Errorf("Val1: expected 500000, got %d", val1Bal)
	}
	if val2Bal != 300000 {
		t.Errorf("Val2: expected 300000, got %d", val2Bal)
	}
	if val3Bal != 200000 {
		t.Errorf("Val3: expected 200000, got %d", val3Bal)
	}

	// Module should be empty
	moduleBal := bankKeeper.ModuleBalances[types.ModuleName].AmountOf("unexus").Int64()
	if moduleBal != 0 {
		t.Errorf("Module should be empty, got %d", moduleBal)
	}

	t.Logf("\n✓ Validators received rewards proportional to stake!")
}

func TestMockStakingKeeperDirect(t *testing.T) {
	stakingKeeper := NewMockStakingKeeper()
	stakingKeeper.AddValidator("nexusvaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a", 500000)
	stakingKeeper.AddValidator("nexusvaloper1qyqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqfhvms4", 300000)
	
	t.Logf("TotalBonded: %d", stakingKeeper.TotalBonded)
	t.Logf("Validators: %d", len(stakingKeeper.Validators))
	
	total, err := stakingKeeper.TotalBondedTokens(nil)
	if err != nil {
		t.Fatalf("TotalBondedTokens error: %v", err)
	}
	t.Logf("TotalBondedTokens returned: %s", total.String())
	
	count := 0
	err = stakingKeeper.IterateBondedValidatorsByPower(nil, func(index int64, val stakingtypes.ValidatorI) bool {
		t.Logf("Validator %d: addr=%s tokens=%s", index, val.GetOperator(), val.GetBondedTokens().String())
		count++
		return false
	})
	if err != nil {
		t.Fatalf("IterateBondedValidatorsByPower error: %v", err)
	}
	t.Logf("Iterated %d validators", count)
}

func TestDistributeWithDebug(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	stakingKeeper := NewMockStakingKeeper()
	
	// Add validators BEFORE creating keeper
	val1Bytes := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	val1Addr := sdk.ValAddress(val1Bytes).String()
	stakingKeeper.AddValidator(val1Addr, 1000000)
	
	t.Logf("StakingKeeper before setup: TotalBonded=%d, Validators=%d", stakingKeeper.TotalBonded, len(stakingKeeper.Validators))
	
	k, ctx := setupKeeperWithBothKeepers(t, bankKeeper, stakingKeeper)
	
	t.Logf("StakingKeeper after setup: TotalBonded=%d, Validators=%d", stakingKeeper.TotalBonded, len(stakingKeeper.Validators))
	
	// Set up reward pool
	rewardPool := int64(1000000)
	k.SetValidatorRewardPool(ctx, rewardPool)
	bankKeeper.SetModuleBalance(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("unexus", rewardPool)))
	
	t.Logf("Pool set: %d", k.GetValidatorRewardPool(ctx))
	t.Logf("Module balance: %s", bankKeeper.ModuleBalances[types.ModuleName])
	
	// Call distribution
	distributed, err := k.DistributeValidatorRewards(ctx, rewardPool)
	t.Logf("DistributeValidatorRewards returned: distributed=%d, err=%v", distributed, err)
	
	// Check validator balance
	val1AccAddr, _ := sdk.ValAddressFromBech32(val1Addr)
	val1Bal := bankKeeper.Balances[sdk.AccAddress(val1AccAddr).String()].AmountOf("unexus").Int64()
	t.Logf("Val1 balance: %d", val1Bal)
	
	poolAfter := k.GetValidatorRewardPool(ctx)
	t.Logf("Pool after: %d", poolAfter)
}

func TestDistributeWithMoreDebug(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	stakingKeeper := NewMockStakingKeeper()
	
	val1Bytes := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	val1Addr := sdk.ValAddress(val1Bytes).String()
	stakingKeeper.AddValidator(val1Addr, 1000000)
	
	// Create keeper manually to check
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey("mem_mining")
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	stateStore.LoadLatestVersion()
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1}, false, log.NewNopLogger())
	
	// Check that interface is satisfied
	var sk types.StakingKeeper = stakingKeeper
	var bk types.BankKeeper = bankKeeper
	t.Logf("StakingKeeper interface: %T, nil=%v", sk, sk == nil)
	t.Logf("BankKeeper interface: %T, nil=%v", bk, bk == nil)
	
	k := keeper.NewKeeper(cdc, storeKey, memKey, sk, bk, "authority")
	k.SetParams(ctx, types.DefaultParams())
	
	// Test TotalBondedTokens directly through keeper's staking keeper
	total, err := sk.TotalBondedTokens(ctx)
	t.Logf("Direct TotalBondedTokens: %s, err=%v", total.String(), err)
	
	// Set up and distribute
	rewardPool := int64(1000000)
	k.SetValidatorRewardPool(ctx, rewardPool)
	bankKeeper.SetModuleBalance(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("unexus", rewardPool)))
	
	distributed, err := k.DistributeValidatorRewards(ctx, rewardPool)
	t.Logf("Distributed: %d, err=%v", distributed, err)
	
	val1AccAddr, _ := sdk.ValAddressFromBech32(val1Addr)
	val1Bal := bankKeeper.Balances[sdk.AccAddress(val1AccAddr).String()].AmountOf("unexus").Int64()
	t.Logf("Val1 balance after: %d", val1Bal)
}

func TestValidatorAddressConversion(t *testing.T) {
	// Test the validator address conversion
	val1Bytes := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	val1Addr := sdk.ValAddress(val1Bytes).String()
	
	valAddr, err := sdk.ValAddressFromBech32(val1Addr)
	t.Logf("ValAddressFromBech32: addr=%v, err=%v", valAddr, err)
	
	if err == nil {
		accAddr := sdk.AccAddress(valAddr)
		t.Logf("AccAddress: %s", accAddr.String())
	}
	
	// Try creating a valid address
	// Generate from bytes
	valBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	valAddr2 := sdk.ValAddress(valBytes)
	t.Logf("Generated ValAddress: %s", valAddr2.String())
	
	// Now convert back
	valAddr3, err := sdk.ValAddressFromBech32(valAddr2.String())
	t.Logf("Converted back: addr=%v, err=%v", valAddr3, err)
}

func TestEmissionSchedule(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Logf("=== EMISSION SCHEDULE TEST ===")

	// Test epoch calculation
	epoch := k.GetCurrentEpoch(ctx)
	t.Logf("Current epoch: %d", epoch)

	// Test emission rate
	rate := k.GetCurrentEmissionRate(ctx)
	t.Logf("Current emission rate: %d unex/minute", rate)

	// Expected: Epoch 1 = 100% of base = 35,950,000,000 unex/minute
	expectedRate := int64(35_950_000_000)
	if rate != expectedRate {
		t.Errorf("Expected rate %d, got %d", expectedRate, rate)
	}

	// Test escrow accumulation
	initialEscrow := k.GetEmissionEscrow(ctx)
	t.Logf("Initial escrow: %d", initialEscrow)

	// Simulate processing emissions
	err := k.ProcessEmissions(ctx)
	if err != nil {
		t.Errorf("ProcessEmissions failed: %v", err)
	}

	t.Logf("Emission schedule verified for Epoch 1")
}

func TestEmissionEpochs(t *testing.T) {
	t.Logf("=== EMISSION EPOCHS ===")

	epochs := keeper.GetEmissionEpochs()
	baseRate := int64(35_950_000_000)

	for i, epoch := range epochs {
		rate := (baseRate * epoch.RatePercent) / 1000
		yearStart := (i * 2) + 1
		yearEnd := yearStart + 1
		if epoch.EndMinute == -1 {
			t.Logf("Epoch %d (Year %d+):   %.3f%% = %d unex/min (perpetual)",
				i+1, yearStart, float64(epoch.RatePercent)/10, rate)
		} else {
			t.Logf("Epoch %d (Year %d-%d): %.3f%% = %d unex/min",
				i+1, yearStart, yearEnd, float64(epoch.RatePercent)/10, rate)
		}
	}
}

func TestBackgroundJobGeneration(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Logf("=== BACKGROUND JOB GENERATION ===")

	// Initially no active jobs
	activeCount := k.GetActiveJobCount(ctx)
	t.Logf("Initial active job count: %d", activeCount)

	// Check and generate should create a synthetic job
	k.CheckAndGenerateBackgroundJob(ctx)

	// Should now have an active job
	activeCount = k.GetActiveJobCount(ctx)
	t.Logf("After generation, active count: %d", activeCount)

	if activeCount != 1 {
		t.Errorf("Expected 1 active job, got %d", activeCount)
	}

	// Get current job
	currentJobID := k.GetCurrentJobID(ctx)
	t.Logf("Current job ID: %s", currentJobID)

	if currentJobID == "" {
		t.Error("Expected current job ID to be set")
	}

	// Verify job exists and is background
	job, found := k.GetJob(ctx, currentJobID)
	if !found {
		t.Error("Job not found")
	}

	t.Logf("Job details: type=%s, threshold=%d, is_background=%t",
		job.ProblemType, job.Threshold, job.IsBackground)

	if !job.IsBackground {
		t.Error("Expected job to be background")
	}

	if job.ProblemType != "ising_synthetic" {
		t.Errorf("Expected ising_synthetic, got %s", job.ProblemType)
	}

	// Should not generate another job while one is active
	k.CheckAndGenerateBackgroundJob(ctx)
	activeCount = k.GetActiveJobCount(ctx)
	if activeCount != 1 {
		t.Errorf("Should still have 1 job, got %d", activeCount)
	}

	t.Logf("✓ Background job generation working!")
}

func TestPublicJobQueue(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Logf("=== PUBLIC JOB QUEUE (RANDOM SELECTION) ===")

	// Add jobs to queue
	k.AddToPublicJobQueue(ctx, "pub_job_1")
	k.AddToPublicJobQueue(ctx, "pub_job_2")
	k.AddToPublicJobQueue(ctx, "pub_job_3")

	queueLen := k.GetPublicJobQueueLength(ctx)
	t.Logf("Queue length: %d", queueLen)

	if queueLen != 3 {
		t.Errorf("Expected 3 jobs in queue, got %d", queueLen)
	}

	// Select random job
	selected := k.SelectRandomFromQueue(ctx)
	t.Logf("Randomly selected: %s", selected)

	if selected == "" {
		t.Error("Expected a job to be selected")
	}

	// Queue should now have 2
	queueLen = k.GetPublicJobQueueLength(ctx)
	if queueLen != 2 {
		t.Errorf("Expected 2 jobs after selection, got %d", queueLen)
	}

	// Remaining queue
	queue := k.GetPublicJobQueue(ctx)
	t.Logf("Remaining queue: %v", queue)

	// Selected job should not be in queue
	for _, id := range queue {
		if id == selected {
			t.Errorf("Selected job %s should not be in queue", selected)
		}
	}

	t.Logf("✓ Public job queue with random selection working!")
}

func TestDifficultyAdjustment(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Logf("=== DIFFICULTY ADJUSTMENT ===")

	initialSize := k.GetCurrentProblemSize(ctx)
	t.Logf("Initial problem size: %d", initialSize)

	// Simulate 3 fast solves (under 8 minutes = 480 seconds)
	k.AddSolveTime(ctx, 300) // 5 min
	k.AddSolveTime(ctx, 350) // ~6 min
	shouldAdjust := k.AddSolveTime(ctx, 400) // ~7 min

	t.Logf("Should adjust after 3 solves: %t", shouldAdjust)

	if !shouldAdjust {
		t.Error("Expected adjustment after 3 solves")
	}

	// Trigger adjustment
	k.AdjustDifficulty(ctx)

	newSize := k.GetCurrentProblemSize(ctx)
	t.Logf("New problem size: %d (was %d)", newSize, initialSize)

	// Fast solves should increase difficulty
	if newSize <= initialSize {
		t.Errorf("Expected size to increase, got %d <= %d", newSize, initialSize)
	}

	t.Logf("✓ Difficulty adjustment working!")
}
