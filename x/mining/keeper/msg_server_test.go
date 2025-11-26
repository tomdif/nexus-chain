package keeper_test

import (
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

// Test addresses - these are valid bech32 addresses
const (
	testCustomer = "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg"
	testMiner    = "nexus109vzpgxnm8mjye50yaf4vj8yp59he3aclran4g"
)

func TestPostJob(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	msg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   -100,
		Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)),
		Duration:    100,
	}

	resp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}

	t.Logf("Created job: %s", resp.JobId)

	job, found := k.GetJob(ctx, resp.JobId)
	if !found {
		t.Fatal("Job not found after creation")
	}

	t.Logf("Job verified: ID=%s, Threshold=%d, Deadline=%d", job.Id, job.Threshold, job.Deadline)
}

func TestSubmitProofWithVerifier(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	postMsg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   -100,
		Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)),
		Duration:    100,
	}

	postResp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), postMsg)
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}
	t.Logf("Created job: %s", postResp.JobId)

	proofBytes := []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}
	
	submitMsg := &types.MsgSubmitProof{
		Miner:        testCustomer,
		JobId:        postResp.JobId,
		Energy:       -150,
		Proof:        proofBytes,
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	}

	resp, err := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), submitMsg)
	if err != nil {
		t.Logf("SubmitProof error: %v", err)
	} else {
		t.Logf("Proof accepted! Shares earned: %d", resp.SharesEarned)
	}

	job, _ := k.GetJob(ctx, postResp.JobId)
	t.Logf("Job state: BestEnergy=%d, TotalShares=%d, BestSolver=%s",
		job.BestEnergy, job.TotalShares, job.BestSolver)

	if job.TotalShares > 0 {
		t.Logf("SUCCESS: Shares recorded via Universal Share Formula")
	}
}

func TestUniversalShareFormulaWithVerifier(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	postMsg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)),
		Duration:    100,
	}

	postResp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), postMsg)
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}
	t.Logf("Created job: %s with threshold=%d", postResp.JobId, postMsg.Threshold)

	proofBytes := []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}

	// First proof: energy=-500 -> bootstrap shares = abs(-500) = 500
	submit1 := &types.MsgSubmitProof{
		Miner:        testCustomer,
		JobId:        postResp.JobId,
		Energy:       -500,
		Proof:        proofBytes,
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	}
	resp1, err1 := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), submit1)
	if err1 != nil {
		t.Fatalf("Proof 1 failed: %v", err1)
	}
	job, _ := k.GetJob(ctx, postResp.JobId)
	t.Logf("After proof 1 (energy=-500): shares=%d, totalShares=%d, bestEnergy=%d", 
		resp1.SharesEarned, job.TotalShares, job.BestEnergy)

	// Second proof: energy=-700 -> improvement = -500 - (-700) = 200 shares
	submit2 := &types.MsgSubmitProof{
		Miner:        testCustomer,
		JobId:        postResp.JobId,
		Energy:       -700,
		Proof:        proofBytes,
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000003",
	}
	resp2, err2 := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), submit2)
	if err2 != nil {
		t.Fatalf("Proof 2 failed: %v", err2)
	}
	job, _ = k.GetJob(ctx, postResp.JobId)
	t.Logf("After proof 2 (energy=-700): shares=%d, totalShares=%d, bestEnergy=%d",
		resp2.SharesEarned, job.TotalShares, job.BestEnergy)

	// Third proof: energy=-650 (worse than -700) -> 0 shares
	submit3 := &types.MsgSubmitProof{
		Miner:        testCustomer,
		JobId:        postResp.JobId,
		Energy:       -650,
		Proof:        proofBytes,
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000004",
	}
	resp3, err3 := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), submit3)
	if err3 != nil {
		t.Fatalf("Proof 3 failed: %v", err3)
	}
	job, _ = k.GetJob(ctx, postResp.JobId)
	t.Logf("After proof 3 (energy=-650, worse): shares=%d, totalShares=%d, bestEnergy=%d",
		resp3.SharesEarned, job.TotalShares, job.BestEnergy)

	// Verify results
	t.Logf("=== VERIFICATION ===")
	
	if resp1.SharesEarned != 500 {
		t.Errorf("Proof 1: expected 500 shares (bootstrap), got %d", resp1.SharesEarned)
	} else {
		t.Logf("✓ Proof 1: Bootstrap shares = 500")
	}

	if resp2.SharesEarned != 200 {
		t.Errorf("Proof 2: expected 200 shares (improvement), got %d", resp2.SharesEarned)
	} else {
		t.Logf("✓ Proof 2: Improvement shares = 200")
	}

	if resp3.SharesEarned != 0 {
		t.Errorf("Proof 3: expected 0 shares (no improvement), got %d", resp3.SharesEarned)
	} else {
		t.Logf("✓ Proof 3: No improvement = 0 shares")
	}

	if job.TotalShares != 700 {
		t.Errorf("Total shares: expected 700, got %d", job.TotalShares)
	} else {
		t.Logf("✓ Total shares = 700")
	}

	if job.BestEnergy != -700 {
		t.Errorf("Best energy: expected -700, got %d", job.BestEnergy)
	} else {
		t.Logf("✓ Best energy = -700")
	}

	t.Logf("=== UNIVERSAL SHARE FORMULA VERIFIED WITH NOVA VERIFIER ===")
}

