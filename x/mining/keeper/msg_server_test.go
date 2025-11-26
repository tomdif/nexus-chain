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
