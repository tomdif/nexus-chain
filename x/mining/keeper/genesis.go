package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, gs types.GenesisState) {
	k.SetParams(ctx, gs.Params)
	
	for _, job := range gs.Jobs {
		k.SetJob(ctx, job)
	}
	
	for _, cp := range gs.Checkpoints {
		k.SetCheckpoint(ctx, cp)
		if cp.Id > k.GetLastCheckpointID(ctx) {
			k.SetLastCheckpointID(ctx, cp.Id)
		}
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params:      k.GetParams(ctx),
		Jobs:        []types.Job{},
		Checkpoints: []types.Checkpoint{},
	}
}