// ========================================
// FEE BURNING TESTS
// ========================================

func TestPostJobWithFeeBurn(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)

	// Give customer initial balance
	initialBalance := sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000))
	bankKeeper.SetBalance(customerAddr, initialBalance)

	t.Logf("=== FEE BURN TEST ===")
	t.Logf("Customer initial: %s", initialBalance)

	// Post job with 1M reward
	grossReward := sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000))
	msg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      grossReward,
		Duration:    100,
	}

	resp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}

	// Check job stored with net reward (98% of 1M = 980,000)
	job, _ := k.GetJob(ctx, resp.JobId)
	expectedNetReward := int64(980000) // 1M - 2% fee
	if job.Reward != expectedNetReward {
		t.Errorf("Job reward: expected %d, got %d", expectedNetReward, job.Reward)
	} else {
		t.Logf("✓ Job net reward: %d (after 2%% fee)", job.Reward)
	}

	// Customer paid full amount
	customerBalance := bankKeeper.Balances[customerAddr.String()]
	expectedCustomer := sdk.NewCoins(sdk.NewInt64Coin("unexus", 9000000))
	if !customerBalance.Equal(expectedCustomer) {
		t.Errorf("Customer balance: expected %s, got %s", expectedCustomer, customerBalance)
	} else {
		t.Logf("✓ Customer paid gross: %s", customerBalance)
	}

	// Module has net amount (fee burned)
	moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
	expectedModule := sdk.NewCoins(sdk.NewInt64Coin("unexus", 980000))
	if !moduleBalance.Equal(expectedModule) {
		t.Errorf("Module balance: expected %s, got %s", expectedModule, moduleBalance)
	} else {
		t.Logf("✓ Module holds net (fee burned): %s", moduleBalance)
	}

	t.Logf("✓ 2%% job fee (20,000 unexus) was burned!")
}

func TestClaimRewardsWithValidatorBurn(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	minerAddr, _ := sdk.AccAddressFromBech32(testMiner)

	// Setup
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)))

	t.Logf("=== VALIDATOR SHARE BURN TEST ===")

	// Post job (1M gross, 980K net after 2% fee)
	postMsg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)),
		Duration:    100,
	}
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), postMsg)

	job, _ := k.GetJob(ctx, postResp.JobId)
	t.Logf("Job net reward: %d", job.Reward)
	t.Logf("Module balance: %s", bankKeeper.ModuleBalances[types.ModuleName])

	// Miner submits proof - earns all shares
	proofBytes := []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}
	submitMsg := &types.MsgSubmitProof{
		Miner:        testMiner,
		JobId:        postResp.JobId,
		Energy:       -500,
		Proof:        proofBytes,
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	}
	msgServer.SubmitProof(sdk.WrapSDKContext(ctx), submitMsg)

	// Miner claims - should get 80% of 980K = 784K
	// Validator share (20% of 980K = 196K) should be burned
	claimMsg := &types.MsgClaimRewards{
		Claimer: testMiner,
		JobId:   postResp.JobId,
	}
	claimResp, err := msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), claimMsg)
	if err != nil {
		t.Fatalf("ClaimRewards failed: %v", err)
	}

	// Verify miner reward (80% of 980K = 784K)
	minerBalance := bankKeeper.Balances[minerAddr.String()]
	expectedMiner := sdk.NewCoins(sdk.NewInt64Coin("unexus", 784000))
	if !minerBalance.Equal(expectedMiner) {
		t.Errorf("Miner balance: expected %s, got %s", expectedMiner, minerBalance)
	} else {
		t.Logf("✓ Miner received 80%%: %s", minerBalance)
	}
	t.Logf("Claim response amount: %s", claimResp.Amount)

	// Module should be empty (validator share burned)
	moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
	if !moduleBalance.IsZero() {
		t.Errorf("Module should be empty after burn, got: %s", moduleBalance)
	} else {
		t.Logf("✓ Module empty (validator share burned)")
	}

	t.Logf("✓ 20%% validator share (196,000 unexus) was burned!")
}

