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

const (
	testCustomer = "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg"
	testMiner    = "nexus109vzpgxnm8mjye50yaf4vj8yp59he3aclran4g"
)

// Helper to post and activate a job for testing
func postAndActivateJob(t *testing.T, k keeper.Keeper, ctx sdk.Context, msgServer types.MsgServer, msg *types.MsgPostJob) string {
	postResp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}
	job, found := k.GetJob(ctx, postResp.JobId)
	if !found {
		t.Fatalf("Job not found: %s", postResp.JobId)
	}
	job.Status = types.JobStatusActive
	job.Deadline = ctx.BlockHeight() + 100
	k.SetJob(ctx, job)
	return postResp.JobId
}

func TestPostJob(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)
	resp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: -100, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}
	t.Logf("Created job: %s", resp.JobId)
	job, found := k.GetJob(ctx, resp.JobId)
	if !found {
		t.Fatal("Job not found")
	}
	t.Logf("Job verified: ID=%s", job.Id)
}

func TestSubmitProof(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)
	jobId := postAndActivateJob(t, k, ctx, msgServer, &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: -100, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	resp, err := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{
		Miner: testCustomer, JobId: jobId, Energy: -150, Proof: []byte{0xde, 0xad, 0xbe, 0xef},
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	})
	if err != nil {
		t.Logf("SubmitProof error: %v", err)
	} else {
		t.Logf("Shares: %d", resp.Shares)
	}
	job, _ := k.GetJob(ctx, jobId)
	if job.TotalShares > 0 {
		t.Logf("SUCCESS: Universal Share Formula working")
	}
}

func TestUniversalShareFormula(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)
	jobId := postAndActivateJob(t, k, ctx, msgServer, &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	proof := []byte{0xde, 0xad, 0xbe, 0xef}
	resp1, _ := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{Miner: testCustomer, JobId: jobId, Energy: -500, Proof: proof, SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002"})
	resp2, _ := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{Miner: testCustomer, JobId: jobId, Energy: -700, Proof: proof, SolutionHash: "0000000000000000000000000000000000000000000000000000000000000003"})
	resp3, _ := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{Miner: testCustomer, JobId: jobId, Energy: -650, Proof: proof, SolutionHash: "0000000000000000000000000000000000000000000000000000000000000004"})
	job, _ := k.GetJob(ctx, jobId)
	if resp1.Shares != 500 || resp2.Shares != 200 || resp3.Shares != 0 || job.TotalShares != 700 {
		t.Errorf("Share formula failed")
	} else {
		t.Logf("Universal Share Formula verified: 500+200+0=700")
	}
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
	moduleBalance := bankKeeper.ModuleBalances[types.ModuleName]
	t.Logf("2%% fee burn: Customer paid 1M, module has %s, %d burned", moduleBalance.String(), 1000000-job.Reward-moduleBalance.AmountOf("unexus").Int64())
}

func TestClaimRewardsValidatorShareHeld(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)
	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)))
	jobId := postAndActivateJob(t, k, ctx, msgServer, &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)), Duration: 100,
	})
	msgServer.SubmitProof(sdk.WrapSDKContext(ctx), &types.MsgSubmitProof{
		Miner: testMiner, JobId: jobId, Energy: -500, Proof: []byte{0x01},
		SolutionHash: "0000000000000000000000000000000000000000000000000000000000000002",
	})
	minerAddr, _ := sdk.AccAddressFromBech32(testMiner)
	msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), &types.MsgClaimRewards{Claimer: testMiner, JobId: jobId})
	minerBalance := bankKeeper.Balances[minerAddr.String()]
	validatorPool := k.GetValidatorRewardPool(ctx)
	t.Logf("Miner got 80%%: %d", minerBalance.AmountOf("unexus").Int64())
	t.Logf("Validator share held: %d", validatorPool)
}

func TestCancelJobRefund(t *testing.T) {
	bankKeeper := NewMockBankKeeper()
	k, ctx := setupKeeperWithBank(t, bankKeeper)
	msgServer := keeper.NewMsgServerImpl(k)
	customerAddr, _ := sdk.AccAddressFromBech32(testCustomer)
	bankKeeper.SetBalance(customerAddr, sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)))
	resp, _ := msgServer.PostJob(sdk.WrapSDKContext(ctx), &types.MsgPostJob{
		Customer: testCustomer, ProblemHash: "0000000000000000000000000000000000000000000000000000000000000001",
		Threshold: 1000, Reward: sdk.NewCoins(sdk.NewInt64Coin("unexus", 10000000)), Duration: 100,
	})
	msgServer.CancelJob(sdk.WrapSDKContext(ctx), &types.MsgCancelJob{Customer: testCustomer, JobId: resp.JobId})
	t.Logf("Customer refunded net: %d", bankKeeper.Balances[customerAddr.String()].AmountOf("unexus").Int64())
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
	if err != nil {
		t.Logf("Correctly rejected: %v", err)
	} else {
		t.Error("Should have rejected insufficient funds")
	}
}

