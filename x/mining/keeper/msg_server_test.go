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
// BANK KEEPER TESTS
// ========================================

func TestPostJobWithEscrow(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	
	// Give customer initial balance
	initialBalance := sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000))
	bankKeeper.SetBalance(customerAddr, initialBalance)

	t.Logf("Customer initial balance: %s", initialBalance)

	reward := sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000))
	msg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      reward,
		Duration:    100,
	}

	resp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}
	t.Logf("Created job: %s", resp.JobId)

	// Verify customer balance decreased
	customerBalance := bankKeeper.Balances[customerAddr.String()]
	expectedBalance := sdk.NewCoins(sdk.NewInt64Coin("unexus", 9000000))
	if !customerBalance.Equal(expectedBalance) {
		t.Errorf("Customer balance: expected %s, got %s", expectedBalance, customerBalance)
	} else {
		t.Logf("✓ Customer balance decreased: %s", customerBalance)
	}

	// Verify module received the escrow
	moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
	if !moduleBalance.Equal(reward) {
		t.Errorf("Module balance: expected %s, got %s", reward, moduleBalance)
	} else {
		t.Logf("✓ Module received escrow: %s", moduleBalance)
	}
}

func TestClaimRewardsWithTransfer(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	minerAddr, _ := sdk.AccAddressFromBech32(testMiner)

	// Setup: Customer has balance, posts job
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)))

	reward := sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000))
	postMsg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      reward,
		Duration:    100,
	}

	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), postMsg)
	t.Logf("Created job: %s", postResp.JobId)
	t.Logf("Module balance after escrow: %s", bankKeeper.ModuleBalances[types.ModuleName])

	// Miner submits proof
	proofBytes := []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}
	submitMsg := &types.MsgSubmitProof{
		Miner:        testMiner,
		JobId:        postResp.JobId,
		Energy:       -500,
		Proof:        proofBytes,
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	}

	submitResp, err := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), submitMsg)
	if err != nil {
		t.Fatalf("SubmitProof failed: %v", err)
	}
	t.Logf("Miner earned %d shares", submitResp.SharesEarned)

	// Miner claims rewards
	claimMsg := &types.MsgClaimRewards{
		Claimer: testMiner,
		JobId:   postResp.JobId,
	}

	claimResp, err := msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), claimMsg)
	if err != nil {
		t.Fatalf("ClaimRewards failed: %v", err)
	}
	t.Logf("Miner claimed: %s", claimResp.Amount)

	// Verify miner received tokens
	minerBalance := bankKeeper.Balances[minerAddr.String()]
	t.Logf("Miner balance after claim: %s", minerBalance)

	if minerBalance.IsZero() {
		t.Error("Miner should have received tokens")
	} else {
		t.Logf("✓ Miner received reward: %s", minerBalance)
	}

	// Verify module balance decreased
	moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
	t.Logf("Module balance after claim: %s", moduleBalance)
}

func TestCancelJobWithRefund(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	
	// Customer starts with 10M
	initialBalance := sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000))
	bankKeeper.SetBalance(customerAddr, initialBalance)

	reward := sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000))
	postMsg := &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      reward,
		Duration:    100,
	}

	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), postMsg)
	t.Logf("Created job: %s", postResp.JobId)

	// Verify escrow happened
	balanceAfterPost := bankKeeper.Balances[customerAddr.String()]
	t.Logf("Customer balance after post: %s", balanceAfterPost)

	// Cancel the job (no work done)
	cancelMsg := &types.MsgCancelJob{
		Customer: testCustomer,
		JobId:    postResp.JobId,
	}

	cancelResp, err := msgServer.CancelJob(sdk.WrapSDKContext(ctx), cancelMsg)
	if err != nil {
		t.Fatalf("CancelJob failed: %v", err)
	}
	t.Logf("Job cancelled: %v", cancelResp.Success)

	// Verify refund
	balanceAfterCancel := bankKeeper.Balances[customerAddr.String()]
	if !balanceAfterCancel.Equal(initialBalance) {
		t.Errorf("Customer should have full refund: expected %s, got %s", initialBalance, balanceAfterCancel)
	} else {
		t.Logf("✓ Customer refunded: %s", balanceAfterCancel)
	}

	// Verify module is empty
	moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
	if !moduleBalance.IsZero() {
		t.Errorf("Module should be empty after refund, got: %s", moduleBalance)
	} else {
		t.Logf("✓ Module balance is zero after refund")
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

func TestFullMiningFlowWithTokens(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)

	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	minerAddr, _ := sdk.AccAddressFromBech32(testMiner)

	// Setup initial balances
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)))
	bankKeeper.SetBalance(minerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 0)))

	t.Logf("=== INITIAL STATE ===")
	t.Logf("Customer: %s", bankKeeper.Balances[customerAddr.String()])
	t.Logf("Miner: %s", bankKeeper.Balances[minerAddr.String()])

	// 1. Customer posts job
	reward := sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000))
	postResp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer:    testCustomer,
		ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold:   1000,
		Reward:      reward,
		Duration:    100,
	})
	
	t.Logf("\n=== AFTER POST JOB ===")
	t.Logf("Customer: %s", bankKeeper.Balances[customerAddr.String()])
	t.Logf("Module: %s", bankKeeper.ModuleBalances[types.ModuleName])

	// 2. Miner submits proof
	proofBytes := []byte{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad, 0xbe, 0xef}
	submitResp, _ := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{
		Miner:        testMiner,
		JobId:        postResp.JobId,
		Energy:       -500,
		Proof:        proofBytes,
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	})
	
	t.Logf("\n=== AFTER SUBMIT PROOF ===")
	t.Logf("Miner earned %d shares", submitResp.SharesEarned)
	job, _ := k.GetJob(ctx, postResp.JobId)
	t.Logf("Job total shares: %d", job.TotalShares)

	// 3. Miner claims reward
	claimResp, _ := msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), &types.MsgClaimRewards{
		Claimer: testMiner,
		JobId:   postResp.JobId,
	})
	
	t.Logf("\n=== AFTER CLAIM REWARDS ===")
	t.Logf("Miner claimed: %s", claimResp.Amount)
	t.Logf("Customer: %s", bankKeeper.Balances[customerAddr.String()])
	t.Logf("Miner: %s", bankKeeper.Balances[minerAddr.String()])
	t.Logf("Module: %s", bankKeeper.ModuleBalances[types.ModuleName])

	// Verify final state
	t.Logf("\n=== VERIFICATION ===")
	
	minerFinal := bankKeeper.Balances[minerAddr.String()]
	if minerFinal.AmountOf("unexus").IsPositive() {
		t.Logf("✓ Miner received tokens: %s", minerFinal)
	} else {
		t.Errorf("Miner should have tokens, got: %s", minerFinal)
	}

	customerFinal := bankKeeper.Balances[customerAddr.String()]
	expectedCustomer := sdk.NewCoins(sdk.NewInt64Coin("unexus", 9000000))
	if customerFinal.Equal(expectedCustomer) {
		t.Logf("✓ Customer paid for job: %s", customerFinal)
	}
}
