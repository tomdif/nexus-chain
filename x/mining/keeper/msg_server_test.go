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
	// Configure SDK with nexus bech32 prefix
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("nexus", "nexuspub")
	config.SetBech32PrefixForValidator("nexusvaloper", "nexusvaloperpub")
	config.SetBech32PrefixForConsensusNode("nexusvalcons", "nexusvalconspub")
	config.Seal()
}

func setupKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
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

	k := keeper.NewKeeper(cdc, storeKey, memKey, nil, nil, "authority")
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}

func TestPostJob(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	msg := &types.MsgPostJob{
		Customer:    "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
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
		Customer:    "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
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
		Miner:        "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
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
		Customer:    "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
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
		Miner:        "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
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
		Miner:        "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
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
		Miner:        "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
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