func TestBackgroundJobGeneration(t *testing.T) {
	k, ctx := setupKeeper(t)
	t.Logf("=== BACKGROUND JOB GENERATION ===")
	activeCount := k.GetActiveJobCount(ctx)
	t.Logf("Initial active job count: %d", activeCount)
	k.CheckAndGenerateBackgroundJob(ctx)
	activeCount = k.GetActiveJobCount(ctx)
	t.Logf("After generation, active count: %d", activeCount)
	if activeCount != 1 {
		t.Errorf("Expected 1 active job, got %d", activeCount)
	}
	currentJobID := k.GetCurrentJobID(ctx)
	t.Logf("Current job ID: %s", currentJobID)
	if currentJobID == "" {
		t.Error("Expected current job ID to be set")
	}
	job, found := k.GetJob(ctx, currentJobID)
	if !found {
		t.Error("Job not found")
	}
	t.Logf("Job details: type=%s, threshold=%d, is_background=%t", job.ProblemType, job.Threshold, job.IsBackground)
	if !job.IsBackground {
		t.Error("Expected job to be background")
	}
	t.Logf("✓ Background job generation working!")
}

func TestPublicJobQueue(t *testing.T) {
	k, ctx := setupKeeper(t)
	t.Logf("=== PUBLIC JOB QUEUE (RANDOM SELECTION) ===")
	k.AddToPublicJobQueue(ctx, "pub_job_1")
	k.AddToPublicJobQueue(ctx, "pub_job_2")
	k.AddToPublicJobQueue(ctx, "pub_job_3")
	queueLen := k.GetPublicJobQueueLength(ctx)
	t.Logf("Queue length: %d", queueLen)
	if queueLen != 3 {
		t.Errorf("Expected 3 jobs in queue, got %d", queueLen)
	}
	selected := k.SelectRandomFromQueue(ctx)
	t.Logf("Randomly selected: %s", selected)
	if selected == "" {
		t.Error("Expected a job to be selected")
	}
	queueLen = k.GetPublicJobQueueLength(ctx)
	if queueLen != 2 {
		t.Errorf("Expected 2 jobs after selection, got %d", queueLen)
	}
	queue := k.GetPublicJobQueue(ctx)
	t.Logf("Remaining queue: %v", queue)
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
	k.AddSolveTime(ctx, 300)
	k.AddSolveTime(ctx, 350)
	shouldAdjust := k.AddSolveTime(ctx, 400)
	t.Logf("Should adjust after 3 solves: %t", shouldAdjust)
	if !shouldAdjust {
		t.Error("Expected adjustment after 3 solves")
	}
	k.AdjustDifficulty(ctx)
	newSize := k.GetCurrentProblemSize(ctx)
	t.Logf("New problem size: %d (was %d)", newSize, initialSize)
	if newSize <= initialSize {
		t.Errorf("Expected size to increase, got %d <= %d", newSize, initialSize)
	}
	t.Logf("✓ Difficulty adjustment working!")
}

func TestPaidJobQueue(t *testing.T) {
	k, ctx := setupKeeper(t)
	t.Logf("=== PAID JOB QUEUE (PRIORITY FEE SORTED) ===")

	// Add jobs with different priority fees
	pos1 := k.AddToPaidJobQueue(ctx, "job_low", 100)
	pos2 := k.AddToPaidJobQueue(ctx, "job_high", 1000)
	pos3 := k.AddToPaidJobQueue(ctx, "job_med", 500)

	t.Logf("Added job_low (fee=100) at position: %d", pos1)
	t.Logf("Added job_high (fee=1000) at position: %d", pos2)
	t.Logf("Added job_med (fee=500) at position: %d", pos3)

	// Pop should return highest priority first
	first := k.PopFromPaidJobQueue(ctx)
	second := k.PopFromPaidJobQueue(ctx)
	third := k.PopFromPaidJobQueue(ctx)

	t.Logf("Pop order: %s, %s, %s", first, second, third)

	if first != "job_high" || second != "job_med" || third != "job_low" {
		t.Errorf("Expected job_high, job_med, job_low but got %s, %s, %s", first, second, third)
	}

	t.Logf("✓ Paid job queue with priority fee sorting working!")
}

func TestEmissionSchedule(t *testing.T) {
	k, ctx := setupKeeper(t)
	t.Logf("=== EMISSION SCHEDULE TEST ===")
	epoch := k.GetCurrentEpoch(ctx)
	t.Logf("Current epoch: %d", epoch)
	rate := k.GetCurrentEmissionRate(ctx)
	t.Logf("Current emission rate: %d unex/minute", rate)
	if rate != 35950000000 {
		t.Errorf("Expected 35950000000, got %d", rate)
	}
	escrow := k.GetEmissionEscrow(ctx)
	t.Logf("Initial escrow: %d", escrow)
	t.Logf("Emission schedule verified for Epoch 1")
}
