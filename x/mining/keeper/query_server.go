package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

type queryServer struct {
	Keeper
}

func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

func (q queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}

func (q queryServer) Job(goCtx context.Context, req *types.QueryJobRequest) (*types.QueryJobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	job, found := q.Keeper.GetJob(ctx, req.JobId)
	if !found {
		return nil, types.ErrJobNotFound
	}
	return &types.QueryJobResponse{Job: job}, nil
}

func (q queryServer) Jobs(goCtx context.Context, req *types.QueryJobsRequest) (*types.QueryJobsResponse, error) {
	return &types.QueryJobsResponse{Jobs: []types.Job{}}, nil
}

func (q queryServer) MinerShares(goCtx context.Context, req *types.QueryMinerSharesRequest) (*types.QueryMinerSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	miner, err := sdk.AccAddressFromBech32(req.Miner)
	if err != nil {
		return nil, err
	}
	shares := q.Keeper.GetShares(ctx, miner, req.JobId)
	return &types.QueryMinerSharesResponse{
		Miner:  req.Miner,
		JobId:  req.JobId,
		Shares: shares,
	}, nil
}

func (q queryServer) ValidatorMiningRecord(goCtx context.Context, req *types.QueryValidatorMiningRecordRequest) (*types.QueryValidatorMiningRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	valAddr, err := sdk.ValAddressFromBech32(req.Validator)
	if err != nil {
		return nil, err
	}
	record, found := q.Keeper.GetValidatorMiningRecord(ctx, valAddr)
	if !found {
		return nil, types.ErrValidatorNotFound
	}
	return &types.QueryValidatorMiningRecordResponse{Record: record}, nil
}

func (q queryServer) Checkpoint(goCtx context.Context, req *types.QueryCheckpointRequest) (*types.QueryCheckpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	checkpoint, found := q.Keeper.GetCheckpoint(ctx, req.Id)
	if !found {
		return nil, types.ErrCheckpointNotFound
	}
	return &types.QueryCheckpointResponse{Checkpoint: checkpoint}, nil
}

func (q queryServer) LatestCheckpoint(goCtx context.Context, req *types.QueryLatestCheckpointRequest) (*types.QueryLatestCheckpointResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	lastID := q.Keeper.GetLastCheckpointID(ctx)
	if lastID == 0 {
		return nil, types.ErrCheckpointNotFound
	}
	checkpoint, found := q.Keeper.GetCheckpoint(ctx, lastID)
	if !found {
		return nil, types.ErrCheckpointNotFound
	}
	return &types.QueryLatestCheckpointResponse{Checkpoint: checkpoint}, nil
}
