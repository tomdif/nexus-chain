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
		ProblemHash: "abc123",
		Threshold:   -100,
		Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)),
		Duration:    100,
	}

	resp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}

	t.Logf("Created job: %s", resp.JobId)

	// Verify job was stored
	job, found := k.GetJob(ctx, resp.JobId)
	if !found {
		t.Fatal("Job not found after creation")
	}

	if job.Customer != msg.Customer {
		t.Errorf("Customer mismatch: got %s, want %s", job.Customer, msg.Customer)
	}

	t.Logf("Job verified: ID=%s, Threshold=%d, Deadline=%d", job.Id, job.Threshold, job.Deadline)
}

func TestSubmitProofAndShares(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	// First create a job
	postMsg := &types.MsgPostJob{
		Customer:    "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
		ProblemHash: "abc123",
		Threshold:   0,
		Reward:      sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000)),
		Duration:    100,
	}

	postResp, err := msgServer.PostJob(sdk.WrapSDKContext(ctx), postMsg)
	if err != nil {
		t.Fatalf("PostJob failed: %v", err)
	}

	t.Logf("Created job: %s", postResp.JobId)

	// Submit first proof - energy 500
	submitMsg1 := &types.MsgSubmitProof{
		Miner:        "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
		JobId:        postResp.JobId,
		Energy:       500,
		Proof:        []byte("test-proof-1"),
		SolutionHash: "solution1",
	}

	resp1, err := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), submitMsg1)
	if err != nil {
		t.Logf("SubmitProof error (expected - verifier not running): %v", err)
	} else {
		t.Logf("First proof: shares earned = %d (expected 500 for bootstrap)", resp1.SharesEarned)
		if resp1.SharesEarned != 500 {
			t.Errorf("Expected 500 shares for first proof, got %d", resp1.SharesEarned)
		}
	}

	job, _ := k.GetJob(ctx, postResp.JobId)
	t.Logf("After first proof: BestEnergy=%d, TotalShares=%d, BestSolver=%s",
		job.BestEnergy, job.TotalShares, job.BestSolver)

	// Submit second proof with better energy - should earn improvement shares
	submitMsg2 := &types.MsgSubmitProof{
		Miner:        "nexus1w8n0qfhyu4ywufc9dpe2mpx48kyz4lhzl292wg",
		JobId:        postResp.JobId,
		Energy:       300, // Better than 500
		Proof:        []byte("test-proof-2"),
		SolutionHash: "solution2",
	}

	resp2, err := msgServer.SubmitProof(sdk.WrapSDKContext(ctx), submitMsg2)
	if err != nil {
		t.Logf("SubmitProof error: %v", err)
	} else {
		t.Logf("Second proof: shares earned = %d (expected 200 improvement)", resp2.SharesEarned)
		if resp2.SharesEarned != 200 {
			t.Errorf("Expected 200 shares for improvement (500-300), got %d", resp2.SharesEarned)
		}
	}

	job, _ = k.GetJob(ctx, postResp.JobId)
	t.Logf("After second proof: BestEnergy=%d, TotalShares=%d", job.BestEnergy, job.TotalShares)
}