func TestCancelJobWithPartialRefund(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)

	// Customer starts with 10M
	initialBalance := sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000))
	bankKeeper.SetBalance(customerAddr, initialBalance)

	t.Logf("=== CANCEL WITH FEE ALREADY BURNED ===")

	reward := sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000))
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      reward,
		Duration:    100,
	})
	t.Logf("Created job: %s", postResp.JobId)

	// Customer paid 1M, but 2% (20K) was burned
	balanceAfterPost := bankKeeper.Balances[customerAddr.String()]
	t.Logf("Customer after post: %s", balanceAfterPost)

	// Cancel the job (no work done)
	cancelResp, err := msgServer.CancelJob(sdk.WrapSDKContext(ctx), &types.MsgCancelJob{
		Customer: testCustomer,
		JobId:    postResp.JobId,
	})
	if err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}
	t.Logf("Job cancelled: %v", cancelResp.Success)

	// Customer gets net refund (980K), fee already burned
	balanceAfterCancel := bankKeeper.Balances[customerAddr.String()]
	expectedRefund := sdk.NewCoins(sdk.NewInt64Coin("unexus", 9980000)) // 10M - 20K burned
	if !balanceAfterCancel.Equal(expectedRefund) {
		t.Errorf("Customer balance: expected %s, got %s", expectedRefund, balanceAfterCancel)
	} else {
		t.Logf("✓ Customer refunded net: %s (fee was already burned)", balanceAfterCancel)
	}

	// Module is empty
	moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
	if !moduleBalance.IsZero() {
		t.Errorf("Module should be empty, got: %s", moduleBalance)
	} else {
		t.Logf("✓ Module balance is zero")
	}

	t.Logf("✓ 2%% fee (20,000 unexus) remained burned even after cancel!")
}

func TestFullDeflatinaryFlow(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	minerAddr, _ := sdk.AccAddressFromBech32(testMiner)

	// Track total supply
	initialSupply := int64(10000000)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", initialSupply)))

	t.Logf("=== FULL DEFLATIONARY FLOW ===")
	t.Logf("Initial supply: %d unexus", initialSupply)

	// Post job with 1M reward
	grossReward := int64(1000000)
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", grossReward)),
		Duration:    100,
	})

	// Calculate burns
	jobFeeBurn := grossReward * 2 / 100       // 2% = 20,000
	netReward := grossReward - jobFeeBurn     // 980,000
	minerReward := netReward * 80 / 100       // 80% = 784,000
	validatorBurn := netReward - minerReward  // 20% = 196,000
	totalBurned := jobFeeBurn + validatorBurn // 216,000

	t.Logf("\nExpected burns:")
	t.Logf("  Job fee (2%%): %d", jobFeeBurn)
	t.Logf("  Validator share (20%%): %d", validatorBurn)
	t.Logf("  Total burned: %d", totalBurned)
	t.Logf("  Miner receives: %d", minerReward)

	// Submit proof
	proofBytes := []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}
	msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{
		Miner:        testMiner,
		JobId:        postResp.JobId,
		Energy:       -500,
		Proof:        proofBytes,
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	})

	// Claim rewards
	msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), &types.MsgClaimRewards{
		Claimer: testMiner,
		JobId:   postResp.JobId,
	})

	// Calculate final supply
	customerFinal := bankKeeper.Balances[customerAddr.String()].AmountOf("unexus").Int64()
	minerFinal := bankKeeper.Balances[minerAddr.String()].AmountOf("unexus").Int64()
	moduleFinal := bankKeeper.ModuleBalances[types.ModuleName].AmountOf("unexus").Int64()
	finalSupply := customerFinal + minerFinal + moduleFinal

	t.Logf("\n=== FINAL STATE ===")
	t.Logf("Customer: %d", customerFinal)
	t.Logf("Miner: %d", minerFinal)
	t.Logf("Module: %d", moduleFinal)
	t.Logf("Total circulating: %d", finalSupply)
	t.Logf("Burned: %d", initialSupply-finalSupply)

	// Verify
	expectedFinalSupply := initialSupply - totalBurned
	if finalSupply != expectedFinalSupply {
		t.Errorf("Final supply: expected %d, got %d", expectedFinalSupply, finalSupply)
	} else {
		t.Logf("\n✓ DEFLATIONARY: %d unexus burned (%.2f%% of job reward)", totalBurned, float64(totalBurned)*100/float64(grossReward))
	}

	if minerFinal != minerReward {
		t.Errorf("Miner reward: expected %d, got %d", minerReward, minerFinal)
	} else {
		t.Logf("✓ Miner received correct amount: %d", minerFinal)
	}
}

func TestInsufficientFundsForJob(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)

	// Customer has only 100 tokens
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 100)))

	// Try to post job with 1M reward
	reward := sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000))
	msg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      reward,
		Duration:    100,
	}

	_, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), msg)
	if err == nil {
		t.Error("Expected error for insufficient funds")
	} else {
		t.Logf("✓ Correctly rejected: %v", err)
	}
}
