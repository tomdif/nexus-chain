package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	// Process expired jobs at beginning of each block
	return nil
}

func (k Keeper) EndBlocker(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	height := ctx.BlockHeight()
	
	// Create checkpoint every CheckpointInterval blocks
	if height > 0 && height%params.CheckpointInterval == 0 {
		k.createCheckpoint(ctx, height, params)
	}
	
	return nil
}

func (k Keeper) createCheckpoint(ctx sdk.Context, height int64, params types.Params) {
	lastID := k.GetLastCheckpointID(ctx)
	newID := lastID + 1
	
	startHeight := int64(0)
	if lastID > 0 {
		lastCp, found := k.GetCheckpoint(ctx, lastID)
		if found {
			startHeight = lastCp.EndHeight + 1
		}
	}
	
	checkpoint := types.Checkpoint{
		Id:          newID,
		StartHeight: startHeight,
		EndHeight:   height,
		Timestamp:   ctx.BlockTime().Unix(),
	}
	
	k.SetCheckpoint(ctx, checkpoint)
	k.SetLastCheckpointID(ctx, newID)
	
	k.Logger(ctx).Info("Created checkpoint", "id", newID, "height", height)
}
